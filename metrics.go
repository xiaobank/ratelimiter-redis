package ratelimiter

import "sync/atomic"

// Metrics holds counters for rate limiter activity.
type Metrics struct {
	allowed  atomic.Int64
	blocked  atomic.Int64
	total    atomic.Int64
}

// Allowed returns the number of requests that were allowed.
func (m *Metrics) Allowed() int64 {
	return m.allowed.Load()
}

// Blocked returns the number of requests that were blocked.
func (m *Metrics) Blocked() int64 {
	return m.blocked.Load()
}

// Total returns the total number of requests seen.
func (m *Metrics) Total() int64 {
	return m.total.Load()
}

// Reset zeroes all counters.
func (m *Metrics) Reset() {
	m.allowed.Store(0)
	m.blocked.Store(0)
	m.total.Store(0)
}

func (m *Metrics) recordAllowed() {
	m.total.Add(1)
	m.allowed.Add(1)
}

func (m *Metrics) recordBlocked() {
	m.total.Add(1)
	m.blocked.Add(1)
}
