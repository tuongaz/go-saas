package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimiterMiddleware returns a middleware that limits request per second
// for each route. The default rate is 15 requests per second.
func RateLimiterMiddleware() func(http.Handler) http.Handler {
	return httprate.LimitByIP(15, time.Second)
}

// RateLimiterWithOptions returns a middleware with custom options
// for more advanced rate limiting scenarios
func RateLimiterWithOptions(requestsLimit int, windowLength time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitByIP(requestsLimit, windowLength)
}
