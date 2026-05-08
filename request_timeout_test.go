package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func slowHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		case <-r.Context().Done():
			// context cancelled — do nothing, middleware will respond
		}
	})
}

func TestWithRequestTimeout_AllowsFastHandler(t *testing.T) {
	middleware := WithRequestTimeout(100 * time.Millisecond)
	handler := middleware(slowHandler(10 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_BlocksSlowHandler(t *testing.T) {
	middleware := WithRequestTimeout(20 * time.Millisecond)
	handler := middleware(slowHandler(200 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_ZeroDisablesTimeout(t *testing.T) {
	middleware := WithRequestTimeout(0)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_CustomHandler(t *testing.T) {
	customCalled := false
	custom := func(w http.ResponseWriter, r *http.Request) {
		customCalled = true
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	middleware := WithRequestTimeout(20*time.Millisecond, WithTimeoutHandler(custom))
	handler := middleware(slowHandler(200 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if !customCalled {
		t.Fatal("expected custom timeout handler to be called")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestWithRequestTimeout_NilCustomHandlerFallsBack(t *testing.T) {
	middleware := WithRequestTimeout(20*time.Millisecond, WithTimeoutHandler(nil))
	handler := middleware(slowHandler(200 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected default 504, got %d", rec.Code)
	}
}
