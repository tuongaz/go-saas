package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/autopus/bootstrap/model"
)

func TestPrincipalToCtxAndFromCtx(t *testing.T) {
	principal := model.Principal{
		AccountID:      "acc123",
		OrganisationID: "org123",
	}

	ctx := context.Background()
	ctx = PrincipalToCtx(ctx, principal)
	retrievedPrincipal := PrincipalFromCtx(ctx)

	assert.Equal(t, principal, retrievedPrincipal)
}

func TestPrincipalFromCtxWithoutPrincipal(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		PrincipalFromCtx(ctx)
	})
}

func TestAccountIDProjectIDOrganisationID(t *testing.T) {
	principal := model.Principal{
		AccountID:      "acc123",
		OrganisationID: "org123",
	}

	ctx := context.Background()
	ctx = PrincipalToCtx(ctx, principal)

	assert.Equal(t, "acc123", AccountID(ctx))
	assert.Equal(t, "org123", OrganisationID(ctx))
}

func TestGetOrganisationIDFromRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderOrganisationID, "org123")

	organisationID := GetOrganisationIDFromRequest(req)
	assert.Equal(t, "org123", organisationID)
}
