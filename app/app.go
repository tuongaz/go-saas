package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/auth/signer"
	"github.com/autopus/bootstrap/pkg/encrypt"
	"github.com/autopus/bootstrap/pkg/hooks"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/server"
	"github.com/autopus/bootstrap/service/auth"
	"github.com/autopus/bootstrap/service/scheduler"
	"github.com/autopus/bootstrap/store"
)

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

	srv := server.New(a.Cfg)

	if err := a.OnBeforeServe().Trigger(
		ctx,
		&OnBeforeServeEvent{App: a, Server: srv},
	); err != nil {
		return fmt.Errorf("failed to trigger on before serve event: %w", err)
	}

	done := make(chan bool, 1)
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	go func() {
		if err := srv.Serve(); err != nil {
			log.Default().Error("failed to serve: %v", err)
		}

		done <- true
	}()

	<-done

	return a.OnTerminate().Trigger(
		ctx,
		&OnTerminateEvent{App: a},
		func(ctx context.Context, e *OnTerminateEvent) error {
			return e.App.Shutdown()
		},
	)
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
	if !a.Cfg.IsSchedulerServiceEnabled() {
		return nil
	}

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
	if !a.Cfg.IsAuthServiceEnabled() {
		return nil
	}

	encryptor := encrypt.New(a.Cfg.GetEncryptionKey())
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
