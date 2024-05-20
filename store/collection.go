package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Record map[string]any

func (r Record) Get(key string) any {
	return r[key]
}

func (r Record) Decode(obj any) error {
	if err := mapstructure.Decode(r, obj); err != nil {
		return fmt.Errorf("decode record: %w", err)
	}

	return nil
}

type Collection struct {
	table string
	db    dbInterface
}

func (c *Collection) CreateRecord(ctx context.Context, record Record) (*Record, error) {
	keys := make([]string, 0, len(record))
	values := make([]any, 0, len(record))
	placeholders := make([]string, 0, len(record))

	i := 1
	for k, v := range record {
		keys = append(keys, k)
		values = append(values, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", c.table, strings.Join(keys, ", "), strings.Join(placeholders, ", "))
	var id string
	err := c.db.GetContext(ctx, &id, query, values...)
	if err != nil {
		return nil, err
	}

	return c.GetRecord(ctx, id)
}

func (c *Collection) GetRecord(ctx context.Context, id any) (*Record, error) {
	query := "SELECT * FROM " + c.table + " WHERE id = $1 LIMIT 1"
	rows, err := c.db.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error reading rows: %v", err)
		}
		return nil, sql.ErrNoRows
	}

	var rec Record = make(map[string]any)
	if err = rows.MapScan(rec); err != nil {
		return nil, fmt.Errorf("failed to scan record into map: %v", err)
	}

	return &rec, nil
}

func (c *Collection) UpdateRecord(ctx context.Context, id any, record Record) (*Record, error) {
	keys := make([]string, 0, len(record))
	values := make([]any, 0, len(record))

	i := 1
	for k, v := range record {
		keys = append(keys, fmt.Sprintf("%s = $%d", k, i))
		values = append(values, v)
		i++
	}

	// Add the ID to the parameters list for the WHERE clause
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", c.table, strings.Join(keys, ", "), i)
	_, err := c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	return c.GetRecord(ctx, id)
}

func (c *Collection) DeleteRecord(ctx context.Context, id any) error {
	query := "DELETE FROM " + c.table + " WHERE id = $1"
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}

func (c *Collection) FindOne(ctx context.Context, filter any) (*Record, error) {
	var rec Record = make(map[string]interface{})
	query, args := buildQuery("SELECT * FROM "+c.table, filter)
	rows, err := c.db.QueryxContext(ctx, query+" LIMIT 1", args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundErr(err)
		}

		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error reading rows: %v", err)
		}
		return nil, NewNotFoundErr(fmt.Errorf("record not found"))
	}

	if err = rows.MapScan(rec); err != nil {
		return nil, fmt.Errorf("failed to scan record into map: %v", err)
	}

	return &rec, nil
}

func (c *Collection) Find(ctx context.Context, filter any) ([]Record, error) {
	query, args := buildQuery("SELECT * FROM "+c.table, filter)
	rows, err := c.db.QueryxContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundErr(err)
		}

		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	var recs []Record
	for rows.Next() {
		rec := make(map[string]interface{})
		if err := rows.MapScan(rec); err != nil {
			return nil, fmt.Errorf("failed to scan record into map: %v", err)
		}
		recs = append(recs, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %v", err)
	}

	return recs, nil
}

func (c *Collection) Count(ctx context.Context, filter any) (int, error) {
	var count int
	query, args := buildQuery("SELECT COUNT(*) FROM "+c.table, filter)
	err := c.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Collection) Exists(ctx context.Context, filter any) (bool, error) {
	count, err := c.Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// buildQuery helps in constructing SQL query strings based on a filter which can be a string or a map[string]any.
func buildQuery(baseQuery string, filter any) (string, []any) {
	switch f := filter.(type) {
	case string:
		// Directly use the string as SQL where clause.
		return baseQuery + " WHERE " + f, nil
	case map[string]any:
		if len(f) == 0 {
			return baseQuery, nil
		}
		// Construct where clause from map with PostgreSQL placeholders.
		var parts []string
		var args []any
		i := 1 // Start indexing from 1 for PostgreSQL placeholders
		for k, v := range f {
			parts = append(parts, fmt.Sprintf("%s = $%d", k, i))
			args = append(args, v)
			i++
		}
		return baseQuery + " WHERE " + strings.Join(parts, " AND "), args
	case Record:
		if len(f) == 0 {
			return baseQuery, nil
		}
		// Construct where clause from map with PostgreSQL placeholders.
		var parts []string
		var args []any
		i := 1 // Start indexing from 1 for PostgreSQL placeholders
		for k, v := range f {
			parts = append(parts, fmt.Sprintf("%s = $%d", k, i))
			args = append(args, v)
			i++
		}
		return baseQuery + " WHERE " + strings.Join(parts, " AND "), args
	default:
		return baseQuery, nil
	}
}
