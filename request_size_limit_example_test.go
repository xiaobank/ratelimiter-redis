package ratelimiter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

// ExampleWithRequestSizeLimit demonstrates how to cap incoming request body
// size to 1 MB using the middleware.
//
// Requests whose Content-Length is within the limit are passed through to the
// wrapped handler unchanged. Requests that exceed the limit receive an HTTP
// 413 Request Entity Too Large response before the body is read.
func ExampleWithRequestSizeLimit() {
	const oneMB = 1 << 20 // 1 048 576 bytes

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "accepted")
	})

	handler := ratelimiter.WithRequestSizeLimit(oneMB)(base)

	// Simulate a small (allowed) request.
	smallBody := strings.NewReader("hello world")
	req := httptest.NewRequest(http.MethodPost, "/upload", smallBody)
	req.ContentLength = int64(smallBody.Len())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Println(rec.Code)

	// Simulate an oversized request (Content-Length exceeds the 1 MB cap).
	largeBody := strings.NewReader(strings.Repeat("x", oneMB+1))
	req2 := httptest.NewRequest(http.MethodPost, "/upload", largeBody)
	req2.ContentLength = int64(oneMB + 1)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	fmt.Println(rec2.Code)

	// Output:
	// 200
	// 413
}
