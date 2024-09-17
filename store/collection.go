package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Filter map[string]any

type Record map[string]any

func (r Record) normalise() {
	// Handle special cases like JSONB and timestamps
	for key, value := range r {
		switch v := value.(type) {
		case []byte:
			var jsonData any
			if err := json.Unmarshal(v, &jsonData); err == nil {
				r[key] = jsonData
			}
		case time.Time:
			// Convert time.Time to string for consistency
			r[key] = v.Format(time.RFC3339)
		}
	}
}

func (r Record) prepareForDB() (keys []string, values []any, placeholders []string, err error) {
	keys = make([]string, 0, len(r))
	values = make([]any, 0, len(r))
	placeholders = make([]string, 0, len(r))

	for k, v := range r {
		keys = append(keys, k)

		switch value := v.(type) {
		case string, int, int64, float64, bool, nil:
			// These types can be inserted directly
			values = append(values, v)
		case *string, *int, *int64, *float64, *bool:
			// Handle all pointer types together
			rv := reflect.ValueOf(value)
			if rv.IsNil() {
				values = append(values, nil)
			} else {
				values = append(values, rv.Elem().Interface())
			}
		default:
			// For any other type, marshal to JSON
			jsonBytes, marshalErr := json.Marshal(value)
			if marshalErr != nil {
				return nil, nil, nil, fmt.Errorf("failed to marshal value to JSON for key %s: %w", k, marshalErr)
			}
			values = append(values, string(jsonBytes))
		}

		placeholders = append(placeholders, fmt.Sprintf("$%d", len(placeholders)+1))
	}

	return keys, values, placeholders, nil
}

func (r Record) Get(key string) any {
	return r[key]
}

func (r Record) Decode(obj any) error {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("encode to json: %w", err)
	}
	if err := json.Unmarshal(jsonData, obj); err != nil {
		return fmt.Errorf("decode json to struct: %w", err)
	}
	return nil
}

type Collection struct {
	table string
	db    dbInterface
}

func (c *Collection) CreateRecord(ctx context.Context, record Record) (*Record, error) {
	keys, values, placeholders, err := record.prepareForDB()
	if err != nil {
		return nil, fmt.Errorf("prepare record for database insertion: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", c.table, strings.Join(keys, ", "), strings.Join(placeholders, ", "))
	if _, err := c.db.ExecContext(ctx, query, values...); err != nil {
		return nil, err
	}

	return &record, nil
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
	rec.normalise()

	return &rec, nil
}

func (c *Collection) UpdateRecord(ctx context.Context, id any, record Record) (*Record, error) {
	keys, values, _, err := record.prepareForDB()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare record for database update: %w", err)
	}

	setStatements := make([]string, len(keys))
	for i, key := range keys {
		setStatements[i] = fmt.Sprintf("%s = $%d", key, i+1)
	}

	// Add the ID to the parameters list for the WHERE clause
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d",
		c.table,
		strings.Join(setStatements, ", "),
		len(values))

	_, err = c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}

	return c.GetRecord(ctx, id)
}

func (c *Collection) Update(ctx context.Context, record Record, args ...any) (int64, error) {
	keys, values, _, err := record.prepareForDB()
	if err != nil {
		return 0, fmt.Errorf("failed to prepare record for database update: %w", err)
	}

	setStatements := make([]string, len(keys))
	for i, key := range keys {
		setStatements[i] = fmt.Sprintf("%s = $%d", key, i+1)
	}

	// Prepare the base query
	query := fmt.Sprintf("UPDATE %s SET %s", c.table, strings.Join(setStatements, ", "))

	// Process additional arguments (WHERE conditions)
	if len(args) > 0 {
		if len(args)%2 != 0 {
			return 0, fmt.Errorf("invalid number of arguments: must be key-value pairs")
		}

		whereConditions := make([]string, 0, len(args)/2)
		for i := 0; i < len(args); i += 2 {
			key, ok := args[i].(string)
			if !ok {
				return 0, fmt.Errorf("argument %d must be a string (column name)", i)
			}
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", key, len(values)+1))
			values = append(values, args[i+1])
		}

		if len(whereConditions) > 0 {
			query += " WHERE " + strings.Join(whereConditions, " AND ")
		}
	}

	// Execute the query
	result, err := c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update query: %w", err)
	}

	// Return the number of affected rows
	return result.RowsAffected()
}

func (c *Collection) DeleteRecord(ctx context.Context, id any) error {
	query := "DELETE FROM " + c.table + " WHERE id = $1"
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}

func (c *Collection) DeleteRecords(ctx context.Context, filter Filter) error {
	query, args := buildQuery("DELETE FROM "+c.table, filter)

	_, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	return nil
}

func (c *Collection) FindOne(ctx context.Context, filter Filter) (*Record, error) {
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

	rec.normalise()

	return &rec, nil
}

func (c *Collection) Find(ctx context.Context, filter Filter) (*List, error) {
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
		rec := make(Record)
		if err := rows.MapScan(rec); err != nil {
			return nil, fmt.Errorf("failed to scan record into map: %v", err)
		}
		rec.normalise()

		recs = append(recs, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %v", err)
	}

	return &List{
		Records: recs,
	}, nil
}

func (c *Collection) Count(ctx context.Context, filter Filter) (int, error) {
	var count int
	query, args := buildQuery("SELECT COUNT(*) FROM "+c.table, filter)
	err := c.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Collection) Exists(ctx context.Context, filter Filter) (bool, error) {
	count, err := c.Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// buildQuery helps in constructing SQL query strings based on a filter which can be a string or a map[string]any.
func buildQuery(baseQuery string, filter Filter) (string, []any) {
	if len(filter) == 0 {
		return baseQuery, nil
	}

	var parts []string
	var args []any
	i := 1

	for k, v := range filter {
		if v == nil {
			parts = append(parts, fmt.Sprintf("%s IS NULL", k))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	return baseQuery + " WHERE " + strings.Join(parts, " AND "), args
}

type List struct {
	Records []Record
}

func (r List) Decode(obj any) error {
	jsonData, err := json.Marshal(r.Records)
	if err != nil {
		return fmt.Errorf("encode to json: %w", err)
	}
	if err := json.Unmarshal(jsonData, obj); err != nil {
		return fmt.Errorf("decode json to struct: %w", err)
	}
	return nil
}
