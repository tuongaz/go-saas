package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	errors2 "github.com/autopus/bootstrap/pkg/errors"
	"github.com/autopus/bootstrap/store/persistence"
)

var _ persistence.Interface = (*SQL)(nil)

func New(datasource string) (*SQL, func(), error) {
	conn, err := sqlx.Connect("sqlite3", datasource)
	if err != nil {
		return nil, nil, err
	}

	return &SQL{
			conn: conn,
		}, func() {
			_ = conn.Close()
		}, nil
}

type SQL struct {
	conn *sqlx.DB
}

func (s *SQL) Conn() *sqlx.DB {
	return s.conn
}

func (s *SQL) DBExists() bool {
	// TODO: better way to check if db exists
	var rows []persistence.AccountRow
	err := s.conn.Select(&rows, "SELECT * FROM account LIMIT 1")
	if err != nil {
		return false
	}

	return true
}

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
		return nil, persistence.NewDBError(fmt.Errorf("named exec context: %w", err))
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

		return persistence.NewDBError(fmt.Errorf("get context: %w", err))
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
		return persistence.NewDBError(fmt.Errorf("select context: %w", err))
	}

	return nil
}
