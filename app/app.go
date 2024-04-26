package app

import (
	"context"
	"fmt"

	aapp "github.com/autopus/bootstrap"
	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/hooks"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/store"
	"github.com/autopus/bootstrap/store/persistence/sql/sqlite"
)

type OnBeforeBootstrapEvent struct {
	App *App
}

type OnAfterBootstrapEvent struct {
	App *App
}

type OnBeforeStoreBootstrapEvent struct {
	App *App
}

type OnAfterStoreBootstrapEvent struct {
	App   *App
	Store store.Interface
}

type App struct {
	Cfg      config.Interface
	Store    store.Interface
	dbCloser func()

	onBeforeBootstrap      *hooks.Hook[*OnBeforeBootstrapEvent]
	onAfterBootstrap       *hooks.Hook[*OnAfterBootstrapEvent]
	onBeforeStoreBootstrap *hooks.Hook[*OnBeforeStoreBootstrapEvent]
	onAfterStoreBootstrap  *hooks.Hook[*OnAfterStoreBootstrapEvent]
}

func New(cfg config.Interface) *App {
	return &App{
		Cfg: cfg,

		onBeforeBootstrap:      &hooks.Hook[*OnBeforeBootstrapEvent]{},
		onAfterBootstrap:       &hooks.Hook[*OnAfterBootstrapEvent]{},
		onBeforeStoreBootstrap: &hooks.Hook[*OnBeforeStoreBootstrapEvent]{},
		onAfterStoreBootstrap:  &hooks.Hook[*OnAfterStoreBootstrapEvent]{},
	}
}

func (a *App) Start() error {
	ctx := context.Background()

	if err := a.bootstrap(ctx); err != nil {
		return fmt.Errorf("app bootstrap: %w", err)
	}

	return nil
}

func (a *App) Shutdown() error {
	if a.dbCloser != nil {
		a.dbCloser()
	}

	return nil
}

func (a *App) OnBeforeBootstrap() *hooks.Hook[*OnBeforeBootstrapEvent] {
	return a.onBeforeBootstrap
}

func (a *App) OnAfterBootstrap() *hooks.Hook[*OnAfterBootstrapEvent] {
	return a.onAfterBootstrap
}

func (a *App) BootstrapStore(ctx context.Context) error {
	if err := a.onBeforeStoreBootstrap.Trigger(ctx, &OnBeforeStoreBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("before store start: %w", err)
	}

	a.Store, a.dbCloser = mustNewStore(a.Cfg)

	if err := a.onAfterStoreBootstrap.Trigger(ctx, &OnAfterStoreBootstrapEvent{
		App:   a,
		Store: a.Store,
	}); err != nil {
		return fmt.Errorf("after store start: %w", err)
	}

	return nil
}

func (a *App) bootstrap(ctx context.Context) error {
	if err := a.onBeforeBootstrap.Trigger(ctx, &OnBeforeBootstrapEvent{}); err != nil {
		return fmt.Errorf("before bootstrap: %w", err)
	}

	if err := a.BootstrapStore(ctx); err != nil {
		return fmt.Errorf("store bootstrap: %w", err)
	}

	if err := a.onAfterBootstrap.Trigger(ctx, &OnAfterBootstrapEvent{}); err != nil {
		return fmt.Errorf("after bootstrap: %w", err)
	}

	return nil
}

func mustNewStore(cfg config.Interface) (store.Interface, func()) {
	log.Default().Info("using sqlite db")
	db, closer, err := sqlite.New(cfg.GetSQLiteDatasource())
	if err != nil {
		log.Default().Error("failed to init sqlite db", log.ErrorAttr(err))
		panic(err)
	}

	if !db.DBExists() {
		_, err = db.Conn().Exec(aapp.SqliteSchema)
		if err != nil {
			log.Default().Error("failed to init sqlite db schema", log.ErrorAttr(err))
			panic(err)
		}
	}

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

	return store.New(db), closer
}
