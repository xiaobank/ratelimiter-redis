package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func shadowOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithShadowMode_AlwaysForwardsToRealHandler(t *testing.T) {
	// A limiter that always blocks.
	alwaysBlock := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	middleware := ratelimiter.WithShadowMode(alwaysBlock, ratelimiter.ShadowConfig{})
	handler := middleware(http.HandlerFunc(shadowOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithShadowMode_InvokesOnBlockWhenLimiterBlocks(t *testing.T) {
	var blocked int32

	alwaysBlock := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	cfg := ratelimiter.NewShadowConfig(
		ratelimiter.WithShadowOnBlock(func(_ *http.Request, _ string) {
			atomic.AddInt32(&blocked, 1)
		}),
	)

	middleware := ratelimiter.WithShadowMode(alwaysBlock, cfg)
	handler := middleware(http.HandlerFunc(shadowOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Allow goroutine to finish.
	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt32(&blocked) != 1 {
		t.Fatalf("expected OnBlock to be called once, got %d", blocked)
	}
}

func TestWithShadowMode_DoesNotInvokeOnBlockWhenAllowed(t *testing.T) {
	var blocked int32

	alwaysAllow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := ratelimiter.NewShadowConfig(
		ratelimiter.WithShadowOnBlock(func(_ *http.Request, _ string) {
			atomic.AddInt32(&blocked, 1)
		}),
	)

	middleware := ratelimiter.WithShadowMode(alwaysAllow, cfg)
	handler := middleware(http.HandlerFunc(shadowOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt32(&blocked) != 0 {
		t.Fatalf("expected OnBlock not to be called, got %d", blocked)
	}
}

func TestWithShadowMode_CustomKeyFuncPassedToOnBlock(t *testing.T) {
	var capturedKey string

	alwaysBlock := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	cfg := ratelimiter.NewShadowConfig(
		ratelimiter.WithShadowKeyFunc(func(r *http.Request) string {
			return "custom-key"
		}),
		ratelimiter.WithShadowOnBlock(func(_ *http.Request, key string) {
			capturedKey = key
		}),
	)

	middleware := ratelimiter.WithShadowMode(alwaysBlock, cfg)
	handler := middleware(http.HandlerFunc(shadowOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	time.Sleep(20 * time.Millisecond)

	if capturedKey != "custom-key" {
		t.Fatalf("expected key 'custom-key', got %q", capturedKey)
	}
}
