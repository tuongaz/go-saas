package store

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/service/auth/store/persistence/postgres"
)

//go:embed persistence/postgres/postgres.sql
var postgresSchema string

var _ Interface = (*Impl)(nil)

type Interface interface {
	CreateAuthToken(ctx context.Context, row CreateAuthTokenInput) (*model.AuthToken, error)
	GetAuthTokenByAccountRoleID(ctx context.Context, accountRoleID string) (*model.AuthToken, error)
	GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*model.AuthToken, error)
	UpdateAuthToken(ctx context.Context, id string, row UpdateAuthTokenInput) error
	CreateOwnerAccount(ctx context.Context, input CreateOwnerAccountInput) (
		*model.Account,
		*model.Organisation,
		*model.AuthProvider,
		*model.AccountRole,
		error,
	)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetAccount(ctx context.Context, accountID string) (*model.Account, error)
	GetAccountRole(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error)
	GetDefaultOwnerAccountByProvider(ctx context.Context, provider string, providerUserID string) (*model.Account, *model.Organisation, error)
	GetAccountRoleByID(ctx context.Context, accountRoleID string) (*model.AccountRole, error)
	GetAccountRoles(ctx context.Context, accountID string) ([]*model.AccountRole, error)
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
