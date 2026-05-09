package ratelimiter

import "net/http"

// WithPriorityKeyFunc sets a custom function to extract the priority level
// from an incoming request.
func WithPriorityKeyFunc(fn PriorityKeyFunc) func(*PriorityConfig) {
	return func(c *PriorityConfig) {
		if fn != nil {
			c.keyFunc = fn
		}
	}
}

// WithHighLimitMultiplier sets the multiplier applied to the base limit for
// high-priority requests. Must be >= 1.0; defaults to 2.0.
func WithHighLimitMultiplier(m float64) func(*PriorityConfig) {
	return func(c *PriorityConfig) {
		if m >= 1.0 {
			c.highLimitMult = m
		}
	}
}

// WithLowLimitMultiplier sets the multiplier applied to the base limit for
// low-priority requests. Must be in (0, 1]; defaults to 0.5.
func WithLowLimitMultiplier(m float64) func(*PriorityConfig) {
	return func(c *PriorityConfig) {
		if m > 0 && m <= 1.0 {
			c.lowLimitMult = m
		}
	}
}

// WithPriorityExceededHandler sets a custom HTTP handler invoked when the
// effective rate limit for a priority tier is exceeded.
func WithPriorityExceededHandler(h http.HandlerFunc) func(*PriorityConfig) {
	return func(c *PriorityConfig) {
		if h != nil {
			c.exceededHandler = h
		}
	}
}
