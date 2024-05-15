package store

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/pkg/log"
)

type Store struct {
	db *sqlx.DB
}

func New(cfg config.Interface) (*Store, error) {
	db, err := sqlx.Connect("postgres", cfg.GetPostgresDataSource())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", cfg.GetDBName())
	err = db.Get(&exists, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check database exists: %w", err)
	}

	if !exists {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.GetDBName()))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		log.Info("Database created.", "database", cfg.GetDBName())
	} else {
		log.Info("Database already exists.", "database", cfg.GetDBName())
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) DB() *sqlx.DB {
	return s.db
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}

	return nil
}

type DBError struct {
	Err error
}

func NewDBError(err error) *DBError {
	return &DBError{Err: err}
}

func (e *DBError) Error() string {
	return e.Err.Error()
}
