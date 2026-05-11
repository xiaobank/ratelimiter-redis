package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func taggingOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestTagging_SetsHeaderWhenTagPresent(t *testing.T) {
	cfg := ratelimiter.NewTagConfig()
	handler := ratelimiter.WithRequestTagging(cfg)(http.HandlerFunc(taggingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "payments")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Tag"); got != "payments" {
		t.Errorf("expected X-Tag=payments, got %q", got)
	}
}

func TestWithRequestTagging_NoHeaderWhenTagAbsent(t *testing.T) {
	cfg := ratelimiter.NewTagConfig()
	handler := ratelimiter.WithRequestTagging(cfg)(http.HandlerFunc(taggingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Tag"); got != "" {
		t.Errorf("expected no X-Tag header, got %q", got)
	}
}

func TestWithRequestTagging_CustomHeaderName(t *testing.T) {
	cfg := ratelimiter.NewTagConfig(ratelimiter.WithTagHeader("X-My-Tag"))
	handler := ratelimiter.WithRequestTagging(cfg)(http.HandlerFunc(taggingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "search")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-My-Tag"); got != "search" {
		t.Errorf("expected X-My-Tag=search, got %q", got)
	}
}

func TestWithRequestTagging_PrefixIsApplied(t *testing.T) {
	cfg := ratelimiter.NewTagConfig(ratelimiter.WithTagPrefix("svc"))
	handler := ratelimiter.WithRequestTagging(cfg)(http.HandlerFunc(taggingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "orders")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Tag"); got != "svc-orders" {
		t.Errorf("expected X-Tag=svc-orders, got %q", got)
	}
}

func TestWithRequestTagging_CustomKeyFunc(t *testing.T) {
	cfg := ratelimiter.NewTagConfig(
		ratelimiter.WithTagKeyFunc(func(r *http.Request) string {
			return r.URL.Query().Get("tag")
		}),
	)
	handler := ratelimiter.WithRequestTagging(cfg)(http.HandlerFunc(taggingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/?tag=billing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Tag"); got != "billing" {
		t.Errorf("expected X-Tag=billing, got %q", got)
	}
}

func TestWithRequestTagging_PanicsOnNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil config")
		}
	}()
	ratelimiter.WithRequestTagging(nil)
}
