package auth

import (
	"context"
	"fmt"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/google/uuid"
)

func (s *Service) Token(
	ctx context.Context,
	data *oauth2.GenerateBasic,
	isGenRefresh bool,
) (string, string, error) {
	accRoles, err := s.store.GetAccountRoles(ctx, data.UserID)
	if err != nil {
		return "", "", fmt.Errorf("get account roles: %w", err)
	}

	if len(accRoles) == 0 {
		return "", "", fmt.Errorf("account has no roles")
	}

	// Currently we only support 1 organisation per account
	accRole := accRoles[0]

	if orgID := GetOrganisationIDFromRequest(data.Request); orgID != "" {
		for _, r := range accRoles {
			if r.OrganisationID == orgID {
				accRole = r
				break
			}
		}
	}

	authToken, err := s.store.GetAuthTokenByAccountRoleID(ctx, accRole.ID)
	if err != nil {
		return "", "", fmt.Errorf("get auth token by account id: %w", err)
	}

	authInfo, err := s.newAuthenticatedInfo(accRole, authToken)
	if err != nil {
		return "", "", fmt.Errorf("new authenticated info: %w", err)
	}

	return authInfo.Token, uuid.NewString(), nil
}
