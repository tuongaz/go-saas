package store

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/tuongaz/go-saas/pkg/timer"
	"github.com/tuongaz/go-saas/pkg/uid"
	"github.com/tuongaz/go-saas/service/auth/model"
	"github.com/tuongaz/go-saas/store"
)

//go:embed schema/postgres.sql
var postgresSchema string

const (
	tableLoginCredentialsUser    = "login_credentials_user"
	tableAccessToken             = "access_token"
	tableAccount                 = "account"
	tableOrganisation            = "organisation"
	tableOrganisationAccountRole = "organisation_account_role"
	tableLoginProvider           = "login_provider"
)

var _ Interface = (*Store)(nil)

type GetAccessTokenInput struct {
	AccountRoleID  string
	ProviderUserID string
	Device         string
}

type CreateOwnerAccountInput struct {
	Name           string
	FirstName      string
	LastName       string
	Provider       string
	ProviderUserID string
	Avatar         string
	Email          string
	Password       string
}

type CreateAccessTokenInput struct {
	AccountRoleID  string
	Device         string
	ProviderUserID string
	RefreshToken   string
}

type UpdateAuthTokenInput struct {
	RefreshToken string
}

type CreateUserInput struct {
	Email    string
	Name     string
	Password string
}

type Interface interface {
	CreateAccessToken(ctx context.Context, input CreateAccessTokenInput) (*model.AccessToken, error)
	UpdateRefreshToken(ctx context.Context, id string, refreshToken string) error
	GetAccessToken(ctx context.Context, input GetAccessTokenInput) (*model.AccessToken, error)
	GetAccessTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, error)

	CreateOwnerAccount(ctx context.Context, input CreateOwnerAccountInput) (
		*model.Account,
		*model.Organisation,
		*model.LoginProvider,
		*model.AccountRole,
		error,
	)
	GetLoginCredentialsUserByEmail(ctx context.Context, email string) (*model.LoginCredentialsUser, error)
	LoginCredentialsUserEmailExists(ctx context.Context, email string) (bool, error)
	GetAccount(ctx context.Context, accountID string) (*model.Account, error)
	GetAccountByLoginProvider(ctx context.Context, provider string, providerUserID string) (*model.Account, error)
	GetAccountRoleByID(ctx context.Context, accountRoleID string) (*model.AccountRole, error)
	GetAccountRoleByOrgAndAccountID(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error)
	GetOrganisationByAccountIDAndRole(ctx context.Context, accountID, role string) (*model.Organisation, error)
	GetOrganisation(ctx context.Context, organisationID string) (*model.Organisation, error)
}

