package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tuongaz/go-saas/app"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/hooks"
	model2 "github.com/tuongaz/go-saas/service/auth/model"
	"github.com/tuongaz/go-saas/service/auth/signer"
	"github.com/tuongaz/go-saas/service/auth/store"
)

type Service struct {
	app                  app.Interface
	store                store.Interface
	signer               signer.Interface
	encryptor            encrypt.Interface
	tokenLifeTimeMinutes time.Duration
	jwtIssuer            string
	redirectURL          string
	providers            map[string]OAuth2ProviderConfig
	onAccountCreated     *hooks.Hook[*OnAccountCreatedEvent]
}

type OAuth2ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	Scopes       []string
	RedirectURL  string
	FailureURL   string
	SuccessURL   string
}

type Config struct {
	JWTSigningSecret        string
	JWTIssuer               string
	JWTTokenLifetimeMinutes int
	Providers               []OAuth2ProviderConfig
}

func WithJWTTokenLifetimeMinutes(minutes int) func(*Config) {
	return func(cfg *Config) {
		cfg.JWTTokenLifetimeMinutes = minutes
	}
}

func WithJWTSigningSecret(secret string) func(*Config) {
	return func(cfg *Config) {
		cfg.JWTSigningSecret = secret
	}
}

func WithJWTIssuer(issuer string) func(*Config) {
	return func(cfg *Config) {
		cfg.JWTIssuer = issuer
	}
}

func WithOauth2Provider(providers ...OAuth2ProviderConfig) func(*Config) {
	return func(cfg *Config) {
		cfg.Providers = append(cfg.Providers, providers...)
	}
}

func Register(appInstance app.Interface, opts ...func(*Config)) *Service {
	cfg := &Config{
		JWTTokenLifetimeMinutes: 30,
		JWTSigningSecret:        "signing-secret-please-change-me",
		JWTIssuer:               "go-saas-issuer",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	providers := make(map[string]OAuth2ProviderConfig)
	for _, p := range cfg.Providers {
		providers[p.Name] = p
	}

	authSrv := &Service{
		app:                  appInstance,
		signer:               signer.NewHS256Signer([]byte(cfg.JWTSigningSecret)),
		encryptor:            encrypt.New(appInstance.Config().GetEncryptionKey()),
		jwtIssuer:            cfg.JWTIssuer,
		tokenLifeTimeMinutes: time.Duration(cfg.JWTTokenLifetimeMinutes) * time.Minute,
		providers:            providers,
		onAccountCreated:     &hooks.Hook[*OnAccountCreatedEvent]{},
	}

	appInstance.OnAfterBootstrap().Add(func(ctx context.Context, e *app.OnAfterBootstrapEvent) error {
		if err := authSrv.bootstrap(); err != nil {
			return fmt.Errorf("auth service bootstrap: %w", err)
		}
		return nil
	})

	return authSrv
}

func (s *Service) GetAccountRole(ctx context.Context, organisationID, accountID string) (*model2.AccountRole, error) {
	accRole, err := s.store.GetAccountRole(ctx, organisationID, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return accRole, nil
}

func (s *Service) GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model2.AuthToken, error) {
	authToken, err := s.store.GetAuthTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("get auth token by refresh token: %w", err)
	}

	return authToken, nil
}

func (s *Service) CreateAuthToken(ctx context.Context, accountRoleID string) (*model2.AuthToken, error) {
	refreshToken := uuid.New().String()
	if _, err := s.store.CreateAuthToken(ctx, store.CreateAuthTokenInput{
		AccountRoleID: accountRoleID,
		RefreshToken:  refreshToken,
	}); err != nil {
		return nil, fmt.Errorf("create auth token: %w", err)
	}

	return &model2.AuthToken{
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) NewToken(ctx context.Context, accountRole *model2.AccountRole) (*model2.AuthenticatedInfo, error) {
	authToken, err := s.CreateAuthToken(ctx, accountRole.ID)
	if err != nil {
		return nil, err
	}

	return s.newAuthenticatedInfo(accountRole, authToken)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model2.AuthenticatedInfo, error) {
	authToken, err := s.store.GetAuthTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	accountRole, err := s.store.GetAccountRoleByID(ctx, authToken.AccountRoleID)
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.New().String()
	if err := s.store.UpdateAuthToken(ctx, authToken.ID, store.UpdateAuthTokenInput{
		RefreshToken: newRefreshToken,
	}); err != nil {
		return nil, err
	}

	authToken.RefreshToken = newRefreshToken

	return s.newAuthenticatedInfo(accountRole, authToken)
}

func (s *Service) bootstrap() error {
	authStore, err := store.New(s.app.Store().DB())
	if err != nil {
		return fmt.Errorf("new auth store: %w", err)
	}
	s.store = authStore

	s.app.OnBeforeServe().Add(func(ctx context.Context, e *app.OnBeforeServeEvent) error {
		rootRouter := e.Server.Router()
		s.setupAPI(rootRouter)

		return nil
	})

	return nil
}

func (s *Service) newAuthenticatedInfo(
	accountRole *model2.AccountRole,
	authToken *model2.AuthToken,
) (*model2.AuthenticatedInfo, error) {
	claims := model2.CustomClaims{
		Organisation: accountRole.OrganisationID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  s.jwtIssuer,
			Subject: accountRole.AccountID,
			Audience: jwt.ClaimStrings{
				s.jwtIssuer,
			},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenLifeTimeMinutes)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	jwtToken, err := s.signer.SignCustomClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}

	return &model2.AuthenticatedInfo{
		RefreshToken: authToken.RefreshToken,
		Type:         "Bearer",
		Token:        jwtToken,
		ExpiresIn:    int64(s.tokenLifeTimeMinutes.Seconds()),
	}, nil
}

func (s *Service) getAuthenticatedInfo(ctx context.Context, accountRole *model2.AccountRole) (*model2.AuthenticatedInfo, error) {
	authToken, err := s.store.GetAuthTokenByAccountRoleID(ctx, accountRole.ID)
	if err != nil {
		return nil, fmt.Errorf("get auth token by account id: %w", err)
	}

	return s.newAuthenticatedInfo(accountRole, authToken)
}
