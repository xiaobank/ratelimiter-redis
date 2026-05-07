package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	ratelimiter "github.com/your-org/ratelimiter-redis"
)

func coalescingOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestCoalescingGroup_SingleCall(t *testing.T) {
	g := ratelimiter.NewCoalescingGroup()
	calls := 0
	val, coalesced, err := g.Do("key1", func() (int, error) {
		calls++
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coalesced {
		t.Error("first call should not be coalesced")
	}
	if val != 42 {
		t.Errorf("expected val 42, got %d", val)
	}
	if calls != 1 {
		t.Errorf("expected 1 fn call, got %d", calls)
	}
}

func TestCoalescingGroup_ConcurrentCallsShareResult(t *testing.T) {
	g := ratelimiter.NewCoalescingGroup()
	var mu sync.Mutex
	calls := 0
	coalesced := 0

	ready := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			_, c, _ := g.Do("shared", func() (int, error) {
				mu.Lock()
				calls++
				mu.Unlock()
				return 1, nil
			})
			if c {
				mu.Lock()
				coalesced++
				mu.Unlock()
			}
		}()
	}
	close(ready)
	wg.Wait()

	if calls == 0 {
		t.Error("expected at least one fn invocation")
	}
	if calls+coalesced != 10 {
		t.Errorf("calls(%d)+coalesced(%d) should equal 10", calls, coalesced)
	}
}

func TestWithRequestCoalescing_SetsHeader(t *testing.T) {
	g := ratelimiter.NewCoalescingGroup()
	keyFunc := func(r *http.Request) (string, error) { return "test-key", nil }

	handler := ratelimiter.WithRequestCoalescing(g, keyFunc,
		http.HandlerFunc(coalescingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Coalesced"); got != "false" {
		t.Errorf("expected X-Coalesced: false, got %q", got)
	}
}

func TestWithRequestCoalescing_PanicsOnNilGroup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil CoalescingGroup")
		}
	}()
	ratelimiter.WithRequestCoalescing(nil, func(r *http.Request) (string, error) { return "", nil },
		http.HandlerFunc(coalescingOKHandler))
}

func TestNewCoalescingGroupWithOptions_DefaultKeyFunc(t *testing.T) {
	g, kf := ratelimiter.NewCoalescingGroupWithOptions()
	if g == nil {
		t.Error("expected non-nil CoalescingGroup")
	}
	if kf == nil {
		t.Error("expected non-nil KeyFunc")
	}
}
