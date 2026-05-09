package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ratelimiter "github.com/your-org/ratelimiter-redis"
)

func throttleOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestThrottle_AllowsUnderConcurrentLimit(t *testing.T) {
	mw := ratelimiter.WithRequestThrottle(5, 10, time.Second)
	handler := mw(http.HandlerFunc(throttleOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestThrottle_BlocksWhenQueueFull(t *testing.T) {
	// maxConcurrent=1, queueSize=0 so the second request is immediately rejected.
	blocked := make(chan struct{})
	release := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(blocked)
		<-release
		w.WriteHeader(http.StatusOK)
	})

	mw := ratelimiter.WithRequestThrottle(1, 0, 100*time.Millisecond)
	handler := mw(slow)

	// First request occupies the only slot.
	go func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}()

	<-blocked // ensure first request is in-flight

	// Second request should be rejected immediately (queue size 0).
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec2, req2)

	close(release)

	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec2.Code)
	}
}

func TestWithRequestThrottle_CustomExceededHandler(t *testing.T) {
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	blocked := make(chan struct{})
	release := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(blocked)
		<-release
		w.WriteHeader(http.StatusOK)
	})

	mw := ratelimiter.WithRequestThrottle(1, 0, 100*time.Millisecond,
		ratelimiter.WithThrottleExceededHandler(customHandler),
	)
	handler := mw(slow)

	go func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}()
	<-blocked

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec2, req2)
	close(release)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}

func TestWithRequestThrottle_ConcurrentRequestsDoNotExceedMax(t *testing.T) {
	const maxConcurrent = 3
	var active int64
	var maxSeen int64

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt64(&active, 1)
		for {
			old := atomic.LoadInt64(&maxSeen)
			if cur <= old || atomic.CompareAndSwapInt64(&maxSeen, old, cur) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(&active, -1)
		w.WriteHeader(http.StatusOK)
	})

	mw := ratelimiter.WithRequestThrottle(maxConcurrent, 20, 500*time.Millisecond)
	handler := mw(slow)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			handler.ServeHTTP(rec, req)
		}()
	}
	wg.Wait()

	if atomic.LoadInt64(&maxSeen) > maxConcurrent {
		t.Fatalf("max concurrent exceeded: got %d, want <= %d", maxSeen, maxConcurrent)
	}
}
