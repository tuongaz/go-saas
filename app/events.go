package app

import (
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/server"
)

type OnBeforeBootstrapEvent struct {
	App *App
}

type OnAfterBootstrapEvent struct {
	App *App
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
