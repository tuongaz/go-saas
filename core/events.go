package core

import (
	"github.com/tuongaz/go-saas/pkg/hooks"
)

type OnBeforeBootstrapEvent struct {
	App *App
}

type OnAfterBootstrapEvent struct {
	App *App
}

type OnBeforeServeEvent struct {
	App *App
}

type OnTerminateEvent struct {
	App *App
}

type OnDatabaseReadyEvent struct {
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

func (a *App) OnDatabaseReady() *hooks.Hook[*OnDatabaseReadyEvent] {
	return a.onDatabaseReady
}
