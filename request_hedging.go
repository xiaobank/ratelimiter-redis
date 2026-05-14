package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

// HedgeOption configures the hedging middleware.
type HedgeOption func(*hedgeConfig)

type hedgeConfig struct {
	delay       time.Duration
	maxHedges   int
	shouldHedge func(*http.Request) bool
}

// WithHedgeDelay sets the delay before a hedged request is fired.
func WithHedgeDelay(d time.Duration) HedgeOption {
	return func(c *hedgeConfig) {
		c.delay = d
	}
}

// WithMaxHedges sets the maximum number of concurrent hedged requests (default 1).
func WithMaxHedges(n int) HedgeOption {
	return func(c *hedgeConfig) {
		if n > 0 {
			c.maxHedges = n
		}
	}
}

// WithHedgePredicate sets a function that decides whether a request should be hedged.
func WithHedgePredicate(fn func(*http.Request) bool) HedgeOption {
	return func(c *hedgeConfig) {
		c.shouldHedge = fn
	}
}

type hedgeResponseWriter struct {
	http.ResponseWriter
	mu      sync.Mutex
	written bool
	code    int
	header  http.Header
}

func (h *hedgeResponseWriter) claim() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.written {
		return false
	}
	h.written = true
	return true
}

// WithRequestHedging fires a duplicate request after a delay and uses whichever
// response arrives first, reducing tail latency.
func WithRequestHedging(next http.Handler, opts ...HedgeOption) http.Handler {
	cfg := &hedgeConfig{
		delay:     20 * time.Millisecond,
		maxHedges: 1,
		shouldHedge: func(_ *http.Request) bool { return true },
	}
	for _, o := range opts {
		o(cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.shouldHedge(r) {
			next.ServeHTTP(w, r)
			return
		}

		shared := &hedgeResponseWriter{ResponseWriter: w}
		doneCh := make(chan struct{}, cfg.maxHedges+1)

		serve := func(req *http.Request) {
			rw := &singleUseResponseWriter{shared: shared}
			next.ServeHTTP(rw, req)
			doneCh <- struct{}{}
		}

		go serve(r)

		timer := time.NewTimer(cfg.delay)
		defer timer.Stop()

		for i := 0; i < cfg.maxHedges; i++ {
			select {
			case <-doneCh:
				return
			case <-timer.C:
				go serve(r.Clone(r.Context()))
			}
		}
		<-doneCh
	})
}

// singleUseResponseWriter forwards to the shared writer only if it wins the race.
type singleUseResponseWriter struct {
	shared *hedgeResponseWriter
	code   int
	buf    []byte
}

func (s *singleUseResponseWriter) Header() http.Header {
	return s.shared.ResponseWriter.Header()
}

func (s *singleUseResponseWriter) WriteHeader(code int) {
	s.code = code
}

func (s *singleUseResponseWriter) Write(b []byte) (int, error) {
	if s.shared.claim() {
		if s.code != 0 {
			s.shared.ResponseWriter.WriteHeader(s.code)
		}
		return s.shared.ResponseWriter.Write(b)
	}
	return len(b), nil
}
