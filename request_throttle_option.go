package ratelimiter

import (
	"net/http"
	"time"
)

// WithThrottleExceededHandler sets a custom handler invoked when the queue is
// full or the wait timeout is exceeded.
func WithThrottleExceededHandler(h http.Handler) func(*ThrottleConfig) {
	return func(cfg *ThrottleConfig) {
		if h != nil {
			cfg.exceededHandler = h
		}
	}
}

// NewThrottleMiddleware is a convenience constructor that wires up
// WithRequestThrottle with common defaults and returns the middleware directly.
//
//	handler := NewThrottleMiddleware(10, 50, 2*time.Second)
func NewThrottleMiddleware(maxConcurrent, queueSize int, queueTimeout time.Duration, opts ...func(*ThrottleConfig)) func(http.Handler) http.Handler {
	return WithRequestThrottle(maxConcurrent, queueSize, queueTimeout, opts...)
}
