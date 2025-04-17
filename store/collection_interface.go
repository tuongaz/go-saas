package store

import (
	"context"

	"github.com/tuongaz/go-saas/store/types"
)

var _ CollectionInterface = &collection{}

// CollectionInterface defines the methods available for interacting with a database collection
type CollectionInterface interface {
	// Table returns the name of the table
	Table() string

	// CreateRecord creates a new record and handles specific errors
	CreateRecord(ctx context.Context, record types.Record) (*types.Record, error)

	// GetRecord retrieves a record by its id
	GetRecord(ctx context.Context, id any) (*types.Record, error)

	// UpdateRecord updates a record by its id and returns the updated record
	UpdateRecord(ctx context.Context, id any, record types.Record) (*types.Record, error)

	// Update updates records based on the provided record and conditions
	Update(ctx context.Context, record types.Record, args ...any) (int64, error)

	// DeleteRecord deletes a record by its id
	DeleteRecord(ctx context.Context, id any) error

	// DeleteRecords deletes records matching the filter
	DeleteRecords(ctx context.Context, filter Filter) error

	// FindOne retrieves a single record matching the filter
	FindOne(ctx context.Context, filter Filter) (*types.Record, error)

	// Find fetches records using the options pattern
	Find(ctx context.Context, opts ...FindOption) (*List, error)

	// Count returns the number of records matching the filter
	Count(ctx context.Context, filter Filter) (int, error)

	// Exists checks if any records match the filter
	Exists(ctx context.Context, filter Filter) (bool, error)
}
