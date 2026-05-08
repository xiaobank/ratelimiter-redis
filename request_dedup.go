package ratelimiter

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DedupStore tracks recently seen request fingerprints to detect duplicates.
type DedupStore struct {
	mu      sync.Mutex
	entries map[string]time.Time
	ttl     time.Duration
}

// NewDedupStore creates a new DedupStore with the given TTL for fingerprint retention.
func NewDedupStore(ttl time.Duration) *DedupStore {
	d := &DedupStore{
		entries: make(map[string]time.Time),
		ttl:     ttl,
	}
	go d.evict()
	return d
}

// IsDuplicate returns true if the fingerprint was seen within the TTL window.
// It records the fingerprint if it is new.
func (d *DedupStore) IsDuplicate(fingerprint string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if exp, ok := d.entries[fingerprint]; ok && time.Now().Before(exp) {
		return true
	}
	d.entries[fingerprint] = time.Now().Add(d.ttl)
	return false
}

func (d *DedupStore) evict() {
	ticker := time.NewTicker(d.ttl)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		d.mu.Lock()
		for k, exp := range d.entries {
			if now.After(exp) {
				delete(d.entries, k)
			}
		}
		d.mu.Unlock()
	}
}

// defaultDedupKeyFunc builds a fingerprint from method, URL path, and client IP.
func defaultDedupKeyFunc(r *http.Request) string {
	raw := fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, r.RemoteAddr)
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

func defaultDedupRejectedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "duplicate request", http.StatusConflict)
}

// WithRequestDedup returns middleware that rejects duplicate requests within the TTL window.
// An optional keyFunc may be supplied; if nil the default (method+path+IP) is used.
func WithRequestDedup(store *DedupStore, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	if store == nil {
		panic("ratelimiter: WithRequestDedup requires a non-nil DedupStore")
	}
	if keyFunc == nil {
		keyFunc = defaultDedupKeyFunc
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fingerprint := keyFunc(r)
			if store.IsDuplicate(fingerprint) {
				defaultDedupRejectedHandler(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
