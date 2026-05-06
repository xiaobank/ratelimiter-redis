package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestFixedWindowCounter_AllowsUnderLimit(t *testing.T) {
	client := newTestClient(t)
	fw := NewFixedWindowCounter(client, time.Minute, 5)
	ctx := context.Background()
	key := "test-fw-allow"

	for i := 0; i < 5; i++ {
		allowed, remaining, _, err := fw.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		expected := 5 - (i + 1)
		if remaining != expected {
			t.Errorf("expected remaining %d, got %d", expected, remaining)
		}
	}
}

func TestFixedWindowCounter_BlocksOverLimit(t *testing.T) {
	client := newTestClient(t)
	fw := NewFixedWindowCounter(client, time.Minute, 3)
	ctx := context.Background()
	key := "test-fw-block"

	for i := 0; i < 3; i++ {
		_, _, _, err := fw.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
	}

	allowed, remaining, _, err := fw.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be blocked after limit exceeded")
	}
	if remaining != 0 {
		t.Errorf("expected remaining 0, got %d", remaining)
	}
}

func TestFixedWindowCounter_ResetsAfterWindow(t *testing.T) {
	client := newTestClient(t)
	// Use a very short window for testing
	fw := NewFixedWindowCounter(client, 2*time.Second, 2)
	ctx := context.Background()
	key := "test-fw-reset"

	for i := 0; i < 2; i++ {
		_, _, _, err := fw.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	allowed, _, _, _ := fw.Allow(ctx, key)
	if allowed {
		t.Error("expected request to be blocked before window reset")
	}

	time.Sleep(3 * time.Second)

	allowed, _, _, err := fw.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed after window reset")
	}
}
