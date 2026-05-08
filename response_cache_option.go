package ratelimiter

import "time"

// ResponseCacheConfig holds configuration for building a ResponseCache
// with custom options.
type ResponseCacheConfig struct {
	ttl    time.Duration
	keyFunc func(*http.Request) string
}

import "net/http"

// CacheOption is a functional option for ResponseCacheConfig.
type CacheOption func(*ResponseCacheConfig)

// WithCacheTTL sets the time-to-live for cached entries.
func WithCacheTTL(ttl time.Duration) CacheOption {
	return func(c *ResponseCacheConfig) {
		c.ttl = ttl
	}
}

// WithCacheKeyFunc sets a custom key function for the response cache.
func WithCacheKeyFunc(fn func(*http.Request) string) CacheOption {
	return func(c *ResponseCacheConfig) {
		c.keyFunc = fn
	}
}

// NewResponseCacheWithOptions builds a ResponseCache and returns the
// configured middleware using the provided options.
func NewResponseCacheWithOptions(opts ...CacheOption) (*ResponseCache, func(*http.Request) string) {
	cfg := &ResponseCacheConfig{
		ttl:     30 * time.Second,
		keyFunc: defaultCacheKeyFunc,
	}
	for _, o := range opts {
		o(cfg)
	}
	return NewResponseCache(cfg.ttl), cfg.keyFunc
}
