package ratelimiter

import (
	"net/http"
	"strconv"
	"time"
)

// RetryAfterCalculator computes the number of seconds a client should wait
// before retrying after being rate limited.
type RetryAfterCalculator interface {
	RetryAfter(key string, window time.Duration) (int, error)
}

// RetryAfterMiddleware wraps an existing handler and injects a Retry-After
// header into 429 responses produced by the rate limiter.
func RetryAfterMiddleware(next http.Handler, calc RetryAfterCalculator, window time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &retryAfterResponseWriter{ResponseWriter: w, window: window, calc: calc, key: r.RemoteAddr}
		next.ServeHTTP(rw, r)
	})
}

type retryAfterResponseWriter struct {
	http.ResponseWriter
	window  time.Duration
	calc    RetryAfterCalculator
	key     string
	wrote   bool
}

func (rw *retryAfterResponseWriter) WriteHeader(status int) {
	if status == http.StatusTooManyRequests && !rw.wrote {
		rw.wrote = true
		seconds := int(rw.window.Seconds())
		if rw.calc != nil {
			if s, err := rw.calc.RetryAfter(rw.key, rw.window); err == nil {
				seconds = s
			}
		}
		rw.ResponseWriter.Header().Set("Retry-After", strconv.Itoa(seconds))
	}
	rw.ResponseWriter.WriteHeader(status)
}

// FixedRetryAfter returns a RetryAfterCalculator that always returns the full
// window duration in seconds, suitable for fixed-window rate limiters.
func FixedRetryAfter(window time.Duration) RetryAfterCalculator {
	return &fixedRetryAfter{window: window}
}

type fixedRetryAfter struct {
	window time.Duration
}

func (f *fixedRetryAfter) RetryAfter(_ string, _ time.Duration) (int, error) {
	return int(f.window.Seconds()), nil
}
