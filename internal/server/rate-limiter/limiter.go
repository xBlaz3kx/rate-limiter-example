package rate_limiter

import (
	"sync"
	"time"
)

// LimiterConfig is the configuration for the rate limiter
type LimiterConfig struct {

	// Limit is the maximum number of requests allowed within the duration of the window
	Limit int

	// Duration is the duration of the window
	Duration time.Duration
}

type clientLimit struct {
	// Number of requests made by the client in the current window
	requestCount int

	// Time when the current window started
	windowStart time.Time
}

// SlidingWindowRateLimiter is a sliding window rate limiter that limits the number of requests per user for a given time window.
type SlidingWindowRateLimiter struct {
	config LimiterConfig

	// userLimits is a map of user IDs to their current request count and the time the window started
	userLimits map[string]clientLimit // clientLimit is the

	mu sync.Mutex
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter with the provided configuration
func NewSlidingWindowRateLimiter(opts ...Options) *SlidingWindowRateLimiter {
	limiter := &SlidingWindowRateLimiter{
		config: LimiterConfig{
			Limit:    200,
			Duration: time.Second * 5,
		},
		userLimits: make(map[string]clientLimit),
	}

	// Apply options
	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

// NewSlidingWindowRateLimiterFromConfig creates a new sliding window rate limiter with the provided configuration
func NewSlidingWindowRateLimiterFromConfig(config LimiterConfig) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		config:     config,
		userLimits: make(map[string]clientLimit),
	}
}

func (l *SlidingWindowRateLimiter) incrementRequestCount(userID string) {
	// Get the current user limits
	defaultLimit := clientLimit{
		requestCount: 0,
		windowStart:  time.Now(),
	}

	userLimits, loaded := l.userLimits[userID]
	if loaded {
		l.userLimits[userID] = defaultLimit
		userLimits = defaultLimit
	}

	// Increment the request count and update the
	userLimits.requestCount++
	l.userLimits[userID] = userLimits
}

func (l *SlidingWindowRateLimiter) resetRequestCount(userID string) {
	// Reset the request count
	defaultLimit := clientLimit{
		requestCount: 0,
		windowStart:  time.Now(),
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.userLimits[userID] = defaultLimit
}

func (l *SlidingWindowRateLimiter) IsLimited(userID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	defer l.incrementRequestCount(userID)

	// Get the current user limits
	userLimits, loaded := l.userLimits[userID]
	if !loaded {
		return false
	}

	// Check if the user has exceeded the limit
	if userLimits.requestCount > l.config.Limit {
		return true
	}

	// Check if the window has expired
	if time.Since(userLimits.windowStart) > l.config.Duration {
		l.resetRequestCount(userID)
	}

	l.incrementRequestCount(userID)
	return false
}
