package ratelimiter

import (
	"net/http"
	"strings"
)

// TagKeyFunc extracts a tag value from a request.
type TagKeyFunc func(r *http.Request) string

// TagConfig holds configuration for request tagging middleware.
type TagConfig struct {
	keyFunc    TagKeyFunc
	headerName string
	prefix     string
}

// TagOption configures a TagConfig.
type TagOption func(*TagConfig)

// WithTagHeader sets the response header name used to echo the tag.
func WithTagHeader(name string) TagOption {
	return func(c *TagConfig) {
		c.headerName = name
	}
}

// WithTagPrefix prepends a fixed string to every tag value.
func WithTagPrefix(prefix string) TagOption {
	return func(c *TagConfig) {
		c.prefix = prefix
	}
}

// WithTagKeyFunc sets the function used to derive a tag from the request.
func WithTagKeyFunc(fn TagKeyFunc) TagOption {
	return func(c *TagConfig) {
		c.keyFunc = fn
	}
}

// defaultTagKeyFunc returns the value of the X-Request-Tag header.
func defaultTagKeyFunc(r *http.Request) string {
	return r.Header.Get("X-Request-Tag")
}

// NewTagConfig builds a TagConfig with optional overrides.
func NewTagConfig(opts ...TagOption) *TagConfig {
	cfg := &TagConfig{
		keyFunc:    defaultTagKeyFunc,
		headerName: "X-Tag",
		prefix:     "",
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// WithRequestTagging is middleware that reads a tag from the request and
// echoes it (optionally prefixed) in the response headers.
func WithRequestTagging(cfg *TagConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		panic("ratelimiter: WithRequestTagging requires a non-nil TagConfig")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tag := cfg.keyFunc(r)
			if tag != "" {
				value := tag
				if cfg.prefix != "" {
					value = strings.TrimRight(cfg.prefix, "-") + "-" + tag
				}
				w.Header().Set(cfg.headerName, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}
