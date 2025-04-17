package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tuongaz/go-saas/store/events"
	"github.com/tuongaz/go-saas/store/types"
)

var _ Interface = (*Store)(nil)

type dbInterface interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
}

type Interface interface {
	Collection(table string) CollectionInterface
	Exec(ctx context.Context, query string, args ...any) error
	Tx(ctx context.Context) (*StoreTx, error)
	DB() *sqlx.DB
	Close() error

	// Event handlers
	AddEventHandler(handler events.Handler)

	// Database events
	OnBeforeRecordCreated(ctx context.Context, table string, record types.Record) error
	OnAfterRecordCreated(ctx context.Context, table string, record types.Record) error
	OnBeforeRecordUpdated(ctx context.Context, table string, record types.Record, oldRecord types.Record) error
	OnAfterRecordUpdated(ctx context.Context, table string, record types.Record, oldRecord types.Record) error
	OnBeforeRecordDeleted(ctx context.Context, table string, record types.Record) error
	OnAfterRecordDeleted(ctx context.Context, table string, record types.Record) error
}

type Store struct {
	db       *sqlx.DB
	handlers []events.Handler
}

func New(datasource string) (*Store, error) {
	db, err := sqlx.Connect("postgres", datasource)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	return &Store{
		db:       db,
		handlers: make([]events.Handler, 0),
	}, nil
}

func (s *Store) Collection(table string) CollectionInterface {
	if !ValidTableName(table) {
		panic(fmt.Sprintf("invalid table name: %s", table))
	}

	return NewCollection(table, s.db, s)
}

func (s *Store) Exec(ctx context.Context, query string, args ...interface{}) error {
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	return nil
}

func (s *Store) DB() *sqlx.DB {
	return s.db
}

func (s *Store) Tx(ctx context.Context) (*StoreTx, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	return &StoreTx{
		tx:    tx,
		store: s,
	}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}

	return nil
}

// QueryValue executes a raw SQL query and scans the result into dest
func (s *Store) QueryValue(ctx context.Context, query string, dest any, args ...any) error {
	return s.db.GetContext(ctx, dest, query, args...)
}

// Query executes a raw SQL query and returns multiple records
func (s *Store) Query(ctx context.Context, query string, args ...any) (*List, error) {
	return s.QueryBuilder().
		WithQuery(query).
		WithArgs(args...).
		Execute(ctx, s)
}

// QueryWithPagination executes a SQL query with pagination support
func (s *Store) QueryWithPagination(ctx context.Context, query string, pagination Pagination, args ...any) (*List, error) {
	return s.QueryBuilder().
		WithQuery(query).
		WithArgs(args...).
		WithPagination(pagination).
		Execute(ctx, s)
}

// QueryOne executes a raw SQL query and returns a single record
func (s *Store) QueryOne(ctx context.Context, query string, args ...any) (*types.Record, error) {
	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error reading rows: %v", err)
		}
		return nil, NewNotFoundErr(fmt.Errorf("no records found"))
	}

	rec := make(types.Record)
	if err = rows.MapScan(rec); err != nil {
		return nil, fmt.Errorf("failed to scan record into map: %v", err)
	}
	rec.Normalise()

	return &rec, nil
}

// AddEventHandler adds an event handler to the store
func (s *Store) AddEventHandler(handler events.Handler) {
	s.handlers = append(s.handlers, handler)
}

// Database events
func (s *Store) OnBeforeRecordCreated(ctx context.Context, table string, record types.Record) error {
	event := &events.OnBeforeRecordCreatedEvent{
		DatabaseEvent: events.DatabaseEvent{
			Table:  table,
			Record: record,
		},
	}
	for _, handler := range s.handlers {
		if err := handler.OnBeforeCreate(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) OnAfterRecordCreated(ctx context.Context, table string, record types.Record) error {
	event := &events.OnAfterRecordCreatedEvent{
		DatabaseEvent: events.DatabaseEvent{
			Table:  table,
			Record: record,
		},
	}
	for _, handler := range s.handlers {
		if err := handler.OnAfterCreate(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) OnBeforeRecordUpdated(ctx context.Context, table string, record types.Record, oldRecord types.Record) error {
	event := &events.OnBeforeRecordUpdatedEvent{
		DatabaseEvent: events.DatabaseEvent{
			Table:  table,
			Record: record,
		},
		OldRecord: oldRecord,
	}
	for _, handler := range s.handlers {
		if err := handler.OnBeforeUpdate(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) OnAfterRecordUpdated(ctx context.Context, table string, record types.Record, oldRecord types.Record) error {
	event := &events.OnAfterRecordUpdatedEvent{
		DatabaseEvent: events.DatabaseEvent{
			Table:  table,
			Record: record,
		},
		OldRecord: oldRecord,
	}
	for _, handler := range s.handlers {
		if err := handler.OnAfterUpdate(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) OnBeforeRecordDeleted(ctx context.Context, table string, record types.Record) error {
	event := &events.OnBeforeRecordDeletedEvent{
		DatabaseEvent: events.DatabaseEvent{
			Table:  table,
			Record: record,
		},
	}
	for _, handler := range s.handlers {
		if err := handler.OnBeforeDelete(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) OnAfterRecordDeleted(ctx context.Context, table string, record types.Record) error {
	event := &events.OnAfterRecordDeletedEvent{
		DatabaseEvent: events.DatabaseEvent{
			Table:  table,
			Record: record,
		},
	}
	for _, handler := range s.handlers {
		if err := handler.OnAfterDelete(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
