package ratelimiter

import (
	"net/http"
)

// RequestSizeLimitConfig holds configuration for the request size limiting middleware.
type RequestSizeLimitConfig struct {
	// MaxBytes is the maximum allowed request body size in bytes.
	MaxBytes int64
	// OnLimitExceeded is called when a request body exceeds MaxBytes.
	OnLimitExceeded http.HandlerFunc
}

// defaultSizeLimitExceededHandler responds with 413 Request Entity Too Large.
func defaultSizeLimitExceededHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusRequestEntityTooLarge)
	_, _ = w.Write([]byte("request body too large"))
}

// WithRequestSizeLimit returns middleware that rejects requests whose body
// exceeds maxBytes. A maxBytes value of 0 disables the limit.
func WithRequestSizeLimit(maxBytes int64, opts ...RequestSizeLimitOption) func(http.Handler) http.Handler {
	cfg := &RequestSizeLimitConfig{
		MaxBytes:        maxBytes,
		OnLimitExceeded: defaultSizeLimitExceededHandler,
	}
	for _, o := range opts {
		o(cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.MaxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			contentLength := r.ContentLength
			if contentLength > cfg.MaxBytes {
				cfg.OnLimitExceeded(w, r)
				return
			}
			// Wrap body to enforce limit even when Content-Length is absent or wrong.
			r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeLimitOption is a functional option for RequestSizeLimitConfig.
type RequestSizeLimitOption func(*RequestSizeLimitConfig)

// WithSizeLimitHandler overrides the default 413 handler.
func WithSizeLimitHandler(h http.HandlerFunc) RequestSizeLimitOption {
	return func(cfg *RequestSizeLimitConfig) {
		if h != nil {
			cfg.OnLimitExceeded = h
		}
	}
}
