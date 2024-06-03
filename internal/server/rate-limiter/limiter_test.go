package rate_limiter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSlidingWindowRateLimiter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	rateLimiter := NewSlidingWindowRateLimiter()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("execution timed out")
		}
	}()

	// Spawn multiple threads to test the rate limiter
	numRoutines := 100
	wg := sync.WaitGroup{}
	wg.Add(numRoutines)
	requestNum := atomic.Int32{}

	for i := 0; i < numRoutines; i++ {
		go func() {
			defer wg.Done()

			// Make 3 requests per routine
			for j := 0; j < 3; j++ {
				requestNum.Add(1)

				limited := rateLimiter.IsLimited("1")
				if limited && requestNum.Load() < 200 {
					t.Errorf("Expected false, got %v", limited)
				}

				logger.Info("Request made", zap.Bool("limited", limited))
			}
		}()
	}

	wg.Wait()

	limited := rateLimiter.IsLimited("1")
	if !limited {
		t.Errorf("Expected true, got %v", limited)
	}

	limited = rateLimiter.IsLimited("2")
	if limited {
		t.Errorf("Expected false, got %v", limited)
	}
}

func TestNewSlidingWindowRateLimiterOpts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// Apply valid options
	limitOpts := WithLimit(100)
	durationOpt := WithDuration(time.Second * 10)
	rateLimiter := NewSlidingWindowRateLimiter(limitOpts, durationOpt)
	assert.Equal(t, 100, rateLimiter.config.Limit)
	assert.Equal(t, time.Second*10, rateLimiter.config.Duration)

	// Apply invalid options
	limitOpts = WithLimit(-100)
	rateLimiter = NewSlidingWindowRateLimiter(limitOpts)
	assert.Equal(t, 200, rateLimiter.config.Limit)
	assert.Equal(t, time.Second*5, rateLimiter.config.Duration)

	durationOpt = WithDuration(time.Millisecond)
	rateLimiter = NewSlidingWindowRateLimiter(durationOpt)
	assert.Equal(t, 200, rateLimiter.config.Limit)
	assert.Equal(t, time.Second*5, rateLimiter.config.Duration)
}
