package auth

import (
	"context"
	"fmt"

	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/service/auth/store"
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
) (*model.AuthenticatedInfo, error) {
	hashedPw, err := s.hashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	ownerAcc, org, _, accountRole, err := s.store.CreateOwnerAccount(ctx, store.CreateOwnerAccountInput{
		Email:    input.Email,
		Name:     input.Name,
		Provider: model.AuthProviderUsernamePassword,
		Password: hashedPw,
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
) (*model.AuthenticatedInfo, error) {
	user, err := s.store.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !s.isPasswordMatched(input.Password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	acc, org, err := s.store.GetDefaultOwnerAccountByProvider(ctx, model.AuthProviderUsernamePassword, user.ID)
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