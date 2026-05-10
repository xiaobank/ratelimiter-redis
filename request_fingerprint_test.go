package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func fingerprintOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestFingerprintConfig_GeneratesConsistentKey(t *testing.T) {
	cfg := DefaultFingerprintConfig()
	keyFn := cfg.KeyFunc()

	r1, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
	r1.RemoteAddr = "10.0.0.1:1234"

	r2, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
	r2.RemoteAddr = "10.0.0.1:5678"

	if keyFn(r1) != keyFn(r2) {
		t.Error("expected same fingerprint for same IP and path regardless of port")
	}
}

func TestFingerprintConfig_DifferentPathsProduceDifferentKeys(t *testing.T) {
	cfg := DefaultFingerprintConfig()
	keyFn := cfg.KeyFunc()

	r1, _ := http.NewRequest(http.MethodGet, "/api/foo", nil)
	r1.RemoteAddr = "10.0.0.1:1234"

	r2, _ := http.NewRequest(http.MethodGet, "/api/bar", nil)
	r2.RemoteAddr = "10.0.0.1:1234"

	if keyFn(r1) == keyFn(r2) {
		t.Error("expected different fingerprints for different paths")
	}
}

func TestFingerprintConfig_HeadersIncludedInKey(t *testing.T) {
	cfg := NewFingerprintConfig(
		WithFingerprintIP(false),
		WithFingerprintPath(false),
		WithFingerprintHeaders("User-Agent"),
	)
	keyFn := cfg.KeyFunc()

	r1, _ := http.NewRequest(http.MethodGet, "/", nil)
	r1.Header.Set("User-Agent", "Mozilla/5.0")

	r2, _ := http.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("User-Agent", "curl/7.0")

	if keyFn(r1) == keyFn(r2) {
		t.Error("expected different fingerprints for different User-Agent headers")
	}
}

func TestWithRequestFingerprint_SetsHeader(t *testing.T) {
	cfg := DefaultFingerprintConfig()
	middleware := WithRequestFingerprint(cfg)
	handler := middleware(http.HandlerFunc(fingerprintOKHandler))

	req, _ := http.NewRequest(http.MethodGet, "/api/resource", nil)
	req.RemoteAddr = "192.168.1.1:9999"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-Fingerprint") == "" {
		t.Error("expected X-Request-Fingerprint header to be set")
	}
}

func TestWithRequestFingerprint_PanicsOnNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil FingerprintConfig")
		}
	}()
	WithRequestFingerprint(nil)
}

func TestFingerprintByToken_IgnoresIPAndPath(t *testing.T) {
	cfg := FingerprintByToken()
	keyFn := cfg.KeyFunc()

	r1, _ := http.NewRequest(http.MethodGet, "/api/foo", nil)
	r1.RemoteAddr = "10.0.0.1:1234"
	r1.Header.Set("Authorization", "Bearer token123")

	r2, _ := http.NewRequest(http.MethodGet, "/api/bar", nil)
	r2.RemoteAddr = "10.0.0.2:5678"
	r2.Header.Set("Authorization", "Bearer token123")

	if keyFn(r1) != keyFn(r2) {
		t.Error("expected same fingerprint for same token regardless of IP/path")
	}
}
