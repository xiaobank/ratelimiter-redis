package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/your-org/ratelimiter-redis"
)

func TestWithGeoFilter_PanicsOnNilFilter(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when GeoFilter is nil")
		}
	}()
	ratelimiter.WithGeoFilter(nil, nil)
}

func TestWithGeoFilter_AllowListMode_BlocksUnknown(t *testing.T) {
	gf := ratelimiter.WithAllowedCountries("GB")
	mw := ratelimiter.WithGeoFilter(gf, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with no country header → empty string → not in allow list → blocked
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newGeoRequest(""))
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unknown country, got %d", rec.Code)
	}
}

func TestWithGeoFilter_AllowListMode_AllowsKnown(t *testing.T) {
	gf := ratelimiter.WithAllowedCountries("GB", "US")
	mw := ratelimiter.WithGeoFilter(gf, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, newGeoRequest("GB"))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for GB, got %d", rec.Code)
	}
}

func TestWithGeoFilter_XRealIPFallback(t *testing.T) {
	gf := ratelimiter.WithBlockedCountries("JP")
	// Override resolver to use X-Country-Code header (second fallback).
	gf.SetCountryCodeFunc(func(r *http.Request) string {
		return r.Header.Get("X-Country-Code")
	})
	mw := ratelimiter.WithGeoFilter(gf, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "JP")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for JP via X-Country-Code, got %d", rec.Code)
	}
}
