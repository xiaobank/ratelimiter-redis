package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

// ThrottleConfig holds configuration for the request throttle middleware.
type ThrottleConfig struct {
	maxConcurrent int
	queueTimeout  time.Duration
	exceededHandler http.Handler
	mu            sync.Mutex
	active        int
	queue         chan struct{}
}

// defaultThrottleExceededHandler returns 503 when the queue is full.
func defaultThrottleExceededHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
}

// WithRequestThrottle limits the number of concurrently handled requests.
// Requests beyond maxConcurrent are queued up to the queue capacity.
// If the queue is full or the request waits longer than queueTimeout, the
// exceededHandler is invoked instead.
func WithRequestThrottle(maxConcurrent, queueSize int, queueTimeout time.Duration, opts ...func(*ThrottleConfig)) func(http.Handler) http.Handler {
	cfg := &ThrottleConfig{
		maxConcurrent:   maxConcurrent,
		queueTimeout:    queueTimeout,
		exceededHandler: http.HandlerFunc(defaultThrottleExceededHandler),
		queue:           make(chan struct{}, queueSize),
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to acquire a slot immediately.
			cfg.mu.Lock()
			if cfg.active < cfg.maxConcurrent {
				cfg.active++
				cfg.mu.Unlock()
				defer func() {
					cfg.mu.Lock()
					cfg.active--
					cfg.mu.Unlock()
				}()
				next.ServeHTTP(w, r)
				return
			}
			cfg.mu.Unlock()

			// Queue the request.
			timer := time.NewTimer(cfg.queueTimeout)
			defer timer.Stop()

			select {
			case cfg.queue <- struct{}{}:
				// Wait until we can run.
				select {
				case <-timer.C:
					<-cfg.queue
					cfg.exceededHandler.ServeHTTP(w, r)
				case <-r.Context().Done():
					<-cfg.queue
				}
			default:
				cfg.exceededHandler.ServeHTTP(w, r)
			}
		})
	}
}
