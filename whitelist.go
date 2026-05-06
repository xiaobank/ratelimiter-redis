package ratelimiter

import (
	"net/http"
)

// WhitelistKeyFunc is a function that extracts a key from a request
// to check against the whitelist.
type WhitelistKeyFunc func(r *http.Request) string

// Whitelist holds a set of keys that are exempt from rate limiting.
type Whitelist struct {
	keys    map[string]struct{}
	keyFunc WhitelistKeyFunc
}

// NewWhitelist creates a Whitelist middleware option using the provided
// key function and list of exempt keys. If keyFunc is nil, the request's
// RemoteAddr is used.
func NewWhitelist(keyFunc WhitelistKeyFunc, keys ...string) *Whitelist {
	if keyFunc == nil {
		keyFunc = func(r *http.Request) string {
			return r.RemoteAddr
		}
	}
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return &Whitelist{
		keys:    m,
		keyFunc: keyFunc,
	}
}

// Contains reports whether the request's extracted key is in the whitelist.
func (w *Whitelist) Contains(r *http.Request) bool {
	key := w.keyFunc(r)
	_, ok := w.keys[key]
	return ok
}

// Add inserts a key into the whitelist.
func (w *Whitelist) Add(key string) {
	w.keys[key] = struct{}{}
}

// Remove deletes a key from the whitelist.
func (w *Whitelist) Remove(key string) {
	delete(w.keys, key)
}

// WithWhitelist wraps next, skipping rate-limit logic for whitelisted requests.
func WithWhitelist(wl *Whitelist, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wl != nil && wl.Contains(r) {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
