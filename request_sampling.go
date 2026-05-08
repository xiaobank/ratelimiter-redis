package ratelimiter

import (
	"math/rand"
	"net/http"
)

// SamplingKeyFunc determines a key from a request used for sampling decisions.
type SamplingKeyFunc func(r *http.Request) string

// SamplingConfig holds configuration for the request sampling middleware.
type SamplingConfig struct {
	// Rate is the fraction of requests to allow through (0.0 to 1.0).
	Rate float64
	// KeyFunc optionally partitions sampling by key (e.g. per-route).
	// If nil, sampling is applied globally.
	KeyFunc SamplingKeyFunc
	// RejectedHandler is called for sampled-out requests.
	// Defaults to a 429 Too Many Requests response.
	RejectedHandler http.Handler
	// RandSource allows injecting a deterministic source for testing.
	RandSource rand.Source
}

func defaultSamplingRejectedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
}

// WithRequestSampling returns middleware that only forwards a fraction of
// incoming requests to the next handler. Requests that are sampled out receive
// a 429 response (or a custom RejectedHandler).
//
// A Rate of 1.0 allows all requests; 0.0 blocks all requests.
func WithRequestSampling(cfg SamplingConfig) func(http.Handler) http.Handler {
	if cfg.Rate < 0 {
		cfg.Rate = 0
	}
	if cfg.Rate > 1 {
		cfg.Rate = 1
	}

	rejected := cfg.RejectedHandler
	if rejected == nil {
		rejected = http.HandlerFunc(defaultSamplingRejectedHandler)
	}

	var rng *rand.Rand
	if cfg.RandSource != nil {
		rng = rand.New(cfg.RandSource)
	} else {
		rng = rand.New(rand.NewSource(rand.Int63())) //nolint:gosec
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rng.Float64() < cfg.Rate {
				next.ServeHTTP(w, r)
				return
			}
			rejected.ServeHTTP(w, r)
		})
	}
}
