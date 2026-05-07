package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/your-org/ratelimiter-redis"
)

func newGeoRequest(cc string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if cc != "" {
		req.Header.Set("CF-IPCountry", cc)
	}
	return req
}

func TestGeoFilter_BlockedCountry_IsBlocked(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("CN", "RU")
	if gf.Allowed(newGeoRequest("CN")) {
		t.Error("expected CN to be blocked")
	}
	if gf.Allowed(newGeoRequest("RU")) {
		t.Error("expected RU to be blocked")
	}
}

func TestGeoFilter_UnblockedCountry_IsAllowed(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("CN")
	if !gf.Allowed(newGeoRequest("US")) {
		t.Error("expected US to be allowed")
	}
}

func TestGeoFilter_AllowedCountry_IsAllowed(t *testing.T) {
	gf := ratelimiter.WithAllowedCountries("US", "DE")
	if !gf.Allowed(newGeoRequest("US")) {
		t.Error("expected US to be allowed")
	}
}

func TestGeoFilter_AllowedCountry_BlocksOthers(t *testing.T) {
	gf := ratelimiter.WithAllowedCountries("US")
	if gf.Allowed(newGeoRequest("CN")) {
		t.Error("expected CN to be blocked when not in allow-list")
	}
}

func TestGeoFilter_CustomCountryCodeFunc(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("FR")
	gf.SetCountryCodeFunc(func(r *http.Request) string {
		return r.Header.Get("X-My-Country")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-My-Country", "FR")
	if gf.Allowed(req) {
		t.Error("expected FR to be blocked via custom resolver")
	}
}

func TestWithGeoFilter_Returns403WhenBlocked(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("CN")
	mw := ratelimiter.WithGeoFilter(gf, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newGeoRequest("CN"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestWithGeoFilter_PassesThroughAllowedRequest(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("CN")
	mw := ratelimiter.WithGeoFilter(gf, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newGeoRequest("US"))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWithGeoFilter_CustomBlockedHandler(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("CN")
	called := false
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	mw := ratelimiter.WithGeoFilter(gf, customHandler)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newGeoRequest("CN"))
	if !called {
		t.Error("expected custom blocked handler to be called")
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("expected 418, got %d", rec.Code)
	}
}
