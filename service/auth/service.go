package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tuongaz/go-saas/app"
	"golang.org/x/crypto/bcrypt"

	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/pkg/auth/oauth2"
	"github.com/tuongaz/go-saas/pkg/auth/signer"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/errors"
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/service/auth/store"
)

const ID = "auth"

func (s *Service) ID() string {
	return ID
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
			r.Post("/token", s.AuthTokenHandler)
			r.Get("/token/authorization", s.TokenAuthorizationHandler)
			r.Get("/{provider}", s.Oauth2AuthenticateHandler)
			r.Get("/{provider}/callback", s.Oauth2LoginSignupCallbackHandler)

			// private routes
			r.With(authMiddleware).Get("/me", s.MeHandler)
		})

		return nil
	})

	return nil
}

func (s *Service) Start(ctx context.Context) error {
	log.Info("auth service started")
	return nil
}

type OnAccountCreatedEvent struct {
	AccountID      string
	OrganisationID string
}

type Service struct {
	app                  app.Interface
	store                store.Interface
	signer               signer.Interface
	encryptor            encrypt.Interface
	tokenLifeTimeMinutes time.Duration
	jwtIssuer            string
	redirectURL          string
	providers            map[string]OAuth2Provider
	onAccountCreated     *hooks.Hook[*OnAccountCreatedEvent]
}

type OAuth2Provider struct {
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
	Providers               []OAuth2Provider
}

func Register(appInstance app.Interface, cfg Config) {
	if cfg.JWTTokenLifetimeMinutes == 0 {
		cfg.JWTTokenLifetimeMinutes = 30
	}

	if cfg.JWTSigningSecret == "" {
		cfg.JWTSigningSecret = "signing-secret-please-change-me"
	}

	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "go-saas-issuer"
	}

	providers := make(map[string]OAuth2Provider)
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

	appInstance.OnBeforeServe().Add(func(ctx context.Context, e *app.OnBeforeServeEvent) error {
		if err := authSrv.Start(ctx); err != nil {
			return fmt.Errorf("starting auth service: %w", err)
		}
		return nil
	})
}

func (s *Service) OnAccountCreated() *hooks.Hook[*OnAccountCreatedEvent] {
	return s.onAccountCreated
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

// oauth2SignupOrLogin creates new account, with new organisation and assign owner role to the account
func (s *Service) oauth2SignupOrLogin(
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

	if ownerAcc == nil {
		newAccount = true
		org, ownerAcc, err = s.oauth2SignupNewAccount(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	accountRole, err := s.store.GetAccountRole(ctx, org.ID, ownerAcc.ID)
	if err != nil {
		return nil, err
	}

	if newAccount {
		if err := s.OnAccountCreated().Trigger(ctx, &OnAccountCreatedEvent{
			AccountID:      ownerAcc.ID,
			OrganisationID: org.ID,
		}); err != nil {
			return nil, fmt.Errorf("notify account created: %w", err)
		}
	}

	return s.getAuthenticatedInfo(ctx, accountRole)
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
		return nil, errors.New("invalid password")
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

func (s *Service) oauth2SignupLogin(w http.ResponseWriter, r *http.Request, oauthProvider OAuth2Provider, user oauth2.User) {
	ctx := r.Context()

	authInfo, err := s.oauth2SignupOrLogin(
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

func (s *Service) getAuthenticatedInfo(ctx context.Context, accountRole *model.AccountRole) (*model.AuthenticatedInfo, error) {
	authToken, err := s.store.GetAuthTokenByAccountRoleID(ctx, accountRole.ID)
	if err != nil {
		return nil, fmt.Errorf("get auth token by account id: %w", err)
	}

	return s.newAuthenticatedInfo(accountRole, authToken)
}
