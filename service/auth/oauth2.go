package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/pkg/auth/oauth2"
	"github.com/tuongaz/go-saas/pkg/errors"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/service/auth/store"
)

// oauth2Authenticate creates new account, with new organisation and assign owner role to the account
func (s *Service) oauth2Authenticate(
	ctx context.Context,
	user oauth2.User,
) (*model.AuthenticatedInfo, error) {
	var ownerAcc *model.Account
	var org *model.Organisation
	var err error
	var newAccount bool

	ownerAcc, org, err = s.store.GetDefaultOwnerAccountByProvider(ctx, user.Provider, user.UserID)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("get account by provider: %w", err)
		}
	}

	if ownerAcc == nil { // new user
		newAccount = true
		org, ownerAcc, err = s.oauth2SignupNewAccount(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	if newAccount {
		if err := s.OnAccountCreated().Trigger(ctx, &OnAccountCreatedEvent{
			AccountID:      ownerAcc.ID,
			OrganisationID: org.ID,
		}); err != nil {
			return nil, fmt.Errorf("notify account created: %w", err)
		}
	}

	accountRole, err := s.store.GetAccountRole(ctx, org.ID, ownerAcc.ID)
	if err != nil {
		return nil, err
	}

	return s.getAuthenticatedInfo(ctx, accountRole)
}

func (s *Service) oauth2SignupLogin(w http.ResponseWriter, r *http.Request, oauthProvider OAuth2ProviderConfig, user oauth2.User) {
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

func (s *Service) oauth2SignupNewAccount(
	ctx context.Context,
	user oauth2.User,
) (*model.Organisation, *model.Account, error) {
	acc, accountOrg, _, accountRole, err := s.store.CreateOwnerAccount(ctx, store.CreateOwnerAccountInput{
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

	if _, err := s.CreateAuthToken(ctx, accountRole.ID); err != nil {
		return nil, nil, fmt.Errorf("create auth token: %w", err)
	}

	return accountOrg, acc, nil
}
