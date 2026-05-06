package ratelimiter

import (
	"net/http"
)

// circuitBreakerResponseWriter wraps http.ResponseWriter to capture the
// status code written by the downstream handler.
type circuitBreakerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *circuitBreakerResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// WithCircuitBreakerMiddleware integrates a CircuitBreaker with the rate
// limiter pipeline. It observes 429 responses produced by the rate limiter
// and records them as blocked requests, and records all other responses as
// allowed, automatically managing circuit state.
//
// Usage:
//
//	rl := New(store, opts...)
//	cb := NewCircuitBreaker(WithTripThreshold(10))
//	http.Handle("/", WithCircuitBreakerMiddleware(rl, cb)(yourHandler))
func WithCircuitBreakerMiddleware(next http.Handler, cb *CircuitBreaker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Short-circuit immediately when the breaker is open.
		if cb.State() == CircuitOpen {
			w.Header().Set("Retry-After", "30")
			http.Error(w, "service temporarily unavailable", http.StatusServiceUnavailable)
			return
		}

		rw := &circuitBreakerResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		if rw.statusCode == http.StatusTooManyRequests {
			cb.RecordBlocked()
		} else {
			cb.RecordAllowed()
		}
	})
}
