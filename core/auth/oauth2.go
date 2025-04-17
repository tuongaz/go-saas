package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tuongaz/go-saas/config"
	model2 "github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/core/auth/store"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/pkg/oauth2"
	store2 "github.com/tuongaz/go-saas/store"
)

// oauth2Authenticate creates new account, with new organisation and assign owner role to the account
func (s *service) oauth2Authenticate(
	ctx context.Context,
	user oauth2.User,
) (*model2.AuthenticatedInfo, error) {
	var ownerAcc *model2.Account
	var org *model2.Organisation
	var err error
	var newAccount bool

	ownerAcc, err = s.store.GetAccountByLoginProvider(ctx, user.Provider, user.UserID)
	if err != nil && !store2.IsNotFoundError(err) {
		return nil, fmt.Errorf("get account by auth provider: %w", err)
	}

	if ownerAcc == nil { // new user
		newAccount = true
		org, ownerAcc, err = s.oauth2SignupNewAccount(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	org, err = s.store.GetOrganisationByAccountIDAndRole(ctx, ownerAcc.ID, string(model2.RoleOwner))
	if err != nil {
		return nil, fmt.Errorf("get default owner account by provider: %w", err)
	}

	if newAccount {
		if err := s.OnAccountCreated().Trigger(ctx, &OnAccountCreatedEvent{
			AccountID:      ownerAcc.ID,
			OrganisationID: org.ID,
		}); err != nil {
			return nil, fmt.Errorf("trigger on account created: %w", err)
		}
	}

	accountRole, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, org.ID, ownerAcc.ID)
	if err != nil {
		return nil, err
	}

	return s.getAuthenticatedInfo(ctx, accountRole, user.UserID, DeviceFromCtx(ctx))
}

func (s *service) oauth2SignupLogin(w http.ResponseWriter, r *http.Request, oauthProvider config.OAuth2ProviderConfig, user oauth2.User) {
	ctx := r.Context()

	authInfo, err := s.oauth2Authenticate(
		ctx,
		user,
	)
	if err != nil {
		log.Default().ErrorContext(ctx, "failed to signup or login", log.ErrorAttr(err))
		http.Redirect(w, r, oauthProvider.FailureURL, http.StatusFound)
		return
	}

	redirectURL := fmt.Sprintf(
		"%s?token=%s&refresh_token=%s",
		oauthProvider.SuccessURL,
		authInfo.Token,
		authInfo.RefreshToken,
	)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (s *service) oauth2SignupNewAccount(
	ctx context.Context,
	user oauth2.User,
) (*model2.Organisation, *model2.Account, error) {
	acc, accountOrg, loginProvider, accountRole, err := s.store.CreateOwnerAccount(ctx, store.CreateOwnerAccountInput{
		Name:           user.Name,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		Provider:       user.Provider,
		ProviderUserID: user.UserID,
		Avatar:         user.AvatarURL,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create owner account: %w", err)
	}

	if _, err := s.CreateAccessToken(ctx, accountRole.ID, loginProvider.ProviderUserID, DeviceFromCtx(ctx)); err != nil {
		return nil, nil, fmt.Errorf("create auth token: %w", err)
	}

	return accountOrg, acc, nil
}
