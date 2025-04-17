package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuongaz/go-saas/store/types"
)

type dbxInterface interface {
	dbInterface
	Commit() error
	Rollback() error
}

type StoreTx struct {
	tx    dbxInterface
	store Interface
}

func (s *StoreTx) Collection(table string) *collection {
	if !ValidTableName(table) {
		panic(fmt.Sprintf("invalid table name: %s", table))
	}

	return &collection{
		table: table,
		db:    s.tx,
		store: s.store,
	}
}

func (s *StoreTx) Exec(ctx context.Context, query string, args ...interface{}) error {
	if _, err := s.tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec tx query: %w", err)
	}

	return nil
}

func (s *StoreTx) Commit() error {
	return s.tx.Commit()
}

func (s *StoreTx) Rollback() error {
	return s.tx.Rollback()
}

// Query executes a raw SQL query within a transaction and returns multiple records
func (s *StoreTx) Query(ctx context.Context, query string, args ...any) (*List, error) {
	rows, err := s.tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

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
	}, nil
}

// QueryBuilder creates a new query options builder for transactions
func (s *StoreTx) QueryBuilder() *RawQueryOptions {
	return &RawQueryOptions{
		Args: []any{},
	}
}

// QueryWithPagination executes a SQL query with pagination support within a transaction
func (s *StoreTx) QueryWithPagination(ctx context.Context, query string, pagination Pagination, args ...any) (*List, error) {
	// Apply pagination to the query
	paginatedQuery := query
	if pagination.Limit > 0 {
		paginatedQuery = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pagination.Limit, pagination.Offset)
	}

	// Execute the query
	list, err := s.Query(ctx, paginatedQuery, args...)
	if err != nil {
		return nil, err
	}

	// Add pagination metadata
	list.Meta.Limit = pagination.Limit
	list.Meta.Offset = pagination.Offset

	// Try to get total count if pagination is used
	if pagination.Limit > 0 {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query",
			strings.Replace(query, "SELECT * FROM", "SELECT 1 FROM", 1))
		var total int
		err := s.tx.GetContext(ctx, &total, countQuery, args...)
		if err == nil {
			list.Meta.Total = total
		}
	}

	return list, nil
}

func (s *StoreTx) QueryValue(ctx context.Context, query string, dest any, args ...any) error {
	return s.tx.GetContext(ctx, dest, query, args...)
}
