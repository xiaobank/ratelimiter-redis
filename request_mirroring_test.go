package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func mirrorOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestMirroring_PanicsOnEmptyTarget(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on empty MirrorTarget")
		}
	}()
	WithRequestMirroring(MirrorConfig{})
}

func TestWithRequestMirroring_PrimaryResponseUnaffected(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	cfg := MirrorConfig{MirrorTarget: shadow.URL}
	middleware := WithRequestMirroring(cfg)
	handler := middleware(http.HandlerFunc(mirrorOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// wait for the goroutine to finish with a timeout
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("shadow server never received mirrored request")
	}
}

func TestWithRequestMirroring_SendsMirroredHeader(t *testing.T) {
	var gotHeader string
	var wg sync.WaitGroup
	wg.Add(1)

	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Mirrored-Request")
		defer wg.Done()
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	cfg := MirrorConfig{MirrorTarget: shadow.URL}
	middleware := WithRequestMirroring(cfg)
	handler := middleware(http.HandlerFunc(mirrorOKHandler))

	req := httptest.NewRequest(http.MethodPost, "/data", strings.NewReader(`{"key":"value"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for shadow request")
	}

	if gotHeader != "true" {
		t.Errorf("expected X-Mirrored-Request: true, got %q", gotHeader)
	}
}

func TestWithRequestMirroring_ShouldMirrorFiltersRequests(t *testing.T) {
	count := 0
	var mu sync.Mutex

	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	cfg := MirrorConfig{
		MirrorTarget: shadow.URL,
		ShouldMirror: func(r *http.Request) bool { return r.Method == http.MethodGet },
	}
	middleware := WithRequestMirroring(cfg)
	handler := middleware(http.HandlerFunc(mirrorOKHandler))

	// POST should NOT be mirrored
	req := httptest.NewRequest(http.MethodPost, "/data", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	if count != 0 {
		t.Errorf("expected 0 mirror calls for POST, got %d", count)
	}
	mu.Unlock()
}
