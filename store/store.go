package store

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/pkg/log"
)

type Store struct {
	db *sqlx.DB
}

func New(cfg config.Interface) (*Store, error) {
	datasource := cfg.GetPostgresDataSource()
	dbName := extractDBName(datasource)
	if dbName == "" {
		dbName = "gosaas"
		datasource = fmt.Sprintf("%s dbname=gosaas", datasource)
	}

	db, err := sqlx.Connect("postgres", datasource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
	err = db.Get(&exists, query)
	if err != nil {
		return nil, fmt.Errorf("failed to check database exists: %w", err)
	}

	if !exists {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		log.Info("Database created.", "database", dbName)
	} else {
		log.Info("Database already exists.", "database", dbName)
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
