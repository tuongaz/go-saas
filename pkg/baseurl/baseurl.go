package baseurl

import (
	"context"
	"net/http"

	"github.com/autopus/bootstrap/config"
)

type ctxString string

const (
	baseURLKey ctxString = "base_url"
)

// NewMiddleware creates a new middleware that sets the base URL in the request context.
func NewMiddleware(cfg config.Interface) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, baseURLKey, getURL(r))

			next.ServeHTTP(w, r)
		})
	}
}

func Get(ctx context.Context) string {
	v := ctx.Value(baseURLKey)
	if url, ok := v.(string); ok {
		return url
	}

	return ""
}

func getURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = r.URL.Scheme
		if scheme == "" {
			scheme = "https"
		}
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}

	return scheme + "://" + host
}
