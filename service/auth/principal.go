package auth

import (
	"context"
	"net/http"

	"github.com/autopus/bootstrap/model"
	"github.com/autopus/bootstrap/pkg/log"
)

var (
	principalKey = "principal"

	HeaderOrganisationID = "X-Organisation-ID"
)

func PrincipalToCtx(ctx context.Context, principal model.Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

func PrincipalFromCtx(ctx context.Context) model.Principal {
	p, ok := ctx.Value(principalKey).(model.Principal)
	if !ok {
		log.Default().ErrorContext(ctx, "principal not found in context")
		panic("principal not found in context")
	}

	return p
}

func AccountID(ctx context.Context) string {
	return PrincipalFromCtx(ctx).AccountID
}
func OrganisationID(ctx context.Context) string {
	return PrincipalFromCtx(ctx).OrganisationID
}
func GetOrganisationIDFromRequest(r *http.Request) string {
	return r.Header.Get(HeaderOrganisationID)
}
