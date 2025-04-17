package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/tuongaz/go-saas/store/types"
)

// collection represents a database table and provides methods to interact with it
type collection struct {
	table string
	db    dbInterface
	store Interface
}

// NewCollection creates a new collection
func NewCollection(table string, db dbInterface, store Interface) CollectionInterface {
	if !ValidTableName(table) {
		panic(fmt.Sprintf("invalid table name: %s", table))
	}

	return &collection{
		table: table,
		db:    db,
		store: store,
	}
}

// handleDBError tries to convert a generic database error into a more specific error type
func handleDBError(err error) error {
	if err == nil {
		return nil
	}

	// Check for SQL "no rows" error
	if errors.Is(err, sql.ErrNoRows) {
		return NewNotFoundErr(err)
	}

	// Convert to string to check for specific Postgres error types
	errStr := err.Error()

	// Check for unique constraint violation (Postgres code 23505)
	if strings.Contains(errStr, "23505") {
		return NewDuplicateKeyErr(err)
	}

	// Check for foreign key constraint violation (Postgres code 23503)
	if strings.Contains(errStr, "23503") {
		return NewForeignKeyErr(err)
	}

	// Check for not null constraint violation (Postgres code 23502)
	if strings.Contains(errStr, "23502") {
		return NewNotNullErr(err)
	}

	// Return the original error if we don't recognize it
	return err
}

func (c *collection) Table() string {
	return c.table
}

// CreateRecord creates a new record and handles specific errors
func (c *collection) CreateRecord(ctx context.Context, record types.Record) (*types.Record, error) {
	if err := c.store.OnBeforeRecordCreated(ctx, c.table, record); err != nil {
		return nil, fmt.Errorf("before create event handler error: %w", err)
	}

	keys, values, placeholders, err := record.PrepareForDB()
	if err != nil {
		return nil, fmt.Errorf("prepare record for database insertion: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", c.table, strings.Join(keys, ", "), strings.Join(placeholders, ", "))
	_, err = c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, handleDBError(err)
	}

	if err := c.store.OnAfterRecordCreated(ctx, c.table, record); err != nil {
		return nil, fmt.Errorf("after create event handler error: %w", err)
	}

	return &record, nil
}

// GetRecord retrieves a record by its id
func (c *collection) GetRecord(ctx context.Context, id any) (*types.Record, error) {
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

	var rec types.Record = make(map[string]any)
	if err = rows.MapScan(rec); err != nil {
		return nil, fmt.Errorf("failed to scan record into map: %v", err)
	}
	rec.Normalise()

	return &rec, nil
}

// UpdateRecord updates a record by its id and returns the updated record
func (c *collection) UpdateRecord(ctx context.Context, id any, record types.Record) (*types.Record, error) {
	// Get the old record for the event
	oldRecord, err := c.GetRecord(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get old record: %w", err)
	}

	if err := c.store.OnBeforeRecordUpdated(ctx, c.table, record, *oldRecord); err != nil {
		return nil, fmt.Errorf("before update event handler error: %w", err)
	}

	keys, values, _, err := record.PrepareForDB()
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
		return nil, handleDBError(err)
	}

	// Get the updated record
	updatedRecord, err := c.GetRecord(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated record: %w", err)
	}

	if err := c.store.OnAfterRecordUpdated(ctx, c.table, *updatedRecord, *oldRecord); err != nil {
		return nil, fmt.Errorf("after update event handler error: %w", err)
	}

	return updatedRecord, nil
}

// Update updates records based on the provided record and conditions
func (c *collection) Update(ctx context.Context, record types.Record, args ...any) (int64, error) {
	keys, values, _, err := record.PrepareForDB()
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
		return 0, handleDBError(err)
	}

	// Return the number of affected rows
	return result.RowsAffected()
}

// DeleteRecord deletes a record by its id
func (c *collection) DeleteRecord(ctx context.Context, id any) error {
	// Get the record before deletion for the event
	record, err := c.GetRecord(ctx, id)
	if err != nil {
		return fmt.Errorf("get record before deletion: %w", err)
	}

	if err := c.store.OnBeforeRecordDeleted(ctx, c.table, *record); err != nil {
		return fmt.Errorf("before delete event handler error: %w", err)
	}

	query := "DELETE FROM " + c.table + " WHERE id = $1"
	_, err = c.db.ExecContext(ctx, query, id)
	if err != nil {
		return handleDBError(err)
	}

	if err := c.store.OnAfterRecordDeleted(ctx, c.table, *record); err != nil {
		return fmt.Errorf("after delete event handler error: %w", err)
	}

	return nil
}

