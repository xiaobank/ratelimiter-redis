package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

// ShadowConfig holds configuration for shadow mode rate limiting.
type ShadowConfig struct {
	// OnBlock is called when a request would have been blocked in shadow mode.
	// It receives the request and the key that triggered the limit.
	OnBlock func(r *http.Request, key string)

	// KeyFunc extracts the rate limit key from the request.
	KeyFunc func(r *http.Request) string
}

type shadowMiddleware struct {
	next    http.Handler
	limiter http.Handler
	cfg     ShadowConfig
	mu      sync.Mutex
}

// WithShadowMode wraps a rate limiter in shadow mode: requests are always
// forwarded to the real handler, but the limiter is evaluated in parallel.
// If the limiter would have blocked the request, cfg.OnBlock is invoked.
// This is useful for testing rate limit policies without impacting traffic.
func WithShadowMode(limiter http.Handler, cfg ShadowConfig) func(http.Handler) http.Handler {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = defaultKeyFunc
	}
	if cfg.OnBlock == nil {
		cfg.OnBlock = func(_ *http.Request, _ string) {}
	}
	return func(next http.Handler) http.Handler {
		return &shadowMiddleware{
			next:    next,
			limiter: limiter,
			cfg:     cfg,
		}
	}
}

func (s *shadowMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Evaluate the limiter in a shadow (dry-run) recorder.
	rec := httptest.NewRecorder()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.limiter.ServeHTTP(rec, r)
		if rec.Code == http.StatusTooManyRequests {
			key := s.cfg.KeyFunc(r)
			s.cfg.OnBlock(r, key)
		}
	}()

	// Always serve the real handler regardless of limiter outcome.
	s.next.ServeHTTP(w, r)
	wg.Wait()
}
