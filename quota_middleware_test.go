package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func quotaOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithQuotaManager_AllowsRequestsUnderLimit(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil,
		ratelimiter.Quota{Name: "api", Limit: 5, Window: time.Minute},
	)

	handler := ratelimiter.WithQuotaManager(qm, nil)(http.HandlerFunc(quotaOKHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestWithQuotaManager_BlocksRequestsOverLimit(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil,
		ratelimiter.Quota{Name: "api", Limit: 2, Window: time.Minute},
	)

	handler := ratelimiter.WithQuotaManager(qm, nil)(http.HandlerFunc(quotaOKHandler))

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		handler.ServeHTTP(rr, req)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}
	if v := rr.Header().Get("X-Quota-Exceeded"); v != "api" {
		t.Errorf("expected X-Quota-Exceeded=api, got %q", v)
	}
}

func TestWithQuotaManager_SetsRemainingHeaders(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil,
		ratelimiter.Quota{Name: "hourly", Limit: 10, Window: time.Hour},
	)

	handler := ratelimiter.WithQuotaManager(qm, nil)(http.HandlerFunc(quotaOKHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.3:1234"
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if v := rr.Header().Get("X-Quota-hourly-Remaining"); v == "" {
		t.Error("expected X-Quota-hourly-Remaining header to be set")
	}
}
