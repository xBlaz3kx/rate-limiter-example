package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_httpClient_SendRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	clientName := "test"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, clientName, r.Form.Get("clientId"))
	}))
	defer svr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := newHttpClient(clientName, svr.URL).SendRequest(ctx)
	assert.NoError(t, err)
}

func Test_httpClient_SendRequest_RateLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	clientName := "test"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)

		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, clientName, r.Form.Get("clientId"))
	}))
	defer svr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := newHttpClient(clientName, svr.URL).SendRequest(ctx)
	assert.EqualError(t, err, "request limit exceeded")
}

func Test_httpClient_SendRequest_Failure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	clientName := "test"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)

		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, clientName, r.Form.Get("clientId"))
	}))
	defer svr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := newHttpClient(clientName, svr.URL).SendRequest(ctx)
	assert.EqualError(t, err, "unexpected status code: 409")
}
