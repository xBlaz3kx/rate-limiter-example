package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	rate_limiter "github.com/xBlaz3kx/rate-limiter-example/internal/server/rate-limiter"
	"go.uber.org/zap"
)

func TestHandler_MissingId(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// Create a new rate limiter
	rateLimiter := rate_limiter.NewSlidingWindowRateLimiter()

	// Test the handler
	r := gin.New()
	h := NewHandler(rateLimiter)
	r.GET("", h.HandleRequest)

	// Make a request
	req, _ := http.NewRequest(http.MethodGet, "/?clientId=", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	response := errorResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.EqualValues(t, badRequest, response)

	// Make a request
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	response = errorResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.EqualValues(t, badRequest, response)
}

func TestHandler_ConcurrentRequests(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Errorf("execution timed out")
			}
		}
	}()

	// Create a new rate limiter
	rateLimiter := rate_limiter.NewSlidingWindowRateLimiter()

	// Test the handler
	r := gin.New()
	h := NewHandler(rateLimiter)
	r.GET("", h.HandleRequest)

	// Spawn multiple threads to test the rate limiter
	numRequests := 202
	wg := sync.WaitGroup{}
	wg.Add(numRequests)

	requestNum := atomic.Int32{}

	// Create a http client
	for j := 0; j < numRequests; j++ {
		go func() {
			defer wg.Done()

			requestNum.Add(1)

			// Make a request
			req, _ := http.NewRequest(http.MethodGet, "/?clientId=1", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			currentReq := requestNum.Load()
			logger.Info("Request made", zap.Int32("requestNum", currentReq))
			// If there are more than 200 requests, the request should be limited
			if currentReq <= 200 {
				assert.Equal(t, http.StatusNoContent, w.Code)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
				response := errorResponse{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.EqualValues(t, badRequest, response)
			}
		}()

		// Wait for a short time to simulate a delay between requests
		time.Sleep(time.Millisecond * 10)
	}

	wg.Wait()
}
