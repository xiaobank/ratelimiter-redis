package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAdaptiveLimiter_DefaultsToBaseLimit(t *testing.T) {
	al := NewAdaptiveLimiter(AdaptiveConfig{
		BaseLimit: 100,
		MinLimit:  10,
		MaxLimit:  200,
	})
	if al.CurrentLimit() != 100 {
		t.Fatalf("expected 100, got %d", al.CurrentLimit())
	}
}

func TestAdaptiveLimiter_HighLoadReducesLimit(t *testing.T) {
	load := 0.9
	al := NewAdaptiveLimiter(AdaptiveConfig{
		BaseLimit:    100,
		MinLimit:     10,
		MaxLimit:     200,
		HighLoadMark: 0.8,
		LowLoadMark:  0.4,
		LoadFunc:     func() float64 { return load },
	})
	al.Recalculate()
	if al.CurrentLimit() != 10 {
		t.Fatalf("expected min limit 10, got %d", al.CurrentLimit())
	}
}

func TestAdaptiveLimiter_LowLoadIncreasesLimit(t *testing.T) {
	load := 0.1
	al := NewAdaptiveLimiter(AdaptiveConfig{
		BaseLimit:    100,
		MinLimit:     10,
		MaxLimit:     200,
		HighLoadMark: 0.8,
		LowLoadMark:  0.4,
		LoadFunc:     func() float64 { return load },
	})
	al.Recalculate()
	if al.CurrentLimit() != 200 {
		t.Fatalf("expected max limit 200, got %d", al.CurrentLimit())
	}
}

func TestAdaptiveLimiter_MidLoadInterpolates(t *testing.T) {
	load := 0.6 // midpoint between 0.4 and 0.8
	al := NewAdaptiveLimiter(AdaptiveConfig{
		BaseLimit:    100,
		MinLimit:     10,
		MaxLimit:     210,
		HighLoadMark: 0.8,
		LowLoadMark:  0.4,
		LoadFunc:     func() float64 { return load },
	})
	al.Recalculate()
	limit := al.CurrentLimit()
	// ratio = (0.6-0.4)/(0.8-0.4) = 0.5; span = 200; result = 210 - 100 = 110
	if limit != 110 {
		t.Fatalf("expected interpolated limit 110, got %d", limit)
	}
}

func TestAdaptiveLimiter_NilLoadFuncIsNoop(t *testing.T) {
	al := NewAdaptiveLimiter(AdaptiveConfig{
		BaseLimit: 50,
		MinLimit:  5,
		MaxLimit:  100,
	})
	al.Recalculate() // should not panic
	if al.CurrentLimit() != 50 {
		t.Fatalf("expected 50, got %d", al.CurrentLimit())
	}
}

func TestWithAdaptiveLimit_SetsHeader(t *testing.T) {
	al := NewAdaptiveLimiter(AdaptiveConfig{
		BaseLimit:    80,
		MinLimit:     10,
		MaxLimit:     160,
		HighLoadMark: 0.8,
		LowLoadMark:  0.4,
		LoadFunc:     func() float64 { return 0.1 }, // low load -> max
	})

	handler := WithAdaptiveLimit(al, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	hdr := rec.Header().Get("X-RateLimit-Adaptive-Limit")
	if hdr == "" {
		t.Fatal("expected X-RateLimit-Adaptive-Limit header to be set")
	}
	val, err := strconv.Atoi(hdr)
	if err != nil {
		t.Fatalf("header value is not an integer: %s", hdr)
	}
	if val != 160 {
		t.Fatalf("expected 160, got %d", val)
	}
}
