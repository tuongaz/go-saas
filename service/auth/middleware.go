package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/httputil"
	"github.com/tuongaz/go-saas/service/auth/model"
)

const (
	deviceKey = "device"
)

// NewMiddleware creates a new middleware that authenticates the user and sets the principal in the context.
func (s *Service) NewMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			bearer := r.Header.Get("Authorization")
			claims, err := s.signer.ParseCustomClaims(strings.Replace(bearer, "Bearer ", "", 1))
			if err != nil {
				httputil.HandleResponse(ctx, w, nil, apierror.NewUnauthorizedErr("invalid credentials", err))
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

func (s *Service) NewDeviceMiddleware() func(next http.Handler) http.Handler {
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
