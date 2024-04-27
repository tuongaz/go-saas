package store

import (
	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/store/persistence"
	"github.com/autopus/bootstrap/store/persistence/sql/sqlite"
)

var _ Interface = (*Impl)(nil)

type Interface interface {
	AuthInterface
}

type Impl struct {
	db persistence.Interface
}

func NewFromDB(db persistence.Interface) *Impl {
	return &Impl{
		db: db,
	}
}

func MustNew(cfg config.Interface) (*Impl, func()) {
	log.Default().Info("using sqlite db")
	db, closer, err := sqlite.New(cfg.GetSQLiteDatasource())
	if err != nil {
		log.Default().Error("failed to init sqlite db", log.ErrorAttr(err))
		panic(err)
	}

	if !db.DBExists() {
		_, err = db.Conn().Exec(cfg.GetSQLiteSchema())
		if err != nil {
			log.Default().Error("failed to init sqlite db schema", log.ErrorAttr(err))
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
