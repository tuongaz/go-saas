package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tuongaz/go-saas/config"
)

var _ Interface = (*Store)(nil)

type Interface interface {
	Collection(table string) *Collection
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Tx(ctx context.Context) (*StoreTx, error)
}

type Store struct {
	db *pgxpool.Pool
}

func New(cfg config.Interface) (*Store, error) {
	datasource := cfg.GetPostgresDataSource()

	poolConfig, err := pgxpool.ParseConfig(datasource)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create pool: %w", err)
	}

	return &Store{
		db: pool,
	}, nil
}

func (s *Store) Collection(table string) *Collection {
	return &Collection{
		table: table,
		db:    s.db,
	}
}

func (s *Store) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return s.db.Exec(ctx, query, args...)
}

func (s *Store) Tx(ctx context.Context) (*StoreTx, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	return &StoreTx{
		tx: tx,
	}, nil
}

func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
