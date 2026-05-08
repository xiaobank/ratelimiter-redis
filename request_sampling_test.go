package ratelimiter_test

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "."
)

func samplingOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestSampling_AllowsAllAtRateOne(t *testing.T) {
	mw := ratelimiter.WithRequestSampling(ratelimiter.SamplingConfig{
		Rate:       1.0,
		RandSource: rand.NewSource(42),
	})
	handler := mw(http.HandlerFunc(samplingOKHandler))

	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d on iteration %d", rec.Code, i)
		}
	}
}

func TestWithRequestSampling_BlocksAllAtRateZero(t *testing.T) {
	mw := ratelimiter.WithRequestSampling(ratelimiter.SamplingConfig{
		Rate:       0.0,
		RandSource: rand.NewSource(42),
	})
	handler := mw(http.HandlerFunc(samplingOKHandler))

	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d on iteration %d", rec.Code, i)
		}
	}
}

func TestWithRequestSampling_PartialRate(t *testing.T) {
	mw := ratelimiter.WithRequestSampling(ratelimiter.SamplingConfig{
		Rate:       0.5,
		RandSource: rand.NewSource(99),
	})
	handler := mw(http.HandlerFunc(samplingOKHandler))

	allowed, blocked := 0, 0
	for i := 0; i < 100; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			allowed++
		} else {
			blocked++
		}
	}
	if allowed == 0 || blocked == 0 {
		t.Fatalf("expected mix of allowed/blocked at 50%% rate, got allowed=%d blocked=%d", allowed, blocked)
	}
}

func TestWithRequestSampling_CustomRejectedHandler(t *testing.T) {
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	mw := ratelimiter.WithRequestSampling(ratelimiter.SamplingConfig{
		Rate:            0.0,
		RandSource:      rand.NewSource(1),
		RejectedHandler: customHandler,
	})
	handler := mw(http.HandlerFunc(samplingOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 from custom handler, got %d", rec.Code)
	}
}
