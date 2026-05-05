# ratelimiter-redis

Pluggable Redis-backed rate limiter middleware for Go HTTP services.

## Installation

```bash
go get github.com/yourusername/ratelimiter-redis
```

## Usage

```go
package main

import (
    "net/http"

    ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func main() {
    // Configure the rate limiter
    limiter := ratelimiter.New(ratelimiter.Options{
        RedisAddr:  "localhost:6379",
        Limit:      100,          // max requests
        Window:     time.Minute,  // per time window
        KeyFunc:    ratelimiter.IPKeyFunc, // identify clients by IP
    })

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Wrap your handler with the rate limiter middleware
    http.ListenAndServe(":8080", limiter.Middleware(mux))
}
```

When a client exceeds the configured limit, the middleware automatically responds with `429 Too Many Requests` and sets the appropriate `Retry-After` header.

## Configuration

| Option      | Type            | Description                              |
|-------------|-----------------|------------------------------------------|
| `RedisAddr` | `string`        | Redis server address (`host:port`)       |
| `Limit`     | `int`           | Maximum number of requests per window    |
| `Window`    | `time.Duration` | Duration of the rate limit window        |
| `KeyFunc`   | `KeyFunc`       | Function to extract a key from a request |

## Features

- Sliding window rate limiting using Redis sorted sets
- Custom key functions (by IP, API key, user ID, etc.)
- Zero external dependencies beyond `go-redis`
- Thread-safe and horizontally scalable

## License

MIT © [yourusername](https://github.com/yourusername)