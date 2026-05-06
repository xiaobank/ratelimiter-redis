package ratelimiter

import (
	"net/http"
	"time"
)

// WithRetryAfter wraps a rate-limiter middleware handler so that any 429
// response automatically includes a Retry-After header. The window parameter
// controls the fallback value when no custom calculator is provided.
//
// Usage:
//
//	rl := New(store, opts...)
//	http.Handle("/api/", WithRetryAfter(rl, window, nil))
func WithRetryAfter(next http.Handler, window time.Duration, calc RetryAfterCalculator) http.Handler {
	if calc == nil {
		calc = FixedRetryAfter(window)
	}
	return RetryAfterMiddleware(next, calc, window)
}

// Option type alias kept for clarity — callers can compose WithRetryAfter
// with any existing middleware chain.
var _ http.Handler = (http.HandlerFunc)(nil) // compile-time interface check
