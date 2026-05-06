package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func tooManyHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
}

func okHandlerRetry(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRetryAfterMiddleware_SetsHeaderOn429(t *testing.T) {
	window := 60 * time.Second
	handler := ratelimiter.RetryAfterMiddleware(
		http.HandlerFunc(tooManyHandler),
		ratelimiter.FixedRetryAfter(window),
		window,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	got := rr.Header().Get("Retry-After")
	if got == "" {
		t.Fatal("expected Retry-After header to be set")
	}
	val, err := strconv.Atoi(got)
	if err != nil {
		t.Fatalf("Retry-After is not an integer: %s", got)
	}
	if val != 60 {
		t.Errorf("expected Retry-After=60, got %d", val)
	}
}

func TestRetryAfterMiddleware_NoHeaderOnOK(t *testing.T) {
	window := 60 * time.Second
	handler := ratelimiter.RetryAfterMiddleware(
		http.HandlerFunc(okHandlerRetry),
		ratelimiter.FixedRetryAfter(window),
		window,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Retry-After"); got != "" {
		t.Errorf("expected no Retry-After header on 200, got %s", got)
	}
}

func TestRetryAfterMiddleware_NilCalculatorUsesWindow(t *testing.T) {
	window := 30 * time.Second
	handler := ratelimiter.RetryAfterMiddleware(
		http.HandlerFunc(tooManyHandler),
		nil,
		window,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Retry-After")
	val, _ := strconv.Atoi(got)
	if val != 30 {
		t.Errorf("expected Retry-After=30, got %d", val)
	}
}
