package ratelimiter

import (
	"net/http"
	"time"
)

// DedupConfig holds options for building a DedupStore via NewDedupWithOptions.
type DedupConfig struct {
	ttl        time.Duration
	keyFunc    func(*http.Request) string
	onDuplicate http.HandlerFunc
}

// DedupOption is a functional option for DedupConfig.
type DedupOption func(*DedupConfig)

// WithDedupTTL sets how long a fingerprint is retained.
func WithDedupTTL(ttl time.Duration) DedupOption {
	return func(c *DedupConfig) {
		c.ttl = ttl
	}
}

// WithDedupKeyFunc overrides the fingerprint function.
func WithDedupKeyFunc(fn func(*http.Request) string) DedupOption {
	return func(c *DedupConfig) {
		c.keyFunc = fn
	}
}

// WithDedupHandler sets a custom HTTP handler invoked when a duplicate is detected.
func WithDedupHandler(h http.HandlerFunc) DedupOption {
	return func(c *DedupConfig) {
		c.onDuplicate = h
	}
}

// NewDedupWithOptions constructs a middleware using functional options.
// It returns a middleware func ready to wrap an http.Handler.
func NewDedupWithOptions(opts ...DedupOption) func(http.Handler) http.Handler {
	cfg := &DedupConfig{
		ttl:     5 * time.Second,
		keyFunc: defaultDedupKeyFunc,
		onDuplicate: defaultDedupRejectedHandler,
	}
	for _, o := range opts {
		o(cfg)
	}
	store := NewDedupStore(cfg.ttl)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fingerprint := cfg.keyFunc(r)
			if store.IsDuplicate(fingerprint) {
				cfg.onDuplicate(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
