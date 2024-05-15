package store

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/service/auth/store/persistence"
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

func NewFromDB(db persistence.Interface) *Impl {
	return &Impl{
		db: db,
	}
}

func MustNew(cfg config.Interface) (*Impl, func()) {
	db, closer, err := postgres.New("host=localhost port=5432 user=postgres password=password sslmode=disable")
	if err != nil {
		log.Default().Error("failed to init postgres db", log.ErrorAttr(err))
		panic(err)
	}

	dbname := "gosaas"

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbname)
	err = db.Connection().Get(&exists, query)
	if err != nil {
		panic(err)
	}

	if !exists {
		_, err := db.Connection().Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			panic(err)
		}
		fmt.Println("Database created.")
	} else {
		fmt.Println("Database already exists.")
	}

	if _, err := db.Connection().Exec(postgresSchema); err != nil {
		panic(err)
	}

	return NewFromDB(db), closer
}
