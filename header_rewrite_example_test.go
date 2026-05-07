package ratelimiter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	ratelimiter "github.com/you/ratelimiter-redis"
)

// ExampleWithHeaderRewrite demonstrates composing header rewriting with the
// rate-limiter to emit IETF draft RateLimit headers instead of the X- prefixed
// variants.
func ExampleWithHeaderRewrite() {
	store := ratelimiter.NewMemoryStore()

	rl := ratelimiter.New(
		ratelimiter.WithStore(store),
		ratelimiter.WithLimit(10),
	)

	cfg := ratelimiter.NewHeaderRewriteConfig(
		ratelimiter.WithLimitHeader("RateLimit-Limit"),
		ratelimiter.WithRemainingHeader("RateLimit-Remaining"),
		ratelimiter.WithResetHeader("RateLimit-Reset"),
	)

	handler := ratelimiter.WithHeaderRewrite(cfg)(
		rl(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.RemoteAddr = "192.168.1.1:5000"
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	// Output: 200
}
