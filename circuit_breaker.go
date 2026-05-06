package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker trips open when the rate limiter blocks too many
// consecutive requests, giving the backend a chance to recover.
type CircuitBreaker struct {
	mu              sync.Mutex
	state           CircuitState
	failures        int
	threshold       int
	resetAfter      time.Duration
	lastFailureTime time.Time
	onTrip          func()
	onReset         func()
}

// CircuitBreakerOption configures a CircuitBreaker.
type CircuitBreakerOption func(*CircuitBreaker)

// WithTripThreshold sets the number of consecutive blocked requests
// before the circuit opens.
func WithTripThreshold(n int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) { cb.threshold = n }
}

// WithResetAfter sets how long to wait before attempting half-open.
func WithResetAfter(d time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) { cb.resetAfter = d }
}

// WithOnTrip registers a callback invoked when the circuit opens.
func WithOnTrip(fn func()) CircuitBreakerOption {
	return func(cb *CircuitBreaker) { cb.onTrip = fn }
}

// WithOnReset registers a callback invoked when the circuit closes.
func WithOnReset(fn func()) CircuitBreakerOption {
	return func(cb *CircuitBreaker) { cb.onReset = fn }
}

// NewCircuitBreaker returns a CircuitBreaker with sensible defaults.
func NewCircuitBreaker(opts ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		threshold:  5,
		resetAfter: 30 * time.Second,
	}
	for _, o := range opts {
		o(cb)
	}
	return cb
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// currentState must be called with cb.mu held.
func (cb *CircuitBreaker) currentState() CircuitState {
	if cb.state == CircuitOpen && time.Since(cb.lastFailureTime) >= cb.resetAfter {
		cb.state = CircuitHalfOpen
	}
	return cb.state
}

// RecordBlocked records a blocked request and may trip the circuit.
func (cb *CircuitBreaker) RecordBlocked() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailureTime = time.Now()
	if cb.state == CircuitClosed && cb.failures >= cb.threshold {
		cb.state = CircuitOpen
		if cb.onTrip != nil {
			go cb.onTrip()
		}
	}
}

// RecordAllowed resets the failure counter and closes the circuit.
func (cb *CircuitBreaker) RecordAllowed() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == CircuitHalfOpen || cb.failures > 0 {
		prevState := cb.state
		cb.state = CircuitClosed
		cb.failures = 0
		if prevState == CircuitHalfOpen && cb.onReset != nil {
			go cb.onReset()
		}
	}
}

// WithCircuitBreaker wraps an existing handler with circuit-breaker logic.
// When the circuit is open, requests receive 503 Service Unavailable.
func WithCircuitBreaker(next http.Handler, cb *CircuitBreaker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cb.State() == CircuitOpen {
			http.Error(w, "service temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
