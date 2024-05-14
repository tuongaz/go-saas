package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tuongaz/go-saas/app"
	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/pkg/auth/signer"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/hooks"
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

func Register(appInstance app.Interface, cfg Config) *Service {
	if cfg.JWTTokenLifetimeMinutes == 0 {
		cfg.JWTTokenLifetimeMinutes = 30
	}

	if cfg.JWTSigningSecret == "" {
		cfg.JWTSigningSecret = "signing-secret-please-change-me"
	}

	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "go-saas-issuer"
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

func (s *Service) GetAccountRole(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error) {
	accRole, err := s.store.GetAccountRole(ctx, organisationID, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return accRole, nil
}

func (s *Service) GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AuthToken, error) {
	authToken, err := s.store.GetAuthTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("get auth token by refresh token: %w", err)
	}

	return authToken, nil
}

func (s *Service) CreateAuthToken(ctx context.Context, accountRoleID string) (*model.AuthToken, error) {
	refreshToken := uuid.New().String()
	if _, err := s.store.CreateAuthToken(ctx, store.CreateAuthTokenInput{
		AccountRoleID: accountRoleID,
		RefreshToken:  refreshToken,
	}); err != nil {
		return nil, fmt.Errorf("create auth token: %w", err)
	}

	return &model.AuthToken{
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) NewToken(ctx context.Context, accountRole *model.AccountRole) (*model.AuthenticatedInfo, error) {
	authToken, err := s.CreateAuthToken(ctx, accountRole.ID)
	if err != nil {
		return nil, err
	}

	return s.newAuthenticatedInfo(accountRole, authToken)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthenticatedInfo, error) {
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
	authStore, authStoreCloser := store.MustNew(s.app.Config())
	s.app.OnTerminate().Add(func(ctx context.Context, e *app.OnTerminateEvent) error {
		if authStoreCloser != nil {
			authStoreCloser()
		}

		return nil
	})
	s.store = authStore

	s.app.OnBeforeServe().Add(func(ctx context.Context, e *app.OnBeforeServeEvent) error {
		rootRouter := e.Server.Router()
		authMiddleware := s.NewMiddleware()

		rootRouter.Route("/auth", func(r chi.Router) {
			// public routes
			r.Post("/signup", s.SignupHandler)
			r.Post("/login", s.LoginHandler)
			r.Get("/{provider}", s.Oauth2AuthenticateHandler)
			r.Get("/{provider}/callback", s.Oauth2LoginSignupCallbackHandler)

			// private routes
			r.With(authMiddleware).Get("/me", s.MeHandler)
		})

		return nil
	})

	return nil
}

func (s *Service) newAuthenticatedInfo(
	accountRole *model.AccountRole,
	authToken *model.AuthToken,
) (*model.AuthenticatedInfo, error) {
	claims := model.CustomClaims{
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

	return &model.AuthenticatedInfo{
		RefreshToken: authToken.RefreshToken,
		Type:         "Bearer",
		Token:        jwtToken,
		ExpiresIn:    int64(s.tokenLifeTimeMinutes.Seconds()),
	}, nil
}

func (s *Service) getAuthenticatedInfo(ctx context.Context, accountRole *model.AccountRole) (*model.AuthenticatedInfo, error) {
	authToken, err := s.store.GetAuthTokenByAccountRoleID(ctx, accountRole.ID)
	if err != nil {
		return nil, fmt.Errorf("get auth token by account id: %w", err)
	}

	return s.newAuthenticatedInfo(accountRole, authToken)
}
