package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
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

type dbInterface interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Collection struct {
	table string
	db    dbInterface
}

func (c *Collection) CreateRecord(ctx context.Context, record Record) (*Record, error) {
	columns := make([]string, 0, len(record))
	values := make([]string, 0, len(record))
	args := make([]any, 0, len(record))

	for k, v := range record {
		columns = append(columns, pq.QuoteIdentifier(k))
		values = append(values, fmt.Sprintf("$%d", len(values)+1))
		args = append(args, v)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) RETURNING *`,
		pq.QuoteIdentifier(c.table),
		strings.Join(columns, ", "),
		strings.Join(values, ", "))

	row := c.db.QueryRow(ctx, query, args...)

	var newRecord = Record{
		"id":               "",
		"name":             "",
		"provider":         "",
		"provider_user_id": "",
		"email":            "",
		"avatar":           "",
		"account_id":       "",
		"last_login":       "",
		"created_at":       "",
		"updated_at":       "",
	}
	err := row.
	if err != nil {
		return nil, err
	}

	return &newRecord, nil
}

func (c *Collection) GetRecord(ctx context.Context, id string) (*Record, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", pq.QuoteIdentifier(c.table))
	row := c.db.QueryRow(ctx, query, id)

	record := make(Record)
	err := row.Scan(&record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Collection) UpdateRecord(ctx context.Context, id string, record Record) (*Record, error) {
	setClauses := make([]string, 0, len(record))
	args := make([]any, 0, len(record)+1)

	for k, v := range record {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), len(args)+1))
		args = append(args, v)
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d RETURNING *",
		pq.QuoteIdentifier(c.table),
		strings.Join(setClauses, ", "),
		len(args))

	row := c.db.QueryRow(ctx, query, args...)

	var updatedRecord Record
	err := row.Scan(&updatedRecord)
	if err != nil {
		return nil, fmt.Errorf("update record (%s) %w", c.table, err)
	}

	return &updatedRecord, nil
}

func (c *Collection) DeleteRecord(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", pq.QuoteIdentifier(c.table))
	_, err := c.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete record (%s) %w", c.table, err)
	}
	return nil
}

func (c *Collection) FindOne(ctx context.Context, filter any) (*Record, error) {
	query, args := buildQuery(c.table, filter)

	record := make(Record)
	err := c.db.QueryRow(ctx, query, args...).Scan(&record)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewNotFoundErr(err)
		}
		return nil, fmt.Errorf("find one (%s) %w", c.table, err)
	}

	return &record, nil
}

func (c *Collection) Find(ctx context.Context, filter any) ([]Record, error) {
	query, args := buildQuery(c.table, filter)
	rows, err := c.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find : query (%s) %w", c.table, err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		record := make(Record)
		if err := rows.Scan(&record); err != nil {
			return nil, fmt.Errorf("find: scan (%s) %w", c.table, err)
		}
		records = append(records, record)
	}

	return records, nil
}

func (c *Collection) Count(ctx context.Context, filter any) (int, error) {
	whereClauses := []string{}
	args := []any{}

	switch f := filter.(type) {
	case string:
		whereClauses = append(whereClauses, f)
	case Record:
		for k, v := range f {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), len(args)+1))
			args = append(args, v)
		}
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(c.table))
	if len(whereClauses) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(whereClauses, " AND "))
	}

	var count int
	err := c.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count: query (%s) %w", c.table, err)
	}

	return count, nil
}

func (c *Collection) Exists(ctx context.Context, filter any) (bool, error) {
	count, err := c.Count(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("exists: count (%s) %w", c.table, err)
	}

	return count > 0, nil
}

func buildQuery(table string, filter any) (string, []any) {
	whereClauses := []string{}
	args := []any{}

	switch f := filter.(type) {
	case string:
		whereClauses = append(whereClauses, f)
	case Record:
		for k, v := range f {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), len(args)+1))
			args = append(args, v)
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", pq.QuoteIdentifier(table))
	if len(whereClauses) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(whereClauses, " AND "))
	}

	return query, args
}