func New(store store.Interface) (*Store, error) {
	if err := store.Exec(context.Background(), postgresSchema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Store{
		store: store,
	}, nil
}

type Store struct {
	store store.Interface
}

func (s *Store) CreateAccessToken(ctx context.Context, input CreateAccessTokenInput) (*model.AccessToken, error) {
	record, err := s.store.Collection(tableAccessToken).CreateRecord(ctx, store.Record{
		"id":               uid.ID(),
		"refresh_token":    input.RefreshToken,
		"account_role_id":  input.AccountRoleID,
		"device":           input.Device,
		"provider_user_id": input.ProviderUserID,
		"created_at":       timer.Now(),
		"updated_at":       timer.Now(),
		"last_accessed_at": timer.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}

	accessToken := &model.AccessToken{}
	if err := record.Decode(accessToken); err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (s *Store) UpdateRefreshToken(ctx context.Context, id string, refreshToken string) error {
	_, err := s.store.Collection(tableAccessToken).UpdateRecord(
		ctx,
		id,
		store.Record{"refresh_token": refreshToken, "updated_at": timer.Now()},
	)
	if err != nil {
		return fmt.Errorf("update refresh token: %w", err)
	}

	return nil
}

func (s *Store) GetAccessToken(ctx context.Context, input GetAccessTokenInput) (*model.AccessToken, error) {
	record, err := s.store.Collection(tableAccessToken).FindOne(ctx, store.Record{
		"account_role_id":  input.AccountRoleID,
		"provider_user_id": input.ProviderUserID,
		"device":           input.Device,
	})
	if err != nil {
		if store.IsNotFoundError(err) {
			return s.CreateAccessToken(ctx, CreateAccessTokenInput{
				AccountRoleID:  input.AccountRoleID,
				Device:         input.Device,
				ProviderUserID: input.ProviderUserID,
				RefreshToken:   uuid.NewString(),
			})
		}

		return nil, fmt.Errorf("get access token: %w", err)
	}

	accessToken := &model.AccessToken{}
	if err := record.Decode(accessToken); err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (s *Store) GetLoginCredentialsUserByEmail(ctx context.Context, email string) (*model.LoginCredentialsUser, error) {
	record, err := s.store.Collection(tableLoginCredentialsUser).FindOne(ctx, store.Record{"email": email})
	if err != nil {
		return nil, err
	}

	user := &model.LoginCredentialsUser{}
	if err := record.Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) LoginCredentialsUserEmailExists(ctx context.Context, email string) (bool, error) {
	return s.store.Collection(tableLoginCredentialsUser).Exists(ctx, store.Record{"email": email})
}

func (s *Store) GetAccessTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, error) {
	record, err := s.store.Collection(tableAccessToken).FindOne(ctx, store.Record{"refresh_token": refreshToken})
	if err != nil {
		return nil, fmt.Errorf("get access token by refresh token: %w", err)
	}

	accessToken := &model.AccessToken{}
	if err := record.Decode(accessToken); err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (s *Store) CreateOwnerAccount(ctx context.Context, input CreateOwnerAccountInput) (
	mAccount *model.Account,
	mOrg *model.Organisation,
	mLoginProvider *model.LoginProvider,
	mAccountRole *model.AccountRole,
	err error,
) {
	tx, err := s.store.Tx(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var userRecord *store.Record
	if input.Provider == model.AuthProviderUsernamePassword {
		found, err := s.LoginCredentialsUserEmailExists(ctx, input.Email)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if found {
			return nil, nil, nil, nil, fmt.Errorf("login credentials already exists")
		}

		userRecord = &store.Record{
			"id":                                    uid.ID(),
			"email":                                 input.Email,
			"name":                                  input.Name,
			"password":                              input.Password,
			"reset_password_code":                   "",
			"reset_password_code_expired_timestamp": nil,
			"created_at":                            timer.Now(),
			"updated_at":                            timer.Now(),
		}
		input.ProviderUserID = userRecord.Get("id").(string)
	}

	accountRecord := store.Record{
		"id":                  uid.ID(),
		"name":                input.Name,
		"first_name":          input.FirstName,
		"last_name":           input.LastName,
		"avatar":              input.Avatar,
		"communication_email": input.Email,
		"created_at":          timer.Now(),
		"updated_at":          timer.Now(),
	}

	orgRecord := store.Record{
		"id":         uid.ID(),
		"created_at": timer.Now(),
		"updated_at": timer.Now(),
	}

	accRoleRecord := store.Record{
		"id":              uid.ID(),
		"organisation_id": orgRecord.Get("id"),
		"account_id":      accountRecord.Get("id"),
		"role":            "OWNER",
		"created_at":      timer.Now(),
		"updated_at":      timer.Now(),
	}

	loginProviderRecord := store.Record{
		"id":               uid.ID(),
		"name":             input.Name,
		"provider":         input.Provider,
		"provider_user_id": input.ProviderUserID,
		"email":            input.Email,
		"avatar":           input.Avatar,
		"account_id":       accountRecord.Get("id"),
		"last_login":       timer.Now(),
		"created_at":       timer.Now(),
		"updated_at":       timer.Now(),
	}

	if userRecord != nil {
		if _, err := tx.Collection(tableLoginCredentialsUser).CreateRecord(ctx, *userRecord); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("create user: %w", err)
		}
	}

	if _, err := tx.Collection(tableAccount).CreateRecord(ctx, accountRecord); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create account: %w", err)
	}

	mAccount = &model.Account{}
	if err := accountRecord.Decode(mAccount); err != nil {
		return nil, nil, nil, nil, err
	}

	if _, err := tx.Collection(tableOrganisation).CreateRecord(ctx, orgRecord); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create organisation: %w", err)
	}

	mOrg = &model.Organisation{}
	if err := orgRecord.Decode(mOrg); err != nil {
		return nil, nil, nil, nil, err
	}

	if _, err := tx.Collection(tableOrganisationAccountRole).CreateRecord(ctx, accRoleRecord); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create account role: %w", err)
	}

	mAccountRole = &model.AccountRole{}
	if err := accRoleRecord.Decode(mAccountRole); err != nil {
		return nil, nil, nil, nil, err
	}

	if _, err := tx.Collection(tableLoginProvider).CreateRecord(ctx, loginProviderRecord); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create auth provider: %w", err)
	}

	mLoginProvider = &model.LoginProvider{}
	if err := loginProviderRecord.Decode(mLoginProvider); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("commit tx: %w", err)
	}

	return mAccount, mOrg, mLoginProvider, mAccountRole, nil
}

func (s *Store) GetAccount(ctx context.Context, accountID string) (*model.Account, error) {
	record, err := s.store.Collection(tableAccount).GetRecord(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	account := &model.Account{}
	if err := record.Decode(account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *Store) GetAccountByLoginProvider(ctx context.Context, provider string, providerUserID string) (*model.Account, error) {
	loginProvider, err := s.store.Collection(tableLoginProvider).FindOne(ctx, store.Record{
		"provider":         provider,
		"provider_user_id": providerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("get account by login provider: %w", err)
	}

	accountID := loginProvider.Get("account_id").(string)

	return s.GetAccount(ctx, accountID)
}

func (s *Store) GetAccountRoleByID(ctx context.Context, accountRoleID string) (*model.AccountRole, error) {
	record, err := s.store.Collection(tableOrganisationAccountRole).GetRecord(ctx, accountRoleID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	accountRole := &model.AccountRole{}
	if err := record.Decode(accountRole); err != nil {
		return nil, err
	}

	return accountRole, nil
}

func (s *Store) GetAccountRoleByOrgAndAccountID(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error) {
	record, err := s.store.Collection(tableOrganisationAccountRole).FindOne(ctx, store.Record{
		"organisation_id": organisationID,
		"account_id":      accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("get account role by org and account: %w", err)
	}

	accountRole := &model.AccountRole{}
	if err := record.Decode(accountRole); err != nil {
		return nil, err
	}

	return accountRole, nil
}

func (s *Store) GetOrganisation(ctx context.Context, organisationID string) (*model.Organisation, error) {
	record, err := s.store.Collection(tableOrganisation).GetRecord(ctx, organisationID)
	if err != nil {
		return nil, fmt.Errorf("get organisation: %w", err)
	}

	org := &model.Organisation{}
	if err := record.Decode(org); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *Store) GetOrganisationByAccountIDAndRole(ctx context.Context, accountID, role string) (*model.Organisation, error) {
	record, err := s.store.Collection(tableOrganisationAccountRole).FindOne(ctx, store.Record{
		"account_id": accountID,
		"role":       role,
	})
	if err != nil {
		return nil, fmt.Errorf("get organisation by account id and role: %w", err)
	}

	orgID := record.Get("organisation_id").(string)

	return s.GetOrganisation(ctx, orgID)
}
