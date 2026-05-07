package ratelimiter_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func sizeOKHandler(w http.ResponseWriter, r *http.Request) {
	// Drain body so MaxBytesReader error surfaces if body is too large.
	_, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestSizeLimit_AllowsSmallBody(t *testing.T) {
	mw := ratelimiter.WithRequestSizeLimit(100)
	handler := mw(http.HandlerFunc(sizeOKHandler))

	body := strings.NewReader("hello")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = int64(body.Len())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestSizeLimit_BlocksLargeContentLength(t *testing.T) {
	mw := ratelimiter.WithRequestSizeLimit(10)
	handler := mw(http.HandlerFunc(sizeOKHandler))

	body := strings.NewReader(strings.Repeat("a", 50))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = 50
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}
}

func TestWithRequestSizeLimit_ZeroDisablesLimit(t *testing.T) {
	mw := ratelimiter.WithRequestSizeLimit(0)
	handler := mw(http.HandlerFunc(sizeOKHandler))

	body := strings.NewReader(strings.Repeat("x", 10_000))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = int64(body.Len())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when limit is 0, got %d", rec.Code)
	}
}

func TestWithRequestSizeLimit_CustomHandler(t *testing.T) {
	customCalled := false
	customHandler := func(w http.ResponseWriter, r *http.Request) {
		customCalled = true
		w.WriteHeader(http.StatusPaymentRequired) // arbitrary sentinel
	}

	mw := ratelimiter.WithRequestSizeLimit(5, ratelimiter.WithSizeLimitHandler(customHandler))
	handler := mw(http.HandlerFunc(sizeOKHandler))

	body := strings.NewReader("this is way too long")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = int64(body.Len())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !customCalled {
		t.Error("expected custom handler to be called")
	}
	if rec.Code != http.StatusPaymentRequired {
		t.Errorf("expected 402, got %d", rec.Code)
	}
}

func TestWithRequestSizeLimit_DefaultResponseBody(t *testing.T) {
	mw := ratelimiter.WithRequestSizeLimit(3)
	handler := mw(http.HandlerFunc(sizeOKHandler))

	body := strings.NewReader("toolong")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = int64(body.Len())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "too large") {
		t.Errorf("expected body to mention 'too large', got: %s", rec.Body.String())
	}
}
