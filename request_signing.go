package ratelimiter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// SigningConfig holds configuration for request signature validation middleware.
type SigningConfig struct {
	secret          []byte
	headerName      string
	keyFunc         func(*http.Request) string
	onInvalid       http.HandlerFunc
}

// WithSigningSecret sets the HMAC secret used to validate signatures.
func WithSigningSecret(secret []byte) func(*SigningConfig) {
	return func(c *SigningConfig) {
		c.secret = secret
	}
}

// WithSigningHeader sets the request header that carries the signature.
func WithSigningHeader(header string) func(*SigningConfig) {
	return func(c *SigningConfig) {
		c.headerName = header
	}
}

// WithSigningKeyFunc sets the function that derives the payload to sign from the request.
func WithSigningKeyFunc(fn func(*http.Request) string) func(*SigningConfig) {
	return func(c *SigningConfig) {
		c.keyFunc = fn
	}
}

// WithSigningInvalidHandler sets the handler invoked when a signature is missing or invalid.
func WithSigningInvalidHandler(h http.HandlerFunc) func(*SigningConfig) {
	return func(c *SigningConfig) {
		c.onInvalid = h
	}
}

func defaultSigningKeyFunc(r *http.Request) string {
	return r.Method + ":" + r.URL.RequestURI()
}

func defaultSigningInvalidHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Forbidden: invalid or missing request signature", http.StatusForbidden)
}

// ComputeSignature returns the HMAC-SHA256 hex signature for the given payload and secret.
func ComputeSignature(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// WithRequestSigning returns middleware that validates an HMAC-SHA256 request signature.
// Requests with a missing or invalid signature are rejected with 403 Forbidden.
func WithRequestSigning(opts ...func(*SigningConfig)) func(http.Handler) http.Handler {
	cfg := &SigningConfig{
		headerName: "X-Signature",
		keyFunc:    defaultSigningKeyFunc,
		onInvalid:  defaultSigningInvalidHandler,
	}
	for _, o := range opts {
		o(cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sig := strings.TrimSpace(r.Header.Get(cfg.headerName))
			if sig == "" {
				cfg.onInvalid(w, r)
				return
			}
			payload := cfg.keyFunc(r)
			expected := ComputeSignature(cfg.secret, payload)
			if !hmac.Equal([]byte(sig), []byte(expected)) {
				cfg.onInvalid(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
