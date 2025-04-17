package core

import (
	"context"

	"github.com/tuongaz/go-saas/core/auth/model"
	authStore "github.com/tuongaz/go-saas/core/auth/store"
	"github.com/tuongaz/go-saas/store/events"
	"github.com/tuongaz/go-saas/store/types"
)

// DatabaseEvent represents a database operation event
type DatabaseEvent struct {
	Table  string
	Record types.Record
}

// OnBeforeRecordCreatedEvent is fired before a record is created
type OnBeforeRecordCreatedEvent struct {
	DatabaseEvent
}

// OnAfterRecordCreatedEvent is fired after a record is created
type OnAfterRecordCreatedEvent struct {
	DatabaseEvent
}

// OnBeforeRecordUpdatedEvent is fired before a record is updated
type OnBeforeRecordUpdatedEvent struct {
	DatabaseEvent
	OldRecord types.Record
}

// OnAfterRecordUpdatedEvent is fired after a record is updated
type OnAfterRecordUpdatedEvent struct {
	DatabaseEvent
	OldRecord types.Record
}

// OnBeforeRecordDeletedEvent is fired before a record is deleted
type OnBeforeRecordDeletedEvent struct {
	DatabaseEvent
}

// OnAfterRecordDeletedEvent is fired after a record is deleted
type OnAfterRecordDeletedEvent struct {
	DatabaseEvent
}

type OnBeforeOrganisationCreatedEvent struct {
	Organisation model.Organisation
}

type OnAfterOrganisationCreatedEvent struct {
	Organisation model.Organisation
}

type OnBeforeOrganisationUpdatedEvent struct {
	Organisation    model.Organisation
	OldOrganisation model.Organisation
}

type OnAfterOrganisationUpdatedEvent struct {
	Organisation    model.Organisation
	OldOrganisation model.Organisation
}

type OnBeforeOrganisationDeletedEvent struct {
	Organisation model.Organisation
}

type OnAfterOrganisationDeletedEvent struct {
	Organisation model.Organisation
}

type OnBeforeAccountCreatedEvent struct {
	Account model.Account
}

type OnAfterAccountCreatedEvent struct {
	Account model.Account
}

type OnBeforeAccountUpdatedEvent struct {
	Account    model.Account
	OldAccount model.Account
}

type OnAfterAccountUpdatedEvent struct {
	Account    model.Account
	OldAccount model.Account
}

type OnBeforeAccountDeletedEvent struct {
	Account model.Account
}

type OnAfterAccountDeletedEvent struct {
	Account model.Account
}

// DatabaseEventHandler defines the interface for database event handlers
type DatabaseEventHandler interface {
	OnBeforeRecordCreated(event *OnBeforeRecordCreatedEvent) error
	OnAfterRecordCreated(event *OnAfterRecordCreatedEvent) error
	OnBeforeRecordUpdated(event *OnBeforeRecordUpdatedEvent) error
	OnAfterRecordUpdated(event *OnAfterRecordUpdatedEvent) error
	OnBeforeRecordDeleted(event *OnBeforeRecordDeletedEvent) error
	OnAfterRecordDeleted(event *OnAfterRecordDeletedEvent) error
}

// AppDatabaseEventHandler implements the store.EventHandler interface and forwards events to the app's event system
type AppDatabaseEventHandler struct {
	app *App
}

// NewAppDatabaseEventHandler creates a new app database event handler
func NewAppDatabaseEventHandler(app *App) *AppDatabaseEventHandler {
	return &AppDatabaseEventHandler{
		app: app,
	}
}

// OnBeforeCreate is called before a record is created
func (h *AppDatabaseEventHandler) OnBeforeCreate(ctx context.Context, event *events.OnBeforeRecordCreatedEvent) error {
	if event.Table == authStore.TableOrganisation {
		org := model.Organisation{}
		if err := event.Record.Decode(&org); err != nil {
			return err
		}

		if err := h.app.onBeforeOrganisationCreated.Trigger(ctx, &OnBeforeOrganisationCreatedEvent{
			Organisation: org,
		}); err != nil {
			return err
		}
	}

	if event.Table == authStore.TableAccount {
		acc := model.Account{}
		if err := event.Record.Decode(&acc); err != nil {
			return err
		}

		if err := h.app.onBeforeAccountCreated.Trigger(ctx, &OnBeforeAccountCreatedEvent{
			Account: acc,
		}); err != nil {
			return err
		}
	}

	coreEvent := &OnBeforeRecordCreatedEvent{
		DatabaseEvent: DatabaseEvent{
			Table:  event.Table,
			Record: event.Record,
		},
	}
	return h.app.onBeforeRecordCreated.Trigger(ctx, coreEvent)
}

// OnAfterCreate is called after a record is created
func (h *AppDatabaseEventHandler) OnAfterCreate(ctx context.Context, event *events.OnAfterRecordCreatedEvent) error {
	if event.Table == authStore.TableOrganisation {
		org := model.Organisation{}
		if err := event.Record.Decode(&org); err != nil {
			return err
		}

		if err := h.app.onAfterOrganisationCreated.Trigger(ctx, &OnAfterOrganisationCreatedEvent{
			Organisation: org,
		}); err != nil {
			return err
		}
	}

	if event.Table == authStore.TableAccount {
		acc := model.Account{}
		if err := event.Record.Decode(&acc); err != nil {
			return err
		}

		if err := h.app.onAfterAccountCreated.Trigger(ctx, &OnAfterAccountCreatedEvent{
			Account: acc,
		}); err != nil {
			return err
		}
	}

	coreEvent := &OnAfterRecordCreatedEvent{
		DatabaseEvent: DatabaseEvent{
			Table:  event.Table,
			Record: event.Record,
		},
	}
	return h.app.onAfterRecordCreated.Trigger(ctx, coreEvent)
}

