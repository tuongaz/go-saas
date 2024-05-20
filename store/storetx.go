package store

import (
	"context"
	"fmt"
)

type dbxInterface interface {
	dbInterface
	Commit() error
	Rollback() error
}

type StoreTx struct {
	tx dbxInterface
}

func (s *StoreTx) Collection(table string) *Collection {
	return &Collection{
		table: table,
		db:    s.tx,
	}
}

func (s *StoreTx) Exec(ctx context.Context, query string, args ...interface{}) error {
	if _, err := s.tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec tx query: %w", err)
	}

	return nil
}

func (s *StoreTx) Commit() error {
	return s.tx.Commit()
}

func (s *StoreTx) Rollback() error {
	return s.tx.Rollback()
}
