package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/server"
)

var _ Interface = (*App)(nil)

type Interface interface {
	Config() config.Interface

	// OnBeforeBootstrap returns the hook that is triggered before the app is bootstrapped.
	OnBeforeBootstrap() *hooks.Hook[*OnBeforeBootstrapEvent]

	// OnAfterBootstrap returns the hook that is triggered after the app is bootstrapped.
	OnAfterBootstrap() *hooks.Hook[*OnAfterBootstrapEvent]

	// OnTerminate returns the hook that is triggered when the app is terminated.
	OnTerminate() *hooks.Hook[*OnTerminateEvent]

	// OnBeforeServe returns the hook that is triggered before the app starts serving.
	OnBeforeServe() *hooks.Hook[*OnBeforeServeEvent]
}

type App struct {
	cfg               config.Interface
	onBeforeBootstrap *hooks.Hook[*OnBeforeBootstrapEvent]
	onAfterBootstrap  *hooks.Hook[*OnAfterBootstrapEvent]
	onBeforeServe     *hooks.Hook[*OnBeforeServeEvent]
	onTerminate       *hooks.Hook[*OnTerminateEvent]
}

func New() (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App{
		cfg:               cfg,
		onBeforeBootstrap: &hooks.Hook[*OnBeforeBootstrapEvent]{},
		onAfterBootstrap:  &hooks.Hook[*OnAfterBootstrapEvent]{},
		onBeforeServe:     &hooks.Hook[*OnBeforeServeEvent]{},
		onTerminate:       &hooks.Hook[*OnTerminateEvent]{},
	}, nil
}

func (a *App) Config() config.Interface {
	return a.cfg
}

func (a *App) Start() error {
	ctx := context.Background()

	if err := a.bootstrap(ctx); err != nil {
		return fmt.Errorf("app bootstrap: %w", err)
	}

	srv := server.New(a.Config())

	if err := a.OnBeforeServe().Trigger(
		ctx,
		&OnBeforeServeEvent{App: a, Server: srv},
	); err != nil {
		return fmt.Errorf("failed to trigger on before serve event: %w", err)
	}

	srv.PrintRoutes()

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

	return nil
}

func (a *App) bootstrap(ctx context.Context) error {
	if err := a.OnBeforeBootstrap().Trigger(ctx, &OnBeforeBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("before bootstrap: %w", err)
	}

	if err := a.OnAfterBootstrap().Trigger(ctx, &OnAfterBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("after bootstrap: %w", err)
	}

	return nil
}
