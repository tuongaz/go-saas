package store

import (
	"context"
	_ "embed"

	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/model"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/service/auth/store/persistence"
	"github.com/tuongaz/go-saas/service/auth/store/persistence/sqlite"
)

//go:embed persistence/sqlite/sqlite.sql
var sqliteSchema string

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
	log.Default().Info("auth: using sqlite db")
	db, closer, err := sqlite.New(cfg.GetSQLiteDatasource())
	if err != nil {
		log.Default().Error("failed to init sqlite db", log.ErrorAttr(err))
		panic(err)
	}

	if !db.DBExists() {
		log.Default().Info("auth: initializing sqlite db schema")
		_, err = db.Conn().Exec(sqliteSchema)
		if err != nil {
			log.Default().Error("auth: failed to init sqlite db schema", log.ErrorAttr(err))
			panic(err)
		}
	}

	// TODO
	//if cfg.PostgresDataSource != "" {
	//	log.Default().Info("using postgres db")
	//	db, closer, err := postgres.New(cfg.PostgresDataSource)
	//	if err != nil {
	//		log.Default().Error("failed to init postgres db", log.ErrorAttr(err))
	//		panic(err)
	//	}
	//	return db, closer
	//}

	//log.Default().Info("using sqlite db")
	//db, closer, err := sqlite.New(cfg.SQLiteDatasource)
	//if err != nil {
	//	log.Default().Error("failed to init sqlite db", log.ErrorAttr(err))
	//	panic(err)
	//}

	return NewFromDB(db), closer
}
