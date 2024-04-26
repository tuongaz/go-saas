package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/model"
	"github.com/autopus/bootstrap/pkg/auth/oauth2"
	"github.com/autopus/bootstrap/pkg/auth/signer"
	"github.com/autopus/bootstrap/pkg/baseurl"
	"github.com/autopus/bootstrap/pkg/encrypt"
	"github.com/autopus/bootstrap/pkg/errors"
	"github.com/autopus/bootstrap/pkg/hooks"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/store"
)

type OnAccountCreatedEvent struct {
	AccountID      string
	OrganisationID string
}

type Options struct {
	JWTLifeTimeMinutes int
}

func WithJWTLifeTimeMinutes(minutes int) func(*Options) {
	return func(c *Options) {
		c.JWTLifeTimeMinutes = minutes
	}
}

func New(
	cfg config.Interface,
	store store.AuthInterface,
	encryptor encrypt.Interface,
	signer signer.Interface,
	opts ...func(*Options),
) (*Service, error) {
	defaultOptions := Options{
		JWTLifeTimeMinutes: 1,
	}
	for _, opt := range opts {
		opt(&defaultOptions)
	}

	s := &Service{
		cfg:                  cfg,
		issuer:               cfg.GetJWTIssuer(),
		store:                store,
		signer:               signer,
		encryptor:            encryptor,
		tokenLifeTimeMinutes: time.Minute * time.Duration(defaultOptions.JWTLifeTimeMinutes),

		onAccountCreated: &hooks.Hook[*OnAccountCreatedEvent]{},
	}

	return s, nil
}

type Service struct {
	cfg                  config.Interface
	store                store.AuthInterface
	issuer               string
	signer               signer.Interface
	encryptor            encrypt.Interface
	tokenLifeTimeMinutes time.Duration

	onAccountCreated *hooks.Hook[*OnAccountCreatedEvent]
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
			Issuer:  s.issuer,
			Subject: accountRole.AccountID,
			Audience: jwt.ClaimStrings{
				s.issuer,
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

func (s *Service) oauth2SignupLogin(w http.ResponseWriter, r *http.Request, user oauth2.User) {
	ctx := r.Context()

	authInfo, err := s.oauth2SignupOrLogin(
		ctx,
		user,
	)
	if err != nil {
		log.Default().ErrorContext(ctx, "failed to signup or login", log.ErrorAttr(err))
		http.Redirect(w, r, fmt.Sprintf("%s/auth/signin-failed", baseurl.Get(ctx)), http.StatusFound)
		return
	}

	redirectURL := fmt.Sprintf(
		"%s?token=%s&refresh_token=%s",
		fmt.Sprintf("%s/auth/signin-success", baseurl.Get(ctx)),
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
