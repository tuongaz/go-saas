package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/core/auth"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/server"
	"github.com/tuongaz/go-saas/store"
)

var _ AppInterface = (*App)(nil)

type AppInterface interface {
	Store() store.Interface

	Auth() *auth.Service

	Config() *config.Config

	// OnBeforeBootstrap returns the hook that is triggered before the app is bootstrapped.
	OnBeforeBootstrap() *hooks.Hook[*OnBeforeBootstrapEvent]

	// OnAfterBootstrap returns the hook that is triggered after the app is bootstrapped.
	OnAfterBootstrap() *hooks.Hook[*OnAfterBootstrapEvent]

	// OnTerminate returns the hook that is triggered when the app is terminated.
	OnTerminate() *hooks.Hook[*OnTerminateEvent]

	// OnBeforeServe returns the hook that is triggered before the app starts serving.
	OnBeforeServe() *hooks.Hook[*OnBeforeServeEvent]

	// OnDatabaseBootstrap returns the hook that is triggered when the database is bootstrapped.
	OnDatabaseBootstrap() *hooks.Hook[*OnDatabaseBootstrapEvent]

	// Start the app
	Start() error

	PublicRoute(pattern string, fn func(r chi.Router))
	PrivateRoute(pattern string, fn func(r chi.Router))
}

type App struct {
	cfg                 *config.Config
	onBeforeBootstrap   *hooks.Hook[*OnBeforeBootstrapEvent]
	onAfterBootstrap    *hooks.Hook[*OnAfterBootstrapEvent]
	onBeforeServe       *hooks.Hook[*OnBeforeServeEvent]
	onTerminate         *hooks.Hook[*OnTerminateEvent]
	onDatabaseBootstrap *hooks.Hook[*OnDatabaseBootstrapEvent]
	store               store.Interface
	auth                *auth.Service
	server              *server.Server
	encryptor           encrypt.Interface
}

func New(opts ...func(cfg *config.Config)) (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, opt := range opts {
		opt(cfg)
	}

	encryptor := encrypt.New(cfg.EncryptionKey)

	return &App{
		cfg:                 cfg,
		onBeforeBootstrap:   &hooks.Hook[*OnBeforeBootstrapEvent]{},
		onAfterBootstrap:    &hooks.Hook[*OnAfterBootstrapEvent]{},
		onBeforeServe:       &hooks.Hook[*OnBeforeServeEvent]{},
		onTerminate:         &hooks.Hook[*OnTerminateEvent]{},
		onDatabaseBootstrap: &hooks.Hook[*OnDatabaseBootstrapEvent]{},
		encryptor:           encryptor,
	}, nil
}

func (a *App) Store() store.Interface {
	return a.store
}

func (a *App) Auth() *auth.Service {
	return a.auth
}

func (a *App) Config() *config.Config {
	return a.cfg
}

func (a *App) Encryptor() encrypt.Interface {
	return a.encryptor
}

func (a *App) PublicRoute(pattern string, fn func(r chi.Router)) {
	a.server.Router().Route(pattern, fn)
}

func (a *App) PrivateRoute(pattern string, fn func(r chi.Router)) {
	a.server.Router().Route(pattern, func(r chi.Router) {
		r.Use(a.auth.NewMiddleware())
		fn(r)
	})
}

func (a *App) Start() error {
	ctx := context.Background()

	if err := a.bootstrap(ctx); err != nil {
		return fmt.Errorf("app bootstrap: %w", err)
	}

	a.server = server.New(a.Config())

	authSrv, err := auth.New(a.Config(), a.store)
	if err != nil {
		return fmt.Errorf("new auth service: %w", err)
	}

	a.auth = authSrv
	a.auth.SetupAPI(a.server.Router())

	a.OnBeforeServe().Trigger(
		ctx,
		&OnBeforeServeEvent{
			App: a,
		},
	)

	a.server.PrintRoutes()

	done := make(chan bool, 1)
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	go func() {
		if err := a.server.Serve(); err != nil {
			log.Default().Error("failed to serve: %v", err)
		}

		done <- true
	}()

	<-done

	a.OnTerminate().Trigger(
		ctx,
		&OnTerminateEvent{App: a},
		func(ctx context.Context, e *OnTerminateEvent) error {
			return e.App.Shutdown()
		},
	)

	return nil
}

func (a *App) Shutdown() error {
	log.Default().Warn("shutting down app")

	return nil
}

func (a *App) bootstrap(ctx context.Context) error {
	a.OnBeforeBootstrap().Trigger(ctx, &OnBeforeBootstrapEvent{
		App: a,
	})

	log.Info("bootstrapping database")
	st, err := store.New(a.Config())
	if err != nil {
		return fmt.Errorf("new store: %w", err)
	}
	a.store = st

	a.OnDatabaseBootstrap().Trigger(ctx, &OnDatabaseBootstrapEvent{
		App: a,
	})

	a.OnTerminate().Add(func(ctx context.Context, e *OnTerminateEvent) error {
		a.store.Close()

		return nil
	})

	a.OnAfterBootstrap().Trigger(ctx, &OnAfterBootstrapEvent{
		App: a,
	})

	return nil
}
