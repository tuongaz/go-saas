package auth

import (
	"net/http"
	"strings"

	"github.com/autopus/bootstrap/model"
	"github.com/autopus/bootstrap/pkg/errors"
	"github.com/autopus/bootstrap/pkg/httputil"
)

func (s *Service) NewMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			claims, err := s.signer.ParseCustomClaims(strings.Replace(bearer, "Bearer ", "", 1))
			if err != nil {
				httputil.HandleResponse(r.Context(), w, nil, errors.NewUnauthorizedErr(err))
				return
			}

			accRole, err := s.GetAccountRole(r.Context(), claims.Organisation, claims.Subject)
			if err != nil {
				httputil.HandleResponse(r.Context(), w, nil, errors.NewUnauthorizedErr(err))
				return
			}

			ctx := PrincipalToCtx(r.Context(), model.Principal{
				OrganisationID: claims.Organisation,
				AccountID:      claims.Subject,
				AccountType:    claims.AccountType,
				Role:           model.Role(accRole.Role),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
