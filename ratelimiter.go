package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// KeyFunc derives a rate-limit key from an incoming HTTP request.
type KeyFunc func(r *http.Request) string

// Counter is the interface implemented by every rate-limiting strategy.
type Counter interface {
	// Allow reports whether the request identified by key is permitted.
	// It also returns the remaining quota and the window/reset duration.
	Allow(ctx context.Context, key string) (allowed bool, remaining int, reset time.Duration, err error)
}

// Options configures the middleware.
type Options struct {
	// Limiter is the rate-limiting strategy to use (required).
	Limiter Counter

	// KeyFunc determines the rate-limit key for each request.
	// Defaults to the client's remote IP address.
	KeyFunc KeyFunc

	// LimitExceededHandler is called when a request is rejected.
	// Defaults to a plain 429 response.
	LimitExceededHandler http.HandlerFunc
}

// New returns an HTTP middleware that enforces the configured rate limit.
func New(opts Options) func(http.Handler) http.Handler {
	if opts.KeyFunc == nil {
		opts.KeyFunc = defaultKeyFunc
	}
	if opts.LimitExceededHandler == nil {
		opts.LimitExceededHandler = defaultLimitExceededHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := opts.KeyFunc(r)
			allowed, remaining, reset, err := opts.Limiter.Allow(r.Context(), key)
			if err != nil {
				// On backend error, fail open to avoid blocking legitimate traffic.
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%.0f", reset.Seconds()))

			if !allowed {
				opts.LimitExceededHandler(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func defaultKeyFunc(r *http.Request) string {
	return IPKeyFunc(false)(r)
}

func defaultLimitExceededHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
}
