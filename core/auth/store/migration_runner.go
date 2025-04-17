package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// RunMigrations runs all SQL migration files in the migrations directory
func RunMigrations(ctx context.Context, db *sqlx.DB) error {
	// Get all migration files
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		// If directory doesn't exist, create it and return
		if os.IsNotExist(err) {
			if err := os.MkdirAll(migrationDir, 0755); err != nil {
				return fmt.Errorf("failed to create migrations directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Get SQL files and sort them
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Create migrations table if it doesn't exist
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Execute each migration in order
	for _, fileName := range sqlFiles {
		// Check if migration has already been applied
		var count int
		err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM migrations WHERE name = $1", fileName)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", fileName, err)
		}

		if count > 0 {
			// Migration already applied, skip it
			continue
		}

		// Read and execute the migration
		filePath := filepath.Join(migrationDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		// Execute within a transaction
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", fileName, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}

		// Record the migration
		if _, err := tx.ExecContext(ctx, "INSERT INTO migrations(name) VALUES($1)", fileName); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %w", fileName, err)
		}

		fmt.Printf("Applied migration: %s\n", fileName)
	}

	return nil
}
