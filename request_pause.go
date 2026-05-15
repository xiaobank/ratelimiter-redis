package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

// PauseController allows pausing and resuming request processing across all
// in-flight and incoming requests. While paused, requests are held until
// resumed or the hold deadline is exceeded.
type PauseController struct {
	mu      sync.RWMutex
	paused  bool
	resume  chan struct{}
	maxHold time.Duration
}

// NewPauseController creates a PauseController with the given maximum hold
// duration. Requests that arrive while paused will block for at most maxHold
// before being passed through regardless.
func NewPauseController(maxHold time.Duration) *PauseController {
	if maxHold <= 0 {
		maxHold = 5 * time.Second
	}
	return &PauseController{
		resume:  make(chan struct{}),
		maxHold: maxHold,
	}
}

// Pause halts request processing. Subsequent calls while already paused are
// no-ops.
func (p *PauseController) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.paused {
		p.paused = true
		p.resume = make(chan struct{})
	}
}

// Resume unblocks all waiting requests. Subsequent calls while already
// resumed are no-ops.
func (p *PauseController) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.paused {
		p.paused = false
		close(p.resume)
	}
}

// IsPaused reports whether the controller is currently paused.
func (p *PauseController) IsPaused() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.paused
}

func (p *PauseController) wait() {
	p.mu.RLock()
	if !p.paused {
		p.mu.RUnlock()
		return
	}
	ch := p.resume
	hold := p.maxHold
	p.mu.RUnlock()

	select {
	case <-ch:
	case <-time.After(hold):
	}
}

// WithRequestPause returns middleware that gates requests through the given
// PauseController. While paused, requests block until the controller is
// resumed or the maxHold duration elapses.
func WithRequestPause(pc *PauseController) func(http.Handler) http.Handler {
	if pc == nil {
		panic("ratelimiter: WithRequestPause requires a non-nil PauseController")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pc.wait()
			next.ServeHTTP(w, r)
		})
	}
}
