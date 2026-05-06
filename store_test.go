package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	ratelimiter "github.com/you/ratelimiter-redis"
)

func TestMemoryStore_IncrementAndGet(t *testing.T) {
	s := ratelimiter.NewMemoryStore()
	ctx := context.Background()

	v1, err := s.Increment(ctx, "key", time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v1 != 1 {
		t.Fatalf("expected 1, got %d", v1)
	}

	v2, err := s.Increment(ctx, "key", time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("expected 2, got %d", v2)
	}

	got, err := s.Get(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestMemoryStore_SetAndGet(t *testing.T) {
	s := ratelimiter.NewMemoryStore()
	ctx := context.Background()

	if err := s.Set(ctx, "k", 42, time.Minute); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	v, _ := s.Get(ctx, "k")
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestMemoryStore_SetNX(t *testing.T) {
	s := ratelimiter.NewMemoryStore()
	ctx := context.Background()

	ok, err := s.SetNX(ctx, "nx", 1, time.Minute)
	if err != nil || !ok {
		t.Fatalf("expected SetNX to succeed on new key")
	}
	ok, err = s.SetNX(ctx, "nx", 2, time.Minute)
	if err != nil || ok {
		t.Fatalf("expected SetNX to fail on existing key")
	}
}

func TestMemoryStore_ExpiresEntries(t *testing.T) {
	s := ratelimiter.NewMemoryStore()
	ctx := context.Background()

	_, _ = s.Increment(ctx, "exp", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	v, _ := s.Get(ctx, "exp")
	if v != 0 {
		t.Fatalf("expected 0 after expiry, got %d", v)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := ratelimiter.NewMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "del", 5, time.Minute)
	_ = s.Delete(ctx, "del")
	v, _ := s.Get(ctx, "del")
	if v != 0 {
		t.Fatalf("expected 0 after delete, got %d", v)
	}
}

func TestMemoryStore_Ping(t *testing.T) {
	s := ratelimiter.NewMemoryStore()
	if err := s.Ping(context.Background()); err != nil {
		t.Fatalf("Ping should not fail: %v", err)
	}
}

func TestMemoryStore_ImplementsStore(t *testing.T) {
	var _ ratelimiter.Store = ratelimiter.NewMemoryStore()
}
