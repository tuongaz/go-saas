package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errors2 "github.com/tuongaz/go-saas/pkg/errors"
	"github.com/tuongaz/go-saas/store"
)

func (s *SQL) namedExecContext(
	ctx context.Context,
	query string,
	row any,
) (sql.Result, error) {
	result, err := s.conn.NamedExecContext(
		ctx,
		query,
		row,
	)
	if err != nil {
		return nil, store.NewDBError(fmt.Errorf("named exec context: %w", err))
	}

	return result, nil
}

func (s *SQL) getContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	err := s.conn.GetContext(
		ctx,
		dest,
		query,
		args...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors2.NewNotFoundErr(fmt.Errorf("get context: %w", err))
		}

		return store.NewDBError(fmt.Errorf("get context: %w", err))
	}

	return nil
}

func (s *SQL) selectContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	err := s.conn.SelectContext(
		ctx,
		dest,
		query,
		args...,
	)
	if err != nil {
		return store.NewDBError(fmt.Errorf("select context: %w", err))
	}

	return nil
}
