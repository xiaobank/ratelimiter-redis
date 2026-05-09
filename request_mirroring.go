package ratelimiter

import (
	"bytes"
	"io"
	"net/http"
)

// MirrorKeyFunc determines whether a given request should be mirrored.
type MirrorKeyFunc func(r *http.Request) bool

// MirrorConfig holds configuration for the request mirroring middleware.
type MirrorConfig struct {
	// ShouldMirror determines if the request should be mirrored. Defaults to mirroring all requests.
	ShouldMirror MirrorKeyFunc
	// MirrorTarget is the base URL to mirror requests to (e.g. "http://shadow-backend").
	MirrorTarget string
	// Client is the HTTP client used to send mirror requests. Defaults to http.DefaultClient.
	Client *http.Client
}

// WithRequestMirroring returns middleware that asynchronously mirrors incoming
// requests to a shadow backend. The primary response is unaffected.
func WithRequestMirroring(cfg MirrorConfig) func(http.Handler) http.Handler {
	if cfg.MirrorTarget == "" {
		panic("ratelimiter: WithRequestMirroring requires a non-empty MirrorTarget")
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	if cfg.ShouldMirror == nil {
		cfg.ShouldMirror = func(r *http.Request) bool { return true }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.ShouldMirror(r) {
				go mirrorRequest(cfg, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func mirrorRequest(cfg MirrorConfig, r *http.Request) {
	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			return
		}
	}

	targetURL := cfg.MirrorTarget + r.RequestURI
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}

	for key, vals := range r.Header {
		for _, v := range vals {
			req.Header.Add(key, v)
		}
	}
	req.Header.Set("X-Mirrored-Request", "true")

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
}
