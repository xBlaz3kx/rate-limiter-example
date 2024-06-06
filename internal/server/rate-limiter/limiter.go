package rate_limiter

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// Configuration for the rate limiter
type Config struct {

	// Limit is the maximum number of requests allowed within the duration of the window
	Limit int

	// Duration is the duration of the window
	Duration time.Duration
}

type clientLimit struct {
	// Number of requests made by the client in the current window
	requestCount int

	// Time when the current window started
	windowStart *time.Time
}

// SlidingWindowRateLimiter is a sliding window rate limiter that limits the number of requests per user for a given time window.
// Note: Not exactly a sliding window implementation, but a simpler version that resets the request count after the window duration.
type SlidingWindowRateLimiter struct {
	config Config

	// userLimits is a map of user IDs to their current request count and the time the window started
	userLimits map[string]clientLimit

	mu     sync.RWMutex
	logger *zap.Logger
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter with the provided configuration
func NewSlidingWindowRateLimiter(opts ...Options) *SlidingWindowRateLimiter {
	limiter := &SlidingWindowRateLimiter{
		config: Config{
			Limit:    200,
			Duration: time.Second * 5,
		},
		userLimits: make(map[string]clientLimit),
		logger:     zap.L().Named("rate-limiter"),
	}

	// Apply options
	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

// NewSlidingWindowRateLimiterFromConfig creates a new sliding window rate limiter with the provided configuration
func NewSlidingWindowRateLimiterFromConfig(config Config) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		config:     config,
		userLimits: make(map[string]clientLimit),
		logger:     zap.L().Named("rate-limiter"),
	}
}

func (l *SlidingWindowRateLimiter) incrementRequestCount(userID string) {
	l.logger.Debug("Incrementing the number request")

	// Get the current user rate limit
	userLimits, exists := l.userLimits[userID]
	if !exists {
		// If the request limit does not exist, create a new entry
		currentTime := time.Now()
		l.userLimits[userID] = clientLimit{
			requestCount: 0,
			windowStart:  &currentTime,
		}

		userLimits = l.userLimits[userID]
	}

	// Increment the request count and update
	userLimits.requestCount++
	l.userLimits[userID] = userLimits
}

func (l *SlidingWindowRateLimiter) IsLimited(userID string) bool {
	l.logger.Debug("Checking if user exceeds rate limit")

	l.mu.Lock()
	defer func() {
		l.incrementRequestCount(userID)
		l.mu.Unlock()
	}()

	// Get the current user limits
	userLimits, exists := l.userLimits[userID]
	if !exists {
		return false
	}

	// Check if the user has exceeded the limit
	if userLimits.requestCount >= l.config.Limit {

		// Check if the window has expired
		if userLimits.windowStart != nil && time.Since(*userLimits.windowStart) > l.config.Duration {
			l.logger.Debug("Window expired, resetting request count")
			// Reset the request count
			currentTime := time.Now()
			l.userLimits[userID] = clientLimit{
				requestCount: 0,
				windowStart:  &currentTime,
			}

			return false
		}

		return true
	}

	return false
}
