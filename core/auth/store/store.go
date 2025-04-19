package store

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"

	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/pkg/timer"
	"github.com/tuongaz/go-saas/pkg/uid"
	"github.com/tuongaz/go-saas/store"
	"github.com/tuongaz/go-saas/store/types"
)

//go:embed postgres.sql
var postgresSchema string

const (
	TableAccount      = "account"
	TableOrganisation = "organisation"

	tableLoginCredentialsUser              = "login_credentials_user"
	tableLoginCredentialsUserResetPassword = "login_credentials_user_reset_password"
	tableAccessToken                       = "access_token"
	tableOrganisationAccountRole           = "organisation_account_role"
	tableLoginProvider                     = "login_provider"
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

	GetLoginCredentialsUserByEmail(ctx context.Context, email string) (*model.LoginCredentialsUser, error)
	LoginCredentialsUserEmailExists(ctx context.Context, email string) (bool, error)
	CreateResetPasswordRequest(ctx context.Context, userID, code string) (*model.ResetPasswordRequest, error)
	UpdateResetPasswordReceipt(ctx context.Context, id string, receipt string) error
	GetResetPasswordRequest(ctx context.Context, code string) (*model.ResetPasswordRequest, error)
	UpdateLoginCredentialsUserPassword(ctx context.Context, userID, password string) error
	DeleteResetPasswordRequest(ctx context.Context, id string) error
	GetLoginProviderByAccountID(ctx context.Context, accountID, provider string) (*model.LoginProvider, error)

	CreateOwnerAccount(ctx context.Context, input CreateOwnerAccountInput) (
		*model.Account,
		*model.Organisation,
		*model.LoginProvider,
		*model.AccountRole,
		error,
	)
	GetAccount(ctx context.Context, accountID string) (*model.Account, error)
	GetAccountByLoginProvider(ctx context.Context, provider string, providerUserID string) (*model.Account, error)
	GetAccountRoleByID(ctx context.Context, accountRoleID string) (*model.AccountRole, error)
	GetAccountRoleByOrgAndAccountID(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error)
	GetOrganisationByAccountIDAndRole(ctx context.Context, accountID, role string) (*model.Organisation, error)
	GetOrganisation(ctx context.Context, organisationID string) (*model.Organisation, error)
	UpdateAccount(ctx context.Context, accountID string, account *model.Account) (*model.Account, error)

	// Organisation operations
	ListOrganisationsByAccountID(ctx context.Context, accountID string) ([]model.Organisation, error)
	CreateOrganisation(ctx context.Context, input CreateOrganisationInput) (*model.Organisation, error)
	UpdateOrganisation(ctx context.Context, input UpdateOrganisationInput) (*model.Organisation, error)
	DeleteOrganisation(ctx context.Context, organisationID string) error

	// Organisation member operations
	AddOrganisationMember(ctx context.Context, input AddOrganisationMemberInput) (*model.AccountRole, error)
	ListOrganisationMembers(ctx context.Context, organisationID string) ([]model.AccountRole, error)
	RemoveOrganisationMember(ctx context.Context, organisationID, accountID string) error
	UpdateOrganisationMemberRole(ctx context.Context, input UpdateOrganisationMemberRoleInput) (*model.AccountRole, error)
}

