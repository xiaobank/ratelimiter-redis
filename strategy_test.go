package ratelimiter

import (
	"testing"
	"time"
)

func TestNewStrategy_ReturnsFixedWindow(t *testing.T) {
	client := newTestClient(t)
	s := NewStrategy(StrategyFixedWindow, client, time.Minute, 10)

	if _, ok := s.(*FixedWindowCounter); !ok {
		t.Errorf("expected *FixedWindowCounter, got %T", s)
	}
}

func TestNewStrategy_ReturnsSlidingWindow(t *testing.T) {
	client := newTestClient(t)
	s := NewStrategy(StrategySlidingWindow, client, time.Minute, 10)

	if _, ok := s.(*SlidingWindowCounter); !ok {
		t.Errorf("expected *SlidingWindowCounter, got %T", s)
	}
}

func TestNewStrategy_DefaultsToSlidingWindow(t *testing.T) {
	client := newTestClient(t)
	s := NewStrategy(StrategyType("unknown"), client, time.Minute, 10)

	if _, ok := s.(*SlidingWindowCounter); !ok {
		t.Errorf("expected *SlidingWindowCounter as default, got %T", s)
	}
}

func TestStrategy_FixedWindowAllow(t *testing.T) {
	client := newTestClient(t)
	s := NewStrategy(StrategyFixedWindow, client, time.Minute, 5)

	for i := 0; i < 5; i++ {
		allowed, _, _, err := s.Allow(t.Context(), "strategy-fw-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	allowed, _, _, err := s.Allow(t.Context(), "strategy-fw-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected 6th request to be blocked")
	}
}

func TestStrategy_SlidingWindowAllow(t *testing.T) {
	client := newTestClient(t)
	s := NewStrategy(StrategySlidingWindow, client, time.Minute, 5)

	for i := 0; i < 5; i++ {
		allowed, _, _, err := s.Allow(t.Context(), "strategy-sw-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	allowed, _, _, err := s.Allow(t.Context(), "strategy-sw-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected 6th request to be blocked")
	}
}
