package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRequest(remoteAddr, xff, xri string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	if xri != "" {
		req.Header.Set("X-Real-IP", xri)
	}
	return req
}

func TestIPKeyFunc_UsesRemoteAddr(t *testing.T) {
	fn := IPKeyFunc(false)
	req := newRequest("192.168.1.10:54321", "", "")
	got := fn(req)
	if got != "192.168.1.10" {
		t.Errorf("expected 192.168.1.10, got %s", got)
	}
}

func TestIPKeyFunc_IgnoresProxyHeadersWhenNotTrusted(t *testing.T) {
	fn := IPKeyFunc(false)
	req := newRequest("10.0.0.1:1234", "203.0.113.5", "")
	got := fn(req)
	if got != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", got)
	}
}

func TestIPKeyFunc_UsesXForwardedForWhenTrusted(t *testing.T) {
	fn := IPKeyFunc(true)
	req := newRequest("10.0.0.1:1234", "203.0.113.5, 10.0.0.1", "")
	got := fn(req)
	if got != "203.0.113.5" {
		t.Errorf("expected 203.0.113.5, got %s", got)
	}
}

func TestIPKeyFunc_UsesXRealIPWhenTrusted(t *testing.T) {
	fn := IPKeyFunc(true)
	req := newRequest("10.0.0.1:1234", "", "203.0.113.99")
	got := fn(req)
	if got != "203.0.113.99" {
		t.Errorf("expected 203.0.113.99, got %s", got)
	}
}

func TestIPKeyFunc_FallsBackToRemoteAddrOnInvalidProxyIP(t *testing.T) {
	fn := IPKeyFunc(true)
	req := newRequest("10.0.0.2:5678", "not-an-ip", "also-bad")
	got := fn(req)
	if got != "10.0.0.2" {
		t.Errorf("expected 10.0.0.2, got %s", got)
	}
}

func TestRouteKeyFunc_CombinesIPAndPath(t *testing.T) {
	fn := RouteKeyFunc(false)
	req := newRequest("172.16.0.5:9000", "", "")
	got := fn(req)
	expected := "172.16.0.5:/api/resource"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
