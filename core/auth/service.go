package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/core/auth/signer"
	"github.com/tuongaz/go-saas/core/auth/store"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/pkg/log"
	store2 "github.com/tuongaz/go-saas/store"
)

type Service struct {
	store            store.Interface
	signer           signer.Interface
	encryptor        encrypt.Interface
	tokenLifeTime    time.Duration
	jwtIssuer        string
	redirectURL      string
	providers        map[string]config.OAuth2ProviderConfig
	onAccountCreated *hooks.Hook[*OnAccountCreatedEvent]
}

func New(cfg *config.Config, st store2.Interface) (*Service, error) {
	authStore, err := store.New(st)
	if err != nil {
		return nil, err
	}

	authSrv := &Service{
		signer:           signer.NewHS256Signer([]byte(cfg.JWTSigningSecret)),
		encryptor:        encrypt.New(cfg.EncryptionKey),
		jwtIssuer:        cfg.JWTIssuer,
		tokenLifeTime:    time.Duration(cfg.JWTTokenLifetimeSeconds) * time.Second,
		providers:        cfg.Oauth2AuthProviders,
		onAccountCreated: &hooks.Hook[*OnAccountCreatedEvent]{},
		store:            authStore,
	}

	return authSrv, nil
}

func (s *Service) Store() store.Interface {
	return s.store
}

func (s *Service) GetAccount(ctx context.Context, accountID string) (*model.Account, error) {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account by id: %w", err)
	}

	return account, nil
}

func (s *Service) GetAccountRole(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error) {
	accRole, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, organisationID, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return accRole, nil
}

func (s *Service) GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, error) {
	authToken, err := s.store.GetAccessTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("get auth token by refresh token: %w", err)
	}

	return authToken, nil
}

func (s *Service) CreateAccessToken(ctx context.Context, accountRoleID, providerUserID, device string) (*model.AccessToken, error) {
	refreshToken := uuid.New().String()
	if _, err := s.store.CreateAccessToken(ctx, store.CreateAccessTokenInput{
		AccountRoleID:  accountRoleID,
		RefreshToken:   refreshToken,
		Device:         device,
		ProviderUserID: providerUserID,
	}); err != nil {
		return nil, fmt.Errorf("create auth token: %w", err)
	}

	return &model.AccessToken{
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthenticatedInfo, error) {
	accessToken, err := s.store.GetAccessTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	accountRole, err := s.store.GetAccountRoleByID(ctx, accessToken.AccountRoleID)
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.New().String()
	newAccessToken, err := s.store.CreateAccessToken(ctx, store.CreateAccessTokenInput{
		AccountRoleID:  accountRole.ID,
		RefreshToken:   newRefreshToken,
		Device:         DeviceFromCtx(ctx),
		ProviderUserID: accessToken.ProviderUserID,
	})
	if err != nil {
		return nil, err
	}

	info, err := s.newAuthenticatedInfo(accountRole, newAccessToken)
	if err != nil {
		return nil, err
	}

	log.Info("refresh token", info.RefreshToken, err)

	return info, err
}

func (s *Service) newAuthenticatedInfo(
	accountRole *model.AccountRole,
	authToken *model.AccessToken,
) (*model.AuthenticatedInfo, error) {
	claims := model.CustomClaims{
		Organisation: accountRole.OrganisationID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  s.jwtIssuer,
			Subject: accountRole.AccountID,
			Audience: jwt.ClaimStrings{
				s.jwtIssuer,
			},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenLifeTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	jwtToken, err := s.signer.SignCustomClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}

	return &model.AuthenticatedInfo{
		RefreshToken: authToken.RefreshToken,
		Type:         "Bearer",
		Token:        jwtToken,
		ExpiresIn:    int64(s.tokenLifeTime.Seconds()),
	}, nil
}

func (s *Service) getAuthenticatedInfo(
	ctx context.Context,
	accountRole *model.AccountRole,
	providerUserID string,
	device string,
) (*model.AuthenticatedInfo, error) {
	accessToken, err := s.store.GetAccessToken(ctx, store.GetAccessTokenInput{
		AccountRoleID:  accountRole.ID,
		ProviderUserID: providerUserID,
		Device:         device,
	})
	if err != nil {
		return nil, fmt.Errorf("get auth token by account id: %w", err)
	}

	return s.newAuthenticatedInfo(accountRole, accessToken)
}
