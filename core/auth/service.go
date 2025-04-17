package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/core/auth/signer"
	"github.com/tuongaz/go-saas/core/auth/store"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/service/emailer"
	coreStore "github.com/tuongaz/go-saas/store"
)

// Interface defines the methods provided by the auth service
type Interface interface {
	// Core functionality
	Store() store.Interface
	GetAccount(ctx context.Context, accountID string) (*model.Account, error)
	GetAccountRole(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error)
	GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, error)
	CreateAccessToken(ctx context.Context, accountRoleID, providerUserID, device string) (*model.AccessToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.AuthenticatedInfo, error)
	OnAccountCreated() *hooks.Hook[*OnAccountCreatedEvent]

	// API Setup
	SetupAPI(router *chi.Mux)

	// Middleware
	NewMiddleware() func(next http.Handler) http.Handler
	NewDeviceMiddleware() func(next http.Handler) http.Handler
	ValidateOrganisation(ctx context.Context, organisationID string) error
	// Authentication handlers
	MeHandler(w http.ResponseWriter, r *http.Request)
	Oauth2EnabledProvidersHandler(w http.ResponseWriter, r *http.Request)
	Oauth2AuthenticateHandler(w http.ResponseWriter, r *http.Request)
	Oauth2LoginSignupCallbackHandler(w http.ResponseWriter, r *http.Request)
	SignupHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	GetResetPasswordHandler(w http.ResponseWriter, r *http.Request)
	ResetPasswordHandler(w http.ResponseWriter, r *http.Request)
	ResetPasswordConfirmHandler(w http.ResponseWriter, r *http.Request)
	RefreshTokenHandler(w http.ResponseWriter, r *http.Request)

	// Organisation handlers
	ListOrganisationsHandler(w http.ResponseWriter, r *http.Request)
	CreateOrganisationHandler(w http.ResponseWriter, r *http.Request)
	GetOrganisationHandler(w http.ResponseWriter, r *http.Request)
	UpdateOrganisationHandler(w http.ResponseWriter, r *http.Request)
	AddOrganisationMemberHandler(w http.ResponseWriter, r *http.Request)
	ListOrganisationMembersHandler(w http.ResponseWriter, r *http.Request)
	RemoveOrganisationMemberHandler(w http.ResponseWriter, r *http.Request)
	UpdateOrganisationMemberRoleHandler(w http.ResponseWriter, r *http.Request)
}

var _ Interface = &service{}

type service struct {
	emailer          emailer.Interface
	cfg              *config.Config
	store            store.Interface
	signer           signer.Interface
	encryptor        encrypt.Interface
	tokenLifeTime    time.Duration
	jwtIssuer        string
	providers        map[string]config.OAuth2ProviderConfig
	onAccountCreated *hooks.Hook[*OnAccountCreatedEvent]
}

func New(cfg *config.Config, emailer emailer.Interface, st coreStore.Interface) (*service, error) {
	authStore, err := store.New(st)
	if err != nil {
		return nil, err
	}

	authSrv := &service{
		cfg:              cfg,
		emailer:          emailer,
		signer:           signer.NewHS512Signer([]byte(cfg.JWTSigningSecret)),
		encryptor:        encrypt.New(cfg.EncryptionKey),
		jwtIssuer:        cfg.JWTIssuer,
		tokenLifeTime:    time.Duration(cfg.JWTTokenLifetimeSeconds) * time.Second,
		providers:        cfg.Oauth2AuthProviders,
		onAccountCreated: &hooks.Hook[*OnAccountCreatedEvent]{},
		store:            authStore,
	}

	return authSrv, nil
}

func (s *service) Store() store.Interface {
	return s.store
}

type InvalidOrganisationError struct{}

func (e *InvalidOrganisationError) Error() string {
	return "invalid organisation"
}

func (s *service) ValidateOrganisation(ctx context.Context, organisationID string) error {
	accountID := AccountID(ctx)
	userOrgs, err := s.store.ListOrganisationsByAccountID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get user organisations: %w", err)
	}

	for _, org := range userOrgs {
		if org.ID == organisationID {
			return nil
		}
	}

	return &InvalidOrganisationError{}
}

func (s *service) GetAccount(ctx context.Context, accountID string) (*model.Account, error) {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account by id: %w", err)
	}

	return account, nil
}

func (s *service) GetAccountRole(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error) {
	accRole, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, organisationID, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return accRole, nil
}

func (s *service) GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, error) {
	authToken, err := s.store.GetAccessTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("get auth token by refresh token: %w", err)
	}

	return authToken, nil
}

func (s *service) CreateAccessToken(ctx context.Context, accountRoleID, providerUserID, device string) (*model.AccessToken, error) {
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

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthenticatedInfo, error) {
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

func (s *service) newAuthenticatedInfo(
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

func (s *service) getAuthenticatedInfo(
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
