package store

import (
	"context"

	"github.com/tuongaz/go-saas/store/types"
)

// DatabaseEvent represents a database event
type DatabaseEvent struct {
	Context context.Context
	Table   string
	Record  types.Record
}

// OnBeforeRecordCreatedEvent represents an event that is triggered before a record is created
type OnBeforeRecordCreatedEvent struct {
	DatabaseEvent
}

// OnAfterRecordCreatedEvent represents an event that is triggered after a record is created
type OnAfterRecordCreatedEvent struct {
	DatabaseEvent
}

// OnBeforeRecordUpdatedEvent represents an event that is triggered before a record is updated
type OnBeforeRecordUpdatedEvent struct {
	DatabaseEvent
	OldRecord types.Record
}

// OnAfterRecordUpdatedEvent represents an event that is triggered after a record is updated
type OnAfterRecordUpdatedEvent struct {
	DatabaseEvent
	OldRecord types.Record
}

// OnBeforeRecordDeletedEvent represents an event that is triggered before a record is deleted
type OnBeforeRecordDeletedEvent struct {
	DatabaseEvent
}

// OnAfterRecordDeletedEvent represents an event that is triggered after a record is deleted
type OnAfterRecordDeletedEvent struct {
	DatabaseEvent
}

// EventHandler defines the interface for database event handlers
type EventHandler interface {
	OnBeforeCreate(ctx context.Context, event *OnBeforeRecordCreatedEvent) error
	OnAfterCreate(ctx context.Context, event *OnAfterRecordCreatedEvent) error
	OnBeforeUpdate(ctx context.Context, event *OnBeforeRecordUpdatedEvent) error
	OnAfterUpdate(ctx context.Context, event *OnAfterRecordUpdatedEvent) error
	OnBeforeDelete(ctx context.Context, event *OnBeforeRecordDeletedEvent) error
	OnAfterDelete(ctx context.Context, event *OnAfterRecordDeletedEvent) error
}
