package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimiterMiddleware returns a middleware that limits request per second
// for each route.
func RateLimiterMiddleware(requestsLimit int, windowLength time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitByIP(requestsLimit, windowLength)
}
