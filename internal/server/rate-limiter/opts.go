package rate_limiter

import (
	"time"
)

type Options func(*SlidingWindowRateLimiter)

func WithLimit(limit int) Options {
	return func(l *SlidingWindowRateLimiter) {
		// Limit must be greater than 0
		if limit < 0 {
			return
		}

		l.config.Limit = limit
	}
}

func WithDuration(duration time.Duration) Options {
	return func(l *SlidingWindowRateLimiter) {
		// Don't apply durations less than 100ms
		if duration < 100*time.Millisecond {
			return
		}

		l.config.Duration = duration
	}
}
