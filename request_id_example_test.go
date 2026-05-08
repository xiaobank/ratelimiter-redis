package ratelimiter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

// ExampleWithRequestID demonstrates attaching unique request IDs to every request.
func ExampleWithRequestID() {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The request ID is available on both the request and response headers.
		id := r.Header.Get(ratelimiter.DefaultRequestIDHeader)
		fmt.Fprintln(w, "id:", id)
	})

	handler := ratelimiter.WithRequestID(
		ratelimiter.WithRequestIDGenerator(func() string { return "demo-id" }),
	)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())
	// Output: id: demo-id
}
