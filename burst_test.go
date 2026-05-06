package ratelimiter_test

import (
	"context"
	"testing"
	"time"
)

func TestBurstCounter_AllowsUnderLimit(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 3, 2, time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		allowed, _, err := counter.Allow(ctx, "user:1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestBurstCounter_AllowsBurstRequests(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 3, 2, time.Minute)
	ctx := context.Background()

	// Use base limit + burst (5 total)
	for i := 0; i < 5; i++ {
		allowed, _, err := counter.Allow(ctx, "user:burst")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("burst request %d should be allowed", i+1)
		}
	}
}

func TestBurstCounter_BlocksOverBurst(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 3, 2, time.Minute)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		counter.Allow(ctx, "user:over") //nolint
	}

	allowed, remaining, err := counter.Allow(ctx, "user:over")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("6th request should be blocked")
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}
}

func TestBurstCounter_ReturnsCorrectRemaining(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 4, 1, time.Minute)
	ctx := context.Background()

	_, remaining, _ := counter.Allow(ctx, "user:rem")
	if remaining != 4 {
		t.Fatalf("expected 4 remaining after 1st request, got %d", remaining)
	}
}

func TestBurstCounter_LimitAndBurstAccessors(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 10, 5, time.Minute)

	if counter.Limit() != 10 {
		t.Fatalf("expected limit 10, got %d", counter.Limit())
	}
	if counter.Burst() != 5 {
		t.Fatalf("expected burst 5, got %d", counter.Burst())
	}
}