func New(store store.Interface) (*Store, error) {
	if err := store.Exec(context.Background(), postgresSchema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	s := &Store{
		store: store,
	}

	// Run migrations
	db := store.DB()
	if err := RunMigrations(context.Background(), db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

type Store struct {
	store store.Interface
}

func (s *Store) CreateAccessToken(ctx context.Context, input CreateAccessTokenInput) (*model.AccessToken, error) {
	record, err := s.store.Collection(tableAccessToken).CreateRecord(ctx, types.Record{
		"id":               uid.ID(),
		"refresh_token":    input.RefreshToken,
		"account_role_id":  input.AccountRoleID,
		"device":           input.Device,
		"provider_user_id": input.ProviderUserID,
		"created_at":       timer.Now(),
		"updated_at":       timer.Now(),
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
		types.Record{"refresh_token": refreshToken, "updated_at": timer.Now()},
	)
	if err != nil {
		return fmt.Errorf("update refresh token: %w", err)
	}

	return nil
}

func (s *Store) GetAccessToken(ctx context.Context, input GetAccessTokenInput) (*model.AccessToken, error) {
	record, err := s.store.Collection(tableAccessToken).FindOne(ctx, store.Filter{
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
	record, err := s.store.Collection(tableLoginCredentialsUser).FindOne(ctx, store.Filter{"email": email})
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
	return s.store.Collection(tableLoginCredentialsUser).Exists(ctx, store.Filter{"email": email})
}

func (s *Store) GetAccessTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, error) {
	record, err := s.store.Collection(tableAccessToken).FindOne(ctx, store.Filter{"refresh_token": refreshToken})
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

	var userRecord *types.Record
	if input.Provider == model.AuthProviderUsernamePassword {
		found, err := s.LoginCredentialsUserEmailExists(ctx, input.Email)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if found {
			return nil, nil, nil, nil, fmt.Errorf("login credentials already exists")
		}

		userRecord = &types.Record{
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

	accountRecord := types.Record{
		"id":                  uid.ID(),
		"name":                input.Name,
		"first_name":          input.FirstName,
		"last_name":           input.LastName,
		"avatar":              input.Avatar,
		"communication_email": input.Email,
		"created_at":          timer.Now(),
		"updated_at":          timer.Now(),
	}

	orgRecord := types.Record{
		"id":          uid.ID(),
		"name":        input.Name,
		"description": "",
		"avatar":      "",
		"metadata":    "{}",
		"owner_id":    accountRecord.Get("id"),
		"created_at":  timer.Now(),
		"updated_at":  timer.Now(),
	}

	accRoleRecord := types.Record{
		"id":              uid.ID(),
		"organisation_id": orgRecord.Get("id"),
		"account_id":      accountRecord.Get("id"),
		"role":            "OWNER",
		"created_at":      timer.Now(),
		"updated_at":      timer.Now(),
	}

	loginProviderRecord := types.Record{
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

	if _, err := tx.Collection(TableAccount).CreateRecord(ctx, accountRecord); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create account: %w", err)
	}

	mAccount = &model.Account{}
	if err := accountRecord.Decode(mAccount); err != nil {
		return nil, nil, nil, nil, err
	}

	if _, err := tx.Collection(TableOrganisation).CreateRecord(ctx, orgRecord); err != nil {
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
	record, err := s.store.Collection(TableAccount).GetRecord(ctx, accountID)
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
	loginProvider, err := s.store.Collection(tableLoginProvider).FindOne(ctx, store.Filter{
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
	record, err := s.store.Collection(tableOrganisationAccountRole).FindOne(ctx, store.Filter{
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

func (s *Store) GetOrganisationByAccountIDAndRole(ctx context.Context, accountID, role string) (*model.Organisation, error) {
	record, err := s.store.Collection(tableOrganisationAccountRole).FindOne(ctx, store.Filter{
		"account_id": accountID,
		"role":       role,
	})
	if err != nil {
		return nil, fmt.Errorf("get organisation by account id and role: %w", err)
	}

	orgID := record.Get("organisation_id").(string)

	return s.GetOrganisation(ctx, orgID)
}

func (s *Store) CreateResetPasswordRequest(ctx context.Context, userID, code string) (*model.ResetPasswordRequest, error) {
	record, err := s.store.Collection(tableLoginCredentialsUserResetPassword).CreateRecord(ctx, types.Record{
		"id":         uid.ID(),
		"user_id":    userID,
		"code":       code,
		"created_at": timer.Now(),
		"updated_at": timer.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("create reset password request: %w", err)
	}

	resetPasswordRequest := &model.ResetPasswordRequest{}
	if err := record.Decode(resetPasswordRequest); err != nil {
		return nil, err
	}

	return resetPasswordRequest, nil
}

func (s *Store) UpdateResetPasswordReceipt(ctx context.Context, id string, receipt string) error {
	_, err := s.store.Collection(tableLoginCredentialsUserResetPassword).UpdateRecord(ctx, id, types.Record{
		"receipt":    receipt,
		"updated_at": timer.Now(),
	})
	if err != nil {
		return fmt.Errorf("update receipt: %w", err)
	}

	return nil
}

func (s *Store) GetResetPasswordRequest(ctx context.Context, code string) (*model.ResetPasswordRequest, error) {
	record, err := s.store.Collection(tableLoginCredentialsUserResetPassword).FindOne(ctx, store.Filter{"code": code})
	if err != nil {
		return nil, fmt.Errorf("get reset password request: %w", err)
	}

	resetPasswordRequest := &model.ResetPasswordRequest{}
	if err := record.Decode(resetPasswordRequest); err != nil {
		return nil, err
	}

	return resetPasswordRequest, nil
}

func (s *Store) UpdateLoginCredentialsUserPassword(ctx context.Context, userID, password string) error {
	_, err := s.store.Collection(tableLoginCredentialsUser).UpdateRecord(ctx, userID, types.Record{
		"password":   password,
		"updated_at": timer.Now(),
	})
	if err != nil {
		return fmt.Errorf("update login credentials user password: %w", err)
	}

	return nil
}

func (s *Store) DeleteResetPasswordRequest(ctx context.Context, id string) error {
	if err := s.store.Collection(tableLoginCredentialsUserResetPassword).DeleteRecord(ctx, id); err != nil {
		return fmt.Errorf("delete reset password request: %w", err)
	}

	return nil
}

func (s *Store) GetLoginProviderByAccountID(ctx context.Context, accountID, provider string) (*model.LoginProvider, error) {
	record, err := s.store.Collection(tableLoginProvider).FindOne(ctx, store.Filter{
		"account_id": accountID,
		"provider":   provider,
	})
	if err != nil {
		return nil, fmt.Errorf("get login provider by account id: %w", err)
	}

	loginProvider := &model.LoginProvider{}
	if err := record.Decode(loginProvider); err != nil {
		return nil, err
	}

	return loginProvider, nil
}

func (s *Store) UpdateAccount(ctx context.Context, accountID string, account *model.Account) (*model.Account, error) {
	record, err := s.store.Collection(TableAccount).UpdateRecord(ctx, accountID, types.Record{
		"name":                account.Name,
		"first_name":          account.FirstName,
		"last_name":           account.LastName,
		"avatar":              account.Avatar,
		"communication_email": account.CommunicationEmail,
		"updated_at":          timer.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}

	updatedAccount := &model.Account{}
	if err := record.Decode(updatedAccount); err != nil {
		return nil, err
	}

	return updatedAccount, nil
}
