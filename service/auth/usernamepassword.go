package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuongaz/go-saas/pkg/errors/apierror"
	model2 "github.com/tuongaz/go-saas/service/auth/model"
	"github.com/tuongaz/go-saas/service/auth/store"
	store2 "github.com/tuongaz/go-saas/store"
	"golang.org/x/crypto/bcrypt"
)

type SignupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Service) signupUsernamePasswordAccount(
	ctx context.Context,
	input *SignupInput,
) (*model2.AuthenticatedInfo, error) {
	found, err := s.store.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if found {
		return nil, apierror.NewValidationError("Email already exists", nil, nil)
	}

	hashedPw, err := s.hashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	firstName, lastName := splitName(input.Name)

	ownerAcc, org, _, accountRole, err := s.store.CreateOwnerAccount(ctx, store.CreateOwnerAccountInput{
		Email:     input.Email,
		Name:      input.Name,
		FirstName: firstName,
		LastName:  lastName,
		Provider:  model2.AuthProviderUsernamePassword,
		Password:  hashedPw,
	})
	if err != nil {
		return nil, fmt.Errorf("create owner account: %w", err)
	}

	if _, err := s.CreateAuthToken(ctx, accountRole.ID); err != nil {
		return nil, fmt.Errorf("create auth token: %w", err)
	}

	if err := s.OnAccountCreated().Trigger(ctx, &OnAccountCreatedEvent{
		AccountID:      ownerAcc.ID,
		OrganisationID: org.ID,
	}); err != nil {
		return nil, fmt.Errorf("notify account created: %w", err)
	}

	out, err := s.getAuthenticatedInfo(ctx, accountRole)
	if err != nil {
		return nil, fmt.Errorf("get authenticated info: %w", err)
	}

	return out, nil
}

func (s *Service) loginUsernamePasswordAccount(
	ctx context.Context,
	input *LoginInput,
) (*model2.AuthenticatedInfo, error) {
	user, err := s.store.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if store2.IsNotFoundError(err) {
			return nil, apierror.NewUnauthorizedErr("invalid credentials", nil)
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !s.isPasswordMatched(input.Password, user.Password) {
		return nil, apierror.NewUnauthorizedErr("invalid credentials", nil)
	}

	acc, err := s.store.GetAccountByAuthProvider(ctx, model2.AuthProviderUsernamePassword, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get account by auth provider: %w", err)
	}

	org, err := s.store.GetOrganisationByAccountIDAndRole(ctx, acc.ID, string(model2.RoleOwner))
	if err != nil {
		return nil, fmt.Errorf("get default owner account by provider: %w", err)
	}

	accountRole, err := s.store.GetAccountRole(ctx, org.ID, acc.ID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return s.getAuthenticatedInfo(ctx, accountRole)
}

func (s *Service) hashPassword(password string) (string, error) {
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}

	return string(hashedPw), nil
}

func (s *Service) isPasswordMatched(password, hashedPassword string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false
	}

	return true
}

func splitName(name string) (string, string) {
	var firstName, lastName string
	if name != "" {
		names := strings.Split(name, " ")
		firstName = names[0]
		lastName = names[len(names)-1]
	}

	return firstName, lastName
}
