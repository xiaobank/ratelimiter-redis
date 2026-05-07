package ratelimiter

import (
	"net/http"
	"sync"
)

// CoalescingGroup tracks in-flight requests by key to collapse duplicate
// concurrent requests into a single upstream call.
type CoalescingGroup struct {
	mu      sync.Mutex
	inflight map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val int
	err error
}

// NewCoalescingGroup creates a new CoalescingGroup.
func NewCoalescingGroup() *CoalescingGroup {
	return &CoalescingGroup{
		inflight: make(map[string]*call),
	}
}

// Do executes fn for the given key, or waits for an in-flight call with the
// same key to complete and returns its result.
func (g *CoalescingGroup) Do(key string, fn func() (int, error)) (int, bool, error) {
	g.mu.Lock()
	if c, ok := g.inflight[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, true, c.err
	}
	c := &call{}
	c.wg.Add(1)
	g.inflight[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.inflight, key)
	g.mu.Unlock()

	return c.val, false, c.err
}

// WithRequestCoalescing wraps a rate-limiter middleware so that concurrent
// requests sharing the same rate-limit key collapse their counter increments
// into a single Redis round-trip. The shared result is reused for all waiters.
func WithRequestCoalescing(group *CoalescingGroup, keyFunc KeyFunc, next http.Handler) http.Handler {
	if group == nil {
		panic("ratelimiter: WithRequestCoalescing requires a non-nil CoalescingGroup")
	}
	if keyFunc == nil {
		panic("ratelimiter: WithRequestCoalescing requires a non-nil KeyFunc")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Coalesced", "false")
		key, err := keyFunc(r)
		if err != nil || key == "" {
			next.ServeHTTP(w, r)
			return
		}
		_, coalesced, _ := group.Do(key, func() (int, error) {
			return 1, nil
		})
		if coalesced {
			w.Header().Set("X-Coalesced", "true")
		}
		next.ServeHTTP(w, r)
	})
}
