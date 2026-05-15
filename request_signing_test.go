package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func signingOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestSigning_AllowsValidSignature(t *testing.T) {
	secret := []byte("test-secret")
	mw := WithRequestSigning(WithSigningSecret(secret))
	handler := mw(http.HandlerFunc(signingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	payload := req.Method + ":" + req.URL.RequestURI()
	sig := ComputeSignature(secret, payload)
	req.Header.Set("X-Signature", sig)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestSigning_BlocksMissingSignature(t *testing.T) {
	mw := WithRequestSigning(WithSigningSecret([]byte("secret")))
	handler := mw(http.HandlerFunc(signingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestWithRequestSigning_BlocksInvalidSignature(t *testing.T) {
	mw := WithRequestSigning(WithSigningSecret([]byte("secret")))
	handler := mw(http.HandlerFunc(signingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.Header.Set("X-Signature", "deadbeef")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestWithRequestSigning_CustomHeader(t *testing.T) {
	secret := []byte("secret")
	mw := WithRequestSigning(
		WithSigningSecret(secret),
		WithSigningHeader("X-Hub-Signature"),
	)
	handler := mw(http.HandlerFunc(signingOKHandler))

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	payload := req.Method + ":" + req.URL.RequestURI()
	req.Header.Set("X-Hub-Signature", ComputeSignature(secret, payload))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestSigning_CustomInvalidHandler(t *testing.T) {
	called := false
	mw := WithRequestSigning(
		WithSigningSecret([]byte("secret")),
		WithSigningInvalidHandler(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusUnauthorized)
		}),
	)
	handler := mw(http.HandlerFunc(signingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected custom invalid handler to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestWithRequestSigning_CustomKeyFunc(t *testing.T) {
	secret := []byte("secret")
	keyFn := func(r *http.Request) string { return "static-payload" }
	mw := WithRequestSigning(
		WithSigningSecret(secret),
		WithSigningKeyFunc(keyFn),
	)
	handler := mw(http.HandlerFunc(signingOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/any", nil)
	req.Header.Set("X-Signature", ComputeSignature(secret, "static-payload"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
