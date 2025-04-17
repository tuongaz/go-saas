package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/httputil"
)

const (
	deviceKey = "device"
)

// NewMiddleware creates a new middleware that authenticates the user and sets the principal in the context.
func (s *service) NewMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			bearer := r.Header.Get("Authorization")
			if bearer == "" {
				httputil.HandleResponse(ctx, w, nil, apierror.NewUnauthorizedErr("missing authorization header", nil))
				return
			}

			tokenString := strings.Replace(bearer, "Bearer ", "", 1)
			if tokenString == bearer {
				// If no replacement was done, it means the format was wrong
				httputil.HandleResponse(ctx, w, nil, apierror.NewUnauthorizedErr("invalid authorization format, expected 'Bearer {token}'", nil))
				return
			}

			claims, err := s.signer.ParseCustomClaims(tokenString)
			if err != nil {
				var message string
				if errors.Is(err, jwt.ErrTokenExpired) {
					message = "token expired"
				} else {
					message = "invalid token"
				}
				httputil.HandleResponse(ctx, w, nil, apierror.NewUnauthorizedErr(message, err))
				return
			}

			accRole, err := s.GetAccountRole(ctx, claims.Organisation, claims.Subject)
			if err != nil {
				httputil.HandleResponse(ctx, w, nil, apierror.NewUnauthorizedErr("invalid credentials", err))
				return
			}

			ctx = PrincipalToCtx(ctx, model.Principal{
				OrganisationID: claims.Organisation,
				AccountID:      claims.Subject,
				Role:           model.Role(accRole.Role),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *service) NewDeviceMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			device := r.Header.Get("User-Agent")
			if device == "" {
				device = "unknown"
			}

			ctx := r.Context()
			ctx = deviceToCtx(ctx, device)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func deviceToCtx(ctx context.Context, device string) context.Context {
	return context.WithValue(ctx, deviceKey, device)
}

func DeviceFromCtx(ctx context.Context) string {
	return ctx.Value(deviceKey).(string)
}
