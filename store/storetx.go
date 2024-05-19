package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ TxInterface = (*StoreTx)(nil)

type TxInterface interface {
	Collection(table string) *Collection
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type StoreTx struct {
	tx pgx.Tx
}

func newStoreTx(tx pgx.Tx) *StoreTx {
	return &StoreTx{
		tx: tx,
	}
}

func (s *StoreTx) Collection(table string) *Collection {
	return &Collection{
		table: table,
		db:    s.tx,
	}
}

func (s *StoreTx) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return s.tx.Exec(ctx, query, args...)
}

func (s *StoreTx) Commit(ctx context.Context) error {
	return s.tx.Commit(ctx)
}

func (s *StoreTx) Rollback(ctx context.Context) error {
	return s.tx.Rollback(ctx)
}
