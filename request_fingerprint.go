package ratelimiter

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

// FingerprintConfig holds configuration for request fingerprinting.
type FingerprintConfig struct {
	headers    []string
	includeIP  bool
	includePath bool
	keyFunc    func(*http.Request) string
}

// FingerprintOption configures a FingerprintConfig.
type FingerprintOption func(*FingerprintConfig)

// WithFingerprintHeaders includes the given request headers in the fingerprint.
func WithFingerprintHeaders(headers ...string) FingerprintOption {
	return func(c *FingerprintConfig) {
		c.headers = append(c.headers, headers...)
	}
}

// WithFingerprintIP includes the client IP in the fingerprint.
func WithFingerprintIP(include bool) FingerprintOption {
	return func(c *FingerprintConfig) {
		c.includeIP = include
	}
}

// WithFingerprintPath includes the request path in the fingerprint.
func WithFingerprintPath(include bool) FingerprintOption {
	return func(c *FingerprintConfig) {
		c.includePath = include
	}
}

// NewFingerprintConfig creates a FingerprintConfig with the given options.
func NewFingerprintConfig(opts ...FingerprintOption) *FingerprintConfig {
	cfg := &FingerprintConfig{
		includeIP:   true,
		includePath: true,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// KeyFunc returns a KeyFunc that generates a fingerprint-based key for requests.
func (c *FingerprintConfig) KeyFunc() func(*http.Request) string {
	return func(r *http.Request) string {
		var parts []string
		if c.includeIP {
			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
			parts = append(parts, "ip:"+ip)
		}
		if c.includePath {
			parts = append(parts, "path:"+r.URL.Path)
		}
		for _, h := range c.headers {
			v := r.Header.Get(h)
			parts = append(parts, h+":"+v)
		}
		raw := strings.Join(parts, "|")
		sum := sha256.Sum256([]byte(raw))
		return fmt.Sprintf("%x", sum[:8])
	}
}

// WithRequestFingerprint returns middleware that sets X-Request-Fingerprint
// header on each request using the provided FingerprintConfig.
func WithRequestFingerprint(cfg *FingerprintConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		panic("ratelimiter: FingerprintConfig must not be nil")
	}
	keyFn := cfg.KeyFunc()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fingerprint := keyFn(r)
			r.Header.Set("X-Request-Fingerprint", fingerprint)
			w.Header().Set("X-Request-Fingerprint", fingerprint)
			next.ServeHTTP(w, r)
		})
	}
}
