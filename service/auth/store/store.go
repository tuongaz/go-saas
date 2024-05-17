package store

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tuongaz/go-saas/pkg/timer"
	"github.com/tuongaz/go-saas/pkg/uid"
	model2 "github.com/tuongaz/go-saas/service/auth/model"
	"github.com/tuongaz/go-saas/service/auth/store/persistence"
	"github.com/tuongaz/go-saas/service/auth/store/persistence/postgres"
)

//go:embed persistence/postgres/postgres.sql
var postgresSchema string

var _ Interface = (*Impl)(nil)

type Interface interface {
	CreateAuthToken(ctx context.Context, row CreateAuthTokenInput) (*model2.AuthToken, error)
	GetAuthTokenByAccountRoleID(ctx context.Context, accountRoleID string) (*model2.AuthToken, error)
	GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model2.AuthToken, error)
	UpdateRefreshToken(ctx context.Context, id string, refreshToken string) error
	CreateOwnerAccount(ctx context.Context, input CreateOwnerAccountInput) (
		*model2.Account,
		*model2.Organisation,
		*model2.AuthProvider,
		*model2.AccountRole,
		error,
	)
	GetUserByEmail(ctx context.Context, email string) (*model2.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	GetAccount(ctx context.Context, accountID string) (*model2.Account, error)
	GetAccountRole(ctx context.Context, organisationID, accountID string) (*model2.AccountRole, error)
	GetAccountByAuthProvider(ctx context.Context, provider string, providerUserID string) (*model2.Account, error)
	GetOrganisationByAccountIDAndRole(ctx context.Context, accountID, role string) (*model2.Organisation, error)
	GetAccountRoleByID(ctx context.Context, accountRoleID string) (*model2.AccountRole, error)
	GetAccountRoles(ctx context.Context, accountID string) ([]*model2.AccountRole, error)
}

func New(db *sqlx.DB) (*Impl, error) {
	postgresDB := postgres.NewFromDB(db)

	if _, err := db.Exec(postgresSchema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Impl{
		db: postgresDB,
	}, nil
}

type Impl struct {
	db persistence.Interface
}

func (i *Impl) GetUserByEmail(ctx context.Context, email string) (*model2.User, error) {
	row, err := i.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	return toUser(*row), nil
}

func (i *Impl) EmailExists(ctx context.Context, email string) (bool, error) {
	return i.db.EmailExists(ctx, email)
}

func (i *Impl) CreateAuthToken(ctx context.Context, input CreateAuthTokenInput) (*model2.AuthToken, error) {
	row := persistence.AuthTokenRow{
		ID:            uid.ID(),
		AccountRoleID: input.AccountRoleID,
		RefreshToken:  input.RefreshToken,
		CreatedAt:     timer.Now(),
		UpdatedAt:     timer.Now(),
	}

	if _, err := i.db.CreateAuthToken(ctx, row); err != nil {
		return nil, fmt.Errorf("persist auth token: %w", err)
	}

	return toAuthToken(row), nil
}

func (i *Impl) UpdateRefreshToken(ctx context.Context, id string, refreshToken string) error {
	if _, err := i.db.UpdateRefreshToken(ctx, id, refreshToken); err != nil {
		return fmt.Errorf("update refresh token: %w", err)
	}

	return nil
}

func (i *Impl) GetAuthTokenByAccountRoleID(ctx context.Context, accountRoleID string) (*model2.AuthToken, error) {
	row, err := i.db.GetAuthTokenByAccountRoleID(ctx, accountRoleID)
	if err != nil {
		return nil, fmt.Errorf("query auth token: %w", err)
	}

	return toAuthToken(*row), nil
}

func (i *Impl) GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model2.AuthToken, error) {
	row, err := i.db.GetAuthTokenByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("query auth token by refresh token: %w", err)
	}

	return toAuthToken(*row), nil
}

func toUser(row persistence.UserRow) *model2.User {
	return &model2.User{
		ID:                                row.ID,
		Email:                             row.Email,
		Name:                              row.Name,
		Password:                          row.Password,
		ResetPasswordCode:                 row.ResetPasswordCode,
		ResetPasswordCodeExpiredTimestamp: row.ResetPasswordCodeExpiredTimestamp,
		CreatedAt:                         row.CreatedAt,
		UpdatedAt:                         row.UpdatedAt,
	}
}

type CreateAuthTokenInput struct {
	AccountRoleID string
	RefreshToken  string
}

type UpdateAuthTokenInput struct {
	RefreshToken string
}

type CreateUserInput struct {
	Email    string
	Name     string
	Password string
}

