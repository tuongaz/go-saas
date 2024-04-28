package app

import (
	"context"
	"fmt"

	"github.com/autopus/bootstrap/pkg/hooks"
	"github.com/autopus/bootstrap/server"
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