// OnBeforeUpdate is called before a record is updated
func (h *AppDatabaseEventHandler) OnBeforeUpdate(ctx context.Context, event *events.OnBeforeRecordUpdatedEvent) error {
	if event.Table == authStore.TableOrganisation {
		org := model.Organisation{}
		if err := event.Record.Decode(&org); err != nil {
			return err
		}

		oldOrg := model.Organisation{}
		if err := types.Record(event.OldRecord).Decode(&oldOrg); err != nil {
			return err
		}

		if err := h.app.onBeforeOrganisationUpdated.Trigger(ctx, &OnBeforeOrganisationUpdatedEvent{
			Organisation:    org,
			OldOrganisation: oldOrg,
		}); err != nil {
			return err
		}
	}

	if event.Table == authStore.TableAccount {
		acc := model.Account{}
		if err := event.Record.Decode(&acc); err != nil {
			return err
		}

		oldAcc := model.Account{}
		if err := types.Record(event.OldRecord).Decode(&oldAcc); err != nil {
			return err
		}

		if err := h.app.onBeforeAccountUpdated.Trigger(ctx, &OnBeforeAccountUpdatedEvent{
			Account:    acc,
			OldAccount: oldAcc,
		}); err != nil {
			return err
		}
	}

	coreEvent := &OnBeforeRecordUpdatedEvent{
		DatabaseEvent: DatabaseEvent{
			Table:  event.Table,
			Record: event.Record,
		},
		OldRecord: event.OldRecord,
	}
	return h.app.onBeforeRecordUpdated.Trigger(ctx, coreEvent)
}

// OnAfterUpdate is called after a record is updated
func (h *AppDatabaseEventHandler) OnAfterUpdate(ctx context.Context, event *events.OnAfterRecordUpdatedEvent) error {
	if event.Table == authStore.TableOrganisation {
		org := model.Organisation{}
		if err := event.Record.Decode(&org); err != nil {
			return err
		}

		oldOrg := model.Organisation{}
		if err := types.Record(event.OldRecord).Decode(&oldOrg); err != nil {
			return err
		}

		if err := h.app.onAfterOrganisationUpdated.Trigger(ctx, &OnAfterOrganisationUpdatedEvent{
			Organisation:    org,
			OldOrganisation: oldOrg,
		}); err != nil {
			return err
		}
	}

	if event.Table == authStore.TableAccount {
		acc := model.Account{}
		if err := event.Record.Decode(&acc); err != nil {
			return err
		}

		oldAcc := model.Account{}
		if err := types.Record(event.OldRecord).Decode(&oldAcc); err != nil {
			return err
		}

		if err := h.app.onAfterAccountUpdated.Trigger(ctx, &OnAfterAccountUpdatedEvent{
			Account:    acc,
			OldAccount: oldAcc,
		}); err != nil {
			return err
		}
	}

	coreEvent := &OnAfterRecordUpdatedEvent{
		DatabaseEvent: DatabaseEvent{
			Table:  event.Table,
			Record: event.Record,
		},
		OldRecord: event.OldRecord,
	}
	return h.app.onAfterRecordUpdated.Trigger(ctx, coreEvent)
}

// OnBeforeDelete is called before a record is deleted
func (h *AppDatabaseEventHandler) OnBeforeDelete(ctx context.Context, event *events.OnBeforeRecordDeletedEvent) error {
	if event.Table == authStore.TableOrganisation {
		org := model.Organisation{}
		if err := event.Record.Decode(&org); err != nil {
			return err
		}

		if err := h.app.onBeforeOrganisationDeleted.Trigger(ctx, &OnBeforeOrganisationDeletedEvent{
			Organisation: org,
		}); err != nil {
			return err
		}
	}

	if event.Table == authStore.TableAccount {
		acc := model.Account{}
		if err := event.Record.Decode(&acc); err != nil {
			return err
		}

		if err := h.app.onBeforeAccountDeleted.Trigger(ctx, &OnBeforeAccountDeletedEvent{
			Account: acc,
		}); err != nil {
			return err
		}
	}

	coreEvent := &OnBeforeRecordDeletedEvent{
		DatabaseEvent: DatabaseEvent{
			Table:  event.Table,
			Record: event.Record,
		},
	}
	return h.app.onBeforeRecordDeleted.Trigger(ctx, coreEvent)
}

// OnAfterDelete is called after a record is deleted
func (h *AppDatabaseEventHandler) OnAfterDelete(ctx context.Context, event *events.OnAfterRecordDeletedEvent) error {
	if event.Table == authStore.TableOrganisation {
		org := model.Organisation{}
		if err := event.Record.Decode(&org); err != nil {
			return err
		}

		if err := h.app.onAfterOrganisationDeleted.Trigger(ctx, &OnAfterOrganisationDeletedEvent{
			Organisation: org,
		}); err != nil {
			return err
		}
	}

	if event.Table == authStore.TableAccount {
		acc := model.Account{}
		if err := event.Record.Decode(&acc); err != nil {
			return err
		}

		if err := h.app.onAfterAccountDeleted.Trigger(ctx, &OnAfterAccountDeletedEvent{
			Account: acc,
		}); err != nil {
			return err
		}
	}

	coreEvent := &OnAfterRecordDeletedEvent{
		DatabaseEvent: DatabaseEvent{
			Table:  event.Table,
			Record: event.Record,
		},
	}
	return h.app.onAfterRecordDeleted.Trigger(ctx, coreEvent)
}
