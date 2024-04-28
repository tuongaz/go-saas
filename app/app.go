package app

import (
	"context"
	"fmt"

	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/auth/signer"
	"github.com/autopus/bootstrap/pkg/encrypt"
	"github.com/autopus/bootstrap/pkg/hooks"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/scheduler"
	"github.com/autopus/bootstrap/server"
	"github.com/autopus/bootstrap/service/auth"
	"github.com/autopus/bootstrap/store"
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

type OnBeforeServeEvent struct {
	App    *App
	Server *server.Server
}

type OnTerminateEvent struct {
	App *App
}

type OnBeforeSchedulerBootstrapEvent struct {
	App *App
}

type OnAfterSchedulerBootstrapEvent struct {
	App *App
}

type App struct {
	Cfg      config.Interface
	Store    store.Interface
	dbCloser func()

	onBeforeBootstrap          *hooks.Hook[*OnBeforeBootstrapEvent]
	onAfterBootstrap           *hooks.Hook[*OnAfterBootstrapEvent]
	onBeforeStoreBootstrap     *hooks.Hook[*OnBeforeStoreBootstrapEvent]
	onAfterStoreBootstrap      *hooks.Hook[*OnAfterStoreBootstrapEvent]
	onBeforeSchedulerBootstrap *hooks.Hook[*OnBeforeSchedulerBootstrapEvent]
	onAfterSchedulerBootstrap  *hooks.Hook[*OnAfterSchedulerBootstrapEvent]
	onBeforeServe              *hooks.Hook[*OnBeforeServeEvent]
	onTerminate                *hooks.Hook[*OnTerminateEvent]

	authSrv   *auth.Service
	scheduler *scheduler.Scheduler
}

func New(cfg config.Interface) *App {
	return &App{
		Cfg: cfg,

		onBeforeBootstrap:          &hooks.Hook[*OnBeforeBootstrapEvent]{},
		onAfterBootstrap:           &hooks.Hook[*OnAfterBootstrapEvent]{},
		onBeforeStoreBootstrap:     &hooks.Hook[*OnBeforeStoreBootstrapEvent]{},
		onAfterStoreBootstrap:      &hooks.Hook[*OnAfterStoreBootstrapEvent]{},
		onBeforeSchedulerBootstrap: &hooks.Hook[*OnBeforeSchedulerBootstrapEvent]{},
		onAfterSchedulerBootstrap:  &hooks.Hook[*OnAfterSchedulerBootstrapEvent]{},

		onBeforeServe: &hooks.Hook[*OnBeforeServeEvent]{},
		onTerminate:   &hooks.Hook[*OnTerminateEvent]{},
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
	log.Default().Warn("shutting down app")
	if a.dbCloser != nil {
		a.dbCloser()
	}

	return nil
}

func (a *App) Scheduler() *scheduler.Scheduler {
	return a.scheduler
}

func (a *App) OnBeforeBootstrap() *hooks.Hook[*OnBeforeBootstrapEvent] {
	return a.onBeforeBootstrap
}

func (a *App) OnAfterBootstrap() *hooks.Hook[*OnAfterBootstrapEvent] {
	return a.onAfterBootstrap
}

func (a *App) OnBeforeServe() *hooks.Hook[*OnBeforeServeEvent] {
	return a.onBeforeServe
}

func (a *App) OnTerminate() *hooks.Hook[*OnTerminateEvent] {
	return a.onTerminate
}

func (a *App) OnBeforeSchedulerBootstrap() *hooks.Hook[*OnBeforeSchedulerBootstrapEvent] {
	return a.onBeforeSchedulerBootstrap
}

func (a *App) OnAfterSchedulerBootstrap() *hooks.Hook[*OnAfterSchedulerBootstrapEvent] {
	return a.onAfterSchedulerBootstrap
}

func (a *App) BootstrapStore(ctx context.Context) error {
	if err := a.onBeforeStoreBootstrap.Trigger(ctx, &OnBeforeStoreBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("before store start: %w", err)
	}

	a.Store, a.dbCloser = store.MustNew(a.Cfg)

	if err := a.onAfterStoreBootstrap.Trigger(ctx, &OnAfterStoreBootstrapEvent{
		App:   a,
		Store: a.Store,
	}); err != nil {
		return fmt.Errorf("after store start: %w", err)
	}

	return nil
}

func (a *App) GetAuthService() *auth.Service {
	return a.authSrv
}

func (a *App) bootstrap(ctx context.Context) error {
	if err := a.OnBeforeBootstrap().Trigger(ctx, &OnBeforeBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("before bootstrap: %w", err)
	}

	if err := a.BootstrapStore(ctx); err != nil {
		return fmt.Errorf("store bootstrap: %w", err)
	}

	if err := a.bootstrapScheduler(); err != nil {
		return fmt.Errorf("scheduler bootstrap: %w", err)
	}

	if err := a.bootstrapAuthService(); err != nil {
		return fmt.Errorf("auth service bootstrap: %w", err)
	}

	if err := a.OnAfterBootstrap().Trigger(ctx, &OnAfterBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("after bootstrap: %w", err)
	}

	return nil
}

func (a *App) bootstrapScheduler() error {
	if err := a.OnBeforeSchedulerBootstrap().Trigger(context.Background(), &OnBeforeSchedulerBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("before scheduler bootstrap: %w", err)
	}

	s, err := scheduler.New()
	if err != nil {
		return fmt.Errorf("scheduler bootstrap: %w", err)
	}

	a.scheduler = s
	s.Start()

	if err := a.OnAfterSchedulerBootstrap().Trigger(context.Background(), &OnAfterSchedulerBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("after scheduler bootstrap: %w", err)
	}

	return nil
}

func (a *App) bootstrapAuthService() error {
	encryptor := encrypt.New(a.Cfg.GetEncryptionKey())
	// setup authentication service
	authSrv, err := auth.New(
		a.Cfg,
		a.Store,
		encryptor,
		signer.NewHS256Signer([]byte(a.Cfg.GetJWTSigningSecret())),
		auth.WithJWTLifeTimeMinutes(a.Cfg.GetJWTTokenLifetimeMinutes()),
	)

	if err != nil {
		return fmt.Errorf("failed to init auth service: %w", err)
	}

	a.authSrv = authSrv

	return nil
}
