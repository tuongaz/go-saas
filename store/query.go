package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuongaz/go-saas/store/types"
)

// buildQuery helps in constructing SQL query strings based on a simple filter.
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

// buildAdvancedQuery constructs a SQL query with advanced filtering options
func buildAdvancedQuery(baseQuery string, filter AdvancedFilter) (string, []any) {
	if len(filter.Conditions) == 0 {
		return baseQuery, nil
	}

	var parts []string
	var args []any
	i := 1

	for _, condition := range filter.Conditions {
		switch condition.Op {
		case FilterOpIsNull:
			parts = append(parts, fmt.Sprintf("%s IS NULL", condition.Field))
		case FilterOpIsNotNull:
			parts = append(parts, fmt.Sprintf("%s IS NOT NULL", condition.Field))
		case FilterOpIn, FilterOpNotIn:
			// Handle array values for IN/NOT IN operations
			if values, ok := condition.Value.([]any); ok && len(values) > 0 {
				placeholders := make([]string, len(values))
				for j := range values {
					placeholders[j] = fmt.Sprintf("$%d", i)
					args = append(args, values[j])
					i++
				}
				parts = append(parts, fmt.Sprintf("%s %s (%s)",
					condition.Field,
					condition.Op,
					strings.Join(placeholders, ", ")))
			}
		default:
			parts = append(parts, fmt.Sprintf("%s %s $%d", condition.Field, condition.Op, i))
			args = append(args, condition.Value)
			i++
		}
	}

	return baseQuery + " WHERE " + strings.Join(parts, " AND "), args
}

// RawQueryOptions represents options for raw SQL queries
type RawQueryOptions struct {
	// Query is the SQL query string
	Query string
	// Args are the arguments for the query
	Args []any
	// Pagination settings for the query
	Pagination *Pagination
}

// QueryBuilder creates a new query options builder
func (s *Store) QueryBuilder() *RawQueryOptions {
	return &RawQueryOptions{
		Args: []any{},
	}
}

// WithQuery sets the SQL query
func (qo *RawQueryOptions) WithQuery(query string) *RawQueryOptions {
	qo.Query = query
	return qo
}

// WithArgs adds arguments to the query
func (qo *RawQueryOptions) WithArgs(args ...any) *RawQueryOptions {
	qo.Args = append(qo.Args, args...)
	return qo
}

// WithPagination adds pagination to the query
func (qo *RawQueryOptions) WithPagination(pagination Pagination) *RawQueryOptions {
	qo.Pagination = &pagination
	return qo
}

// Execute executes the query with the configured options
func (qo *RawQueryOptions) Execute(ctx context.Context, s *Store) (*List, error) {
	query := qo.Query

	// Apply pagination if provided
	if qo.Pagination != nil && qo.Pagination.Limit > 0 {
		query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, qo.Pagination.Limit, qo.Pagination.Offset)
	}

	rows, err := s.db.QueryxContext(ctx, query, qo.Args...)
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

	// Create metadata if pagination was used
	meta := Metadata{}
	if qo.Pagination != nil {
		meta.Limit = qo.Pagination.Limit
		meta.Offset = qo.Pagination.Offset

		// If pagination is used, try to get the total count
		// This assumes the query doesn't already have COUNT or GROUP BY
		// For complex queries, the user would need to handle counting separately
		if qo.Pagination.Limit > 0 {
			countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query",
				strings.Replace(qo.Query, "SELECT * FROM", "SELECT 1 FROM", 1))
			var total int
			err := s.QueryValue(ctx, countQuery, &total, qo.Args...)
			if err == nil {
				meta.Total = total
			}
		}
	}

	return &List{
		Records: recs,
		Meta:    meta,
	}, nil
}
