package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/tuongaz/go-saas/config"
)

var _ Interface = (*Store)(nil)

type dbInterface interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
}

type Interface interface {
	Collection(table string) *Collection
	Exec(ctx context.Context, query string, args ...any) error
	Tx(ctx context.Context) (*StoreTx, error)
}

type Store struct {
	db *sqlx.DB
}

func New(cfg config.Interface) (*Store, error) {
	datasource := cfg.GetPostgresDataSource()

	db, err := sqlx.Connect("postgres", datasource)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) Collection(table string) *Collection {
	return &Collection{
		table: table,
		db:    s.db,
	}
}

func (s *Store) Exec(ctx context.Context, query string, args ...interface{}) error {
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	return nil
}

func (s *Store) Tx(ctx context.Context) (*StoreTx, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	return &StoreTx{
		tx: tx,
	}, nil
}

func (s *Store) Close() {
	if s.db != nil {
		_ = s.db.Close()
	}
}