func (i *Impl) CreateOwnerAccount(ctx context.Context, input CreateOwnerAccountInput) (
	*model2.Account,
	*model2.Organisation,
	*model2.AuthProvider,
	*model2.AccountRole,
	error,
) {
	var userRow *persistence.UserRow
	if input.Provider == model2.AuthProviderUsernamePassword {
		found, err := i.db.EmailExists(ctx, input.Email)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("get user by email: %w", err)
		}

		if found {
			return nil, nil, nil, nil, fmt.Errorf("user already exists")
		}

		userRow = &persistence.UserRow{
			ID:        uid.ID(),
			Email:     input.Email,
			Name:      input.Name,
			Password:  input.Password,
			CreatedAt: timer.Now(),
			UpdatedAt: timer.Now(),
		}
		input.ProviderUserID = userRow.ID
	}

	accountRow := persistence.AccountRow{
		ID:                 uid.ID(),
		Name:               input.Name,
		FirstName:          input.FirstName,
		LastName:           input.LastName,
		Avatar:             input.Avatar,
		CommunicationEmail: input.Email,
		CreatedAt:          timer.Now(),
		UpdatedAt:          timer.Now(),
	}

	orgRow := persistence.OrganisationRow{
		ID:        uid.ID(),
		CreatedAt: timer.Now(),
		UpdatedAt: timer.Now(),
	}

	accRoleRow := persistence.AccountRoleRow{
		ID:             uid.ID(),
		OrganisationID: orgRow.ID,
		AccountID:      accountRow.ID,
		Role:           "OWNER",
		CreatedAt:      timer.Now(),
		UpdatedAt:      timer.Now(),
	}

	authProviderRow := persistence.AuthProviderRow{
		ID:             uid.ID(),
		Name:           input.Name,
		Provider:       input.Provider,
		ProviderUserID: input.ProviderUserID,
		Email:          input.Email,
		Avatar:         input.Avatar,
		AccountID:      accountRow.ID,
		LastLogin:      timer.Now(),
		CreatedAt:      timer.Now(),
		UpdatedAt:      timer.Now(),
	}

	if err := i.db.CreateOwnerAccount(
		ctx,
		accountRow,
		orgRow,
		authProviderRow,
		accRoleRow,
		userRow,
	); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create owner account: %w", err)
	}

	return toAccount(accountRow), toOrganisation(orgRow), toAuthProvider(authProviderRow), toAccountRole(accRoleRow), nil
}

func (i *Impl) GetAccount(ctx context.Context, accountID string) (*model2.Account, error) {
	row, err := i.db.GetAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query account: %w", err)
	}

	return toAccount(*row), nil
}

func (i *Impl) GetAccountRole(ctx context.Context, organisationID, accountID string) (*model2.AccountRole, error) {
	row, err := i.db.GetAccountRole(ctx, organisationID, accountID)
	if err != nil {
		return nil, fmt.Errorf("query account role: %w", err)
	}

	return toAccountRole(*row), nil
}

func (i *Impl) GetAccountByAuthProvider(ctx context.Context, provider string, providerUserID string) (*model2.Account, error) {
	row, err := i.db.GetAccountByAuthProvider(ctx, provider, providerUserID)
	if err != nil {
		return nil, fmt.Errorf("query account by provider: %w", err)
	}

	return toAccount(*row), nil
}

func (i *Impl) GetOrganisationByAccountIDAndRole(ctx context.Context, accountID, role string) (*model2.Organisation, error) {
	orgRow, err := i.db.GetOrganisationByAccountIDAndRole(ctx, accountID, role)
	if err != nil {
		return nil, fmt.Errorf("query default owner account by provider: %w", err)
	}

	return toOrganisation(*orgRow), nil
}

func (i *Impl) GetAccountRoles(ctx context.Context, accountID string) ([]*model2.AccountRole, error) {
	rows, err := i.db.GetAccountRoles(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query account roles: %w", err)
	}

	accountRoles := make([]*model2.AccountRole, 0, len(rows))
	for _, row := range rows {
		accountRoles = append(accountRoles, toAccountRole(*row))
	}

	return accountRoles, nil
}

func (i *Impl) GetAccountRoleByID(ctx context.Context, accountRoleID string) (*model2.AccountRole, error) {
	row, err := i.db.GetAccountRoleByID(ctx, accountRoleID)
	if err != nil {
		return nil, fmt.Errorf("query account role by id: %w", err)
	}

	return toAccountRole(*row), nil
}

type CreateOwnerAccountInput struct {
	Name           string
	FirstName      string
	LastName       string
	Email          string
	Provider       string
	ProviderUserID string
	Avatar         string
	Password       string
}

func toAccount(row persistence.AccountRow) *model2.Account {
	return &model2.Account{
		ID:                 row.ID,
		Name:               row.Name,
		FirstName:          row.FirstName,
		LastName:           row.LastName,
		Avatar:             row.Avatar,
		CommunicationEmail: row.CommunicationEmail,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func toOrganisation(row persistence.OrganisationRow) *model2.Organisation {
	return &model2.Organisation{
		ID:        row.ID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toAuthProvider(row persistence.AuthProviderRow) *model2.AuthProvider {
	return &model2.AuthProvider{
		ID:             row.ID,
		AccountID:      row.AccountID,
		Provider:       row.Provider,
		Email:          row.Email,
		Name:           row.Name,
		Avatar:         row.Avatar,
		ProviderUserID: row.ProviderUserID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func toAccountRole(row persistence.AccountRoleRow) *model2.AccountRole {
	return &model2.AccountRole{
		ID:             row.ID,
		AccountID:      row.AccountID,
		OrganisationID: row.OrganisationID,
		Role:           row.Role,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func toAuthToken(row persistence.AuthTokenRow) *model2.AuthToken {
	return &model2.AuthToken{
		ID:            row.ID,
		AccountRoleID: row.AccountRoleID,
		RefreshToken:  row.RefreshToken,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
