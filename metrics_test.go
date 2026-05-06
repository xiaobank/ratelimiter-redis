package ratelimiter

import "testing"

func TestMetrics_InitialValues(t *testing.T) {
	m := &Metrics{}
	if m.Allowed() != 0 || m.Blocked() != 0 || m.Total() != 0 {
		t.Fatal("expected all counters to start at zero")
	}
}

func TestMetrics_RecordAllowed(t *testing.T) {
	m := &Metrics{}
	m.recordAllowed()
	m.recordAllowed()
	if m.Allowed() != 2 {
		t.Fatalf("expected allowed=2, got %d", m.Allowed())
	}
	if m.Total() != 2 {
		t.Fatalf("expected total=2, got %d", m.Total())
	}
	if m.Blocked() != 0 {
		t.Fatalf("expected blocked=0, got %d", m.Blocked())
	}
}

func TestMetrics_RecordBlocked(t *testing.T) {
	m := &Metrics{}
	m.recordBlocked()
	if m.Blocked() != 1 {
		t.Fatalf("expected blocked=1, got %d", m.Blocked())
	}
	if m.Total() != 1 {
		t.Fatalf("expected total=1, got %d", m.Total())
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := &Metrics{}
	m.recordAllowed()
	m.recordBlocked()
	m.Reset()
	if m.Allowed() != 0 || m.Blocked() != 0 || m.Total() != 0 {
		t.Fatal("expected all counters to be zero after Reset")
	}
}
