package store

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tuongaz/go-saas/config"
)

type Store struct {
	db *sqlx.DB
}

func New(cfg config.Interface) (*Store, error) {
	datasource := cfg.GetPostgresDataSource()
	dbName := extractDBName(datasource)
	if dbName == "" {
		return nil, fmt.Errorf("dbname required")
	}

	db, err := sqlx.Connect("postgres", datasource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
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

func extractDBName(connStr string) string {
	parts := strings.Split(connStr, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname=")
		}
	}
	return ""
}
