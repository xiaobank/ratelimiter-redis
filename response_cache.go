package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// CacheEntry holds a cached response.
type CacheEntry struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	ExpiresAt  time.Time
}

// ResponseCache is an in-memory cache for HTTP responses.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewResponseCache creates a ResponseCache with the given TTL.
func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get returns a cached entry and whether it was found and is still valid.
func (c *ResponseCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry, true
}

// Set stores a response under the given key.
func (c *ResponseCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry.ExpiresAt = time.Now().Add(c.ttl)
	c.entries[key] = entry
}

// Delete removes an entry from the cache.
func (c *ResponseCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// defaultCacheKeyFunc uses the request method + URL as the cache key.
func defaultCacheKeyFunc(r *http.Request) string {
	return r.Method + ":" + r.URL.String()
}

// WithResponseCache wraps a handler and caches responses for GET requests.
// Cached responses are served directly and marked with X-Cache: HIT.
func WithResponseCache(cache *ResponseCache, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	if cache == nil {
		panic("ratelimiter: WithResponseCache requires a non-nil cache")
	}
	if keyFunc == nil {
		keyFunc = defaultCacheKeyFunc
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}
			key := keyFunc(r)
			if entry, ok := cache.Get(key); ok {
				for k, vals := range entry.Header {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(entry.StatusCode)
				w.Write(entry.Body) //nolint:errcheck
				return
			}
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			result := rec.Result()
			if result.StatusCode == http.StatusOK {
				cache.Set(key, &CacheEntry{
					StatusCode: result.StatusCode,
					Header:     result.Header.Clone(),
					Body:       rec.Body.Bytes(),
				})
			}
			w.Header().Set("X-Cache", "MISS")
			for k, vals := range rec.Header() {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(result.StatusCode)
			w.Write(rec.Body.Bytes()) //nolint:errcheck
		})
	}
}
