package ratelimiter

import (
	"bytes"
	"io"
	"net/http"
)

// BodyLogFunc is called with the request body bytes for each request.
type BodyLogFunc func(r *http.Request, body []byte)

// bodyLoggerConfig holds configuration for the body logger middleware.
type bodyLoggerConfig struct {
	maxBytes int64
	logFunc  BodyLogFunc
	header   string
}

// defaultBodyLogFunc is a no-op default.
var defaultBodyLogFunc BodyLogFunc = func(r *http.Request, body []byte) {}

// WithBodyLogMaxBytes sets the maximum number of body bytes to capture.
func WithBodyLogMaxBytes(n int64) func(*bodyLoggerConfig) {
	return func(c *bodyLoggerConfig) {
		c.maxBytes = n
	}
}

// WithBodyLogFunc sets the function called with captured body bytes.
func WithBodyLogFunc(fn BodyLogFunc) func(*bodyLoggerConfig) {
	return func(c *bodyLoggerConfig) {
		if fn != nil {
			c.logFunc = fn
		}
	}
}

// WithBodyLogHeader sets a response header name to echo the captured byte count.
func WithBodyLogHeader(header string) func(*bodyLoggerConfig) {
	return func(c *bodyLoggerConfig) {
		c.header = header
	}
}

// WithRequestBodyLogger is middleware that reads up to maxBytes of the request
// body, calls logFunc with the captured bytes, and restores the body so
// downstream handlers can still read it.
func WithRequestBodyLogger(opts ...func(*bodyLoggerConfig)) func(http.Handler) http.Handler {
	cfg := &bodyLoggerConfig{
		maxBytes: 4096,
		logFunc:  defaultBodyLogFunc,
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			limited := io.LimitReader(r.Body, cfg.maxBytes)
			captured, err := io.ReadAll(limited)
			if err != nil {
				captured = nil
			}

			// Restore the body for downstream handlers.
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(captured), r.Body))

			cfg.logFunc(r, captured)

			if cfg.header != "" {
				w.Header().Set(cfg.header, itoa(int64(len(captured))))
			}

			next.ServeHTTP(w, r)
		})
	}
}
