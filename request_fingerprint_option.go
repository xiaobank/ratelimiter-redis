package ratelimiter

// DefaultFingerprintConfig returns a FingerprintConfig that includes the
// client IP and request path in the fingerprint.
func DefaultFingerprintConfig() *FingerprintConfig {
	return NewFingerprintConfig(
		WithFingerprintIP(true),
		WithFingerprintPath(true),
	)
}

// FingerprintByUserAgent returns a FingerprintConfig that includes the
// client IP, request path, and User-Agent header.
func FingerprintByUserAgent() *FingerprintConfig {
	return NewFingerprintConfig(
		WithFingerprintIP(true),
		WithFingerprintPath(true),
		WithFingerprintHeaders("User-Agent"),
	)
}

// FingerprintByToken returns a FingerprintConfig that includes the
// Authorization header only — useful for token-scoped rate limiting.
func FingerprintByToken() *FingerprintConfig {
	return NewFingerprintConfig(
		WithFingerprintIP(false),
		WithFingerprintPath(false),
		WithFingerprintHeaders("Authorization"),
	)
}