// DeleteRecords deletes records matching the filter
func (c *collection) DeleteRecords(ctx context.Context, filter Filter) error {
	query, args := buildQuery("DELETE FROM "+c.table, filter)

	_, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	return nil
}

// FindOne retrieves a single record matching the filter
func (c *collection) FindOne(ctx context.Context, filter Filter) (*types.Record, error) {
	var rec types.Record = make(map[string]interface{})
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

	rec.Normalise()

	return &rec, nil
}

// Find fetches records using the options pattern
func (c *collection) Find(ctx context.Context, opts ...FindOption) (*List, error) {
	// Set default options
	options := &FindOptions{
		Filter: Filter{},
	}

	// Apply all option functions
	for _, opt := range opts {
		opt(options)
	}

	var query string
	var args []any

	// Determine which filter to use
	if options.AdvancedFilter != nil {
		// Use advanced filter
		baseQuery := "SELECT"

		// Handle field selection
		if len(options.Fields) > 0 {
			// Validate field names
			for _, field := range options.Fields {
				if !ValidIdentifierName(field) {
					return nil, fmt.Errorf("invalid field name: %s", field)
				}
			}
			baseQuery += " " + strings.Join(options.Fields, ", ")
		} else {
			baseQuery += " *"
		}

		baseQuery += " FROM " + c.table
		query, args = buildAdvancedQuery(baseQuery, *options.AdvancedFilter)
	} else {
		// Use simple filter
		baseQuery := "SELECT"

		// Handle field selection
		if len(options.Fields) > 0 {
			// Validate field names
			for _, field := range options.Fields {
				if !ValidIdentifierName(field) {
					return nil, fmt.Errorf("invalid field name: %s", field)
				}
			}
			baseQuery += " " + strings.Join(options.Fields, ", ")
		} else {
			baseQuery += " *"
		}

		baseQuery += " FROM " + c.table
		query, args = buildQuery(baseQuery, options.Filter)
	}

	// Get total count for metadata
	var totalCount int
	var err error
	if options.AdvancedFilter != nil {
		countQuery, countArgs := buildAdvancedQuery("SELECT COUNT(*) FROM "+c.table, *options.AdvancedFilter)
		err = c.db.GetContext(ctx, &totalCount, countQuery, countArgs...)
	} else {
		totalCount, err = c.Count(ctx, options.Filter)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Apply sorting
	if len(options.Sort) > 0 {
		sortClauses := make([]string, 0, len(options.Sort))
		for _, opt := range options.Sort {
			// Validate field name to prevent SQL injection
			if !ValidIdentifierName(opt.Field) {
				return nil, fmt.Errorf("invalid field name for sorting: %s", opt.Field)
			}

			// Validate sort direction
			if opt.Direction != SortAsc && opt.Direction != SortDesc {
				opt.Direction = SortAsc // Default to ascending if invalid
			}

			sortClauses = append(sortClauses, fmt.Sprintf("%s %s", opt.Field, opt.Direction))
		}

		if len(sortClauses) > 0 {
			query += " ORDER BY " + strings.Join(sortClauses, ", ")
		}
	}

	// Apply pagination if provided
	meta := Metadata{
		Total: totalCount,
	}

	if options.Pagination != nil {
		meta.Limit = options.Pagination.Limit
		meta.Offset = options.Pagination.Offset
		meta.TotalPages = int(math.Ceil(float64(totalCount) / float64(options.Pagination.Limit)))

		if options.Pagination.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", options.Pagination.Limit)
		}

		if options.Pagination.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", options.Pagination.Offset)
		}
	}

	// Execute the query
	rows, err := c.db.QueryxContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewNotFoundErr(err)
		}
		return nil, handleDBError(err)
	}
	defer rows.Close()

	// Process the results
	var recs []types.Record
	for rows.Next() {
		rec := make(types.Record)
		if err := rows.MapScan(rec); err != nil {
			return nil, fmt.Errorf("failed to scan record into map: %v", err)
		}
		rec.Normalise()

		recs = append(recs, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing rows: %v", err)
	}

	return &List{
		Records: recs,
		Meta:    meta,
	}, nil
}

// Count returns the number of records matching the filter
func (c *collection) Count(ctx context.Context, filter Filter) (int, error) {
	var count int
	query, args := buildQuery("SELECT COUNT(*) FROM "+c.table, filter)
	err := c.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Exists checks if any records match the filter
func (c *collection) Exists(ctx context.Context, filter Filter) (bool, error) {
	count, err := c.Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
