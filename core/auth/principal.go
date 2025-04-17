package auth

import (
	"context"

	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/pkg/log"
)

type PrincipalKey string

var (
	principalKey PrincipalKey = "principal"
)

func PrincipalToCtx(ctx context.Context, principal model.Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

func PrincipalFromCtx(ctx context.Context) model.Principal {
	p, ok := ctx.Value(principalKey).(model.Principal)
	if !ok {
		log.Default().ErrorContext(ctx, "principal not found in context")
		return model.Principal{
			Role: model.Role(""),
		}
	}

	return p
}

func AccountID(ctx context.Context) string {
	return PrincipalFromCtx(ctx).AccountID
}
func OrganisationID(ctx context.Context) string {
	return PrincipalFromCtx(ctx).OrganisationID
}
