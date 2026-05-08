package ratelimiter

import (
	"context"
	"net/http"
	"time"
)

// TimeoutExceededHandler is called when a request exceeds the configured timeout.
type TimeoutExceededHandler func(w http.ResponseWriter, r *http.Request)

defaultTimeoutExceededHandler := func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusGatewayTimeout)
	w.Write([]byte("504 Gateway Timeout"))
}

// requestTimeoutConfig holds configuration for the request timeout middleware.
type requestTimeoutConfig struct {
	timeout        time.Duration
	exceededHandler TimeoutExceededHandler
}

// RequestTimeoutOption configures the request timeout middleware.
type RequestTimeoutOption func(*requestTimeoutConfig)

// WithTimeoutHandler sets a custom handler for timed-out requests.
func WithTimeoutHandler(h TimeoutExceededHandler) RequestTimeoutOption {
	return func(c *requestTimeoutConfig) {
		if h != nil {
			c.exceededHandler = h
		}
	}
}

// WithRequestTimeout wraps the given handler with a per-request deadline.
// If the downstream handler does not respond within d, the exceededHandler is
// invoked and the context is cancelled. A zero or negative d disables the
// timeout.
func WithRequestTimeout(d time.Duration, opts ...RequestTimeoutOption) func(http.Handler) http.Handler {
	cfg := &requestTimeoutConfig{
		timeout:        d,
		exceededHandler: defaultTimeoutExceededHandler,
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.timeout <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), cfg.timeout)
			defer cancel()

			done := make(chan struct{})
			var timedOut bool

			go func() {
				defer close(done)
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			select {
			case <-done:
				// handler finished in time
			case <-ctx.Done():
				timedOut = true
			}

			if timedOut {
				cfg.exceededHandler(w, r)
			}
		})
	}
}
