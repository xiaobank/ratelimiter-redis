package ratelimiter

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	DefaultRequestIDHeader = "X-Request-ID"
	DefaultRequestIDLength = 16
)

// RequestIDConfig holds configuration for the request ID middleware.
type RequestIDConfig struct {
	header    string
	generator func() string
	overwrite bool
}

// requestIDOption is a functional option for RequestIDConfig.
type requestIDOption func(*RequestIDConfig)

// WithRequestIDHeader sets the header name used to store/read the request ID.
func WithRequestIDHeader(header string) requestIDOption {
	return func(c *RequestIDConfig) {
		c.header = header
	}
}

// WithRequestIDGenerator sets a custom ID generator function.
func WithRequestIDGenerator(gen func() string) requestIDOption {
	return func(c *RequestIDConfig) {
		c.generator = gen
	}
}

// WithRequestIDOverwrite forces generation of a new ID even if one exists.
func WithRequestIDOverwrite(overwrite bool) requestIDOption {
	return func(c *RequestIDConfig) {
		c.overwrite = overwrite
	}
}

// defaultRequestIDGenerator generates a random hex string of DefaultRequestIDLength bytes.
func defaultRequestIDGenerator() string {
	b := make([]byte, DefaultRequestIDLength)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WithRequestID is middleware that ensures every request has a unique ID.
// It reads an existing ID from the configured header, or generates one if absent.
// The ID is written back to the response header.
func WithRequestID(opts ...requestIDOption) func(http.Handler) http.Handler {
	cfg := &RequestIDConfig{
		header:    DefaultRequestIDHeader,
		generator: defaultRequestIDGenerator,
		overwrite: false,
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(cfg.header)
			if id == "" || cfg.overwrite {
				id = cfg.generator()
				r.Header.Set(cfg.header, id)
			}
			w.Header().Set(cfg.header, id)
			next.ServeHTTP(w, r)
		})
	}
}
