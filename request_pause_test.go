package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	ratelimiter "."
)

func pauseOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestPause_PassesThroughWhenNotPaused(t *testing.T) {
	pc := ratelimiter.NewPauseController(time.Second)
	mw := ratelimiter.WithRequestPause(pc)(http.HandlerFunc(pauseOKHandler))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestPause_BlocksWhilePaused(t *testing.T) {
	pc := ratelimiter.NewPauseController(2 * time.Second)
	mw := ratelimiter.WithRequestPause(pc)(http.HandlerFunc(pauseOKHandler))

	pc.Pause()

	done := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done <- rec.Code
	}()

	select {
	case <-done:
		t.Fatal("expected request to be blocked while paused")
	case <-time.After(80 * time.Millisecond):
		// still blocked — correct
	}

	pc.Resume()

	select {
	case code := <-done:
		if code != http.StatusOK {
			t.Fatalf("expected 200 after resume, got %d", code)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not complete after resume")
	}
}

func TestWithRequestPause_UnblocksAfterMaxHold(t *testing.T) {
	pc := ratelimiter.NewPauseController(100 * time.Millisecond)
	mw := ratelimiter.WithRequestPause(pc)(http.HandlerFunc(pauseOKHandler))

	pc.Pause()

	start := time.Now()
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	elapsed := time.Since(start)

	if elapsed < 90*time.Millisecond {
		t.Fatalf("expected to wait ~100ms, waited %v", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after hold expiry, got %d", rec.Code)
	}
}

func TestWithRequestPause_ConcurrentRequestsAllUnblocked(t *testing.T) {
	pc := ratelimiter.NewPauseController(2 * time.Second)
	mw := ratelimiter.WithRequestPause(pc)(http.HandlerFunc(pauseOKHandler))

	pc.Pause()

	const n = 10
	var wg sync.WaitGroup
	codes := make([]int, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			codes[idx] = rec.Code
		}(i)
	}

	time.Sleep(40 * time.Millisecond)
	pc.Resume()
	wg.Wait()

	for i, c := range codes {
		if c != http.StatusOK {
			t.Errorf("goroutine %d: expected 200, got %d", i, c)
		}
	}
}

func TestWithRequestPause_PanicsOnNilController(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil PauseController")
		}
	}()
	ratelimiter.WithRequestPause(nil)
}
