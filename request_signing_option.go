package ratelimiter

import "net/http"

// NewRequestSigningConfig constructs a SigningConfig from the provided option functions
// and returns the configured middleware.
//
// Example:
//
//	mw := NewRequestSigningConfig(
//		WithSigningSecret([]byte("my-secret")),
//		WithSigningHeader("X-Hub-Signature-256"),
//	)
func NewRequestSigningConfig(opts ...func(*SigningConfig)) func(http.Handler) http.Handler {
	return WithRequestSigning(opts...)
}

// SigningByBodyHash returns a key function that signs the raw URL + a provided
// body hash header value, suitable for webhook-style validation where the caller
// includes a hash of the request body in a separate header.
func SigningByBodyHash(bodyHashHeader string) func(*http.Request) string {
	return func(r *http.Request) string {
		bodyHash := r.Header.Get(bodyHashHeader)
		return r.Method + ":" + r.URL.RequestURI() + ":" + bodyHash
	}
}

// SigningByAPIKey returns a key function that incorporates an API key header
// into the signed payload, binding the signature to a specific caller identity.
func SigningByAPIKey(apiKeyHeader string) func(*http.Request) string {
	return func(r *http.Request) string {
		apiKey := r.Header.Get(apiKeyHeader)
		return r.Method + ":" + r.URL.RequestURI() + ":" + apiKey
	}
}
