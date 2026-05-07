package ratelimiter

import (
	"net/http"
	"sync"
)

// Blacklist holds a set of keys (e.g. IP addresses) that are denied access.
type Blacklist struct {
	mu     sync.RWMutex
	keys   map[string]struct{}
	keyFunc func(*http.Request) string
}

// NewBlacklist creates a new Blacklist with the given initial keys.
// If keyFunc is nil, the default key function (remote IP) is used.
func NewBlacklist(keyFunc func(*http.Request) string, keys ...string) *Blacklist {
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}
	km := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		km[k] = struct{}{}
	}
	return &Blacklist{keys: km, keyFunc: keyFunc}
}

// Contains reports whether the key derived from r is blacklisted.
func (b *Blacklist) Contains(r *http.Request) bool {
	key := b.keyFunc(r)
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.keys[key]
	return ok
}

// Add adds a key to the blacklist.
func (b *Blacklist) Add(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.keys[key] = struct{}{}
}

// Remove removes a key from the blacklist.
func (b *Blacklist) Remove(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.keys, key)
}

// WithBlacklist returns middleware that rejects requests whose key is in the blacklist.
// Rejected requests receive a 403 Forbidden response.
func WithBlacklist(bl *Blacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bl.Contains(r) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
