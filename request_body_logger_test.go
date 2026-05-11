package ratelimiter_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	ratelimiter "ratelimiter-redis"
)

func bodyLoggerOKHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("downstream handler could not read body: %v", err)
		}
		w.Write(body)
	})
}

func TestWithRequestBodyLogger_CapturesBody(t *testing.T) {
	var mu sync.Mutex
	var captured []byte

	mw := ratelimiter.WithRequestBodyLogger(
		ratelimiter.WithBodyLogFunc(func(r *http.Request, body []byte) {
			mu.Lock()
			captured = append([]byte(nil), body...)
			mu.Unlock()
		}),
	)

	handler := mw(bodyLoggerOKHandler(t))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello world"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	mu.Lock()
	defer mu.Unlock()
	if string(captured) != "hello world" {
		t.Errorf("expected captured body %q, got %q", "hello world", captured)
	}
}

func TestWithRequestBodyLogger_RestoresBodyForDownstream(t *testing.T) {
	mw := ratelimiter.WithRequestBodyLogger()
	handler := mw(bodyLoggerOKHandler(t))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("downstream payload"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "downstream payload" {
		t.Errorf("expected downstream to receive full body, got %q", rec.Body.String())
	}
}

func TestWithRequestBodyLogger_RespectsMaxBytes(t *testing.T) {
	var mu sync.Mutex
	var captured []byte

	mw := ratelimiter.WithRequestBodyLogger(
		ratelimiter.WithBodyLogMaxBytes(5),
		ratelimiter.WithBodyLogFunc(func(r *http.Request, body []byte) {
			mu.Lock()
			captured = append([]byte(nil), body...)
			mu.Unlock()
		}),
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello world"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	mu.Lock()
	defer mu.Unlock()
	if len(captured) > 5 {
		t.Errorf("expected at most 5 bytes captured, got %d", len(captured))
	}
}

func TestWithRequestBodyLogger_SetsHeader(t *testing.T) {
	mw := ratelimiter.WithRequestBodyLogger(
		ratelimiter.WithBodyLogHeader("X-Body-Bytes"),
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("test"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Body-Bytes") != "4" {
		t.Errorf("expected X-Body-Bytes header to be '4', got %q", rec.Header().Get("X-Body-Bytes"))
	}
}

func TestWithRequestBodyLogger_NilBodyIsNoop(t *testing.T) {
	called := false
	mw := ratelimiter.WithRequestBodyLogger(
		ratelimiter.WithBodyLogFunc(func(r *http.Request, body []byte) {
			called = true
		}),
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Body = nil
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Error("expected logFunc not to be called for nil body")
	}
}
