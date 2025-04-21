package store

// QueryOptions represents the legacy options struct (kept for backward compatibility)
type QueryOptions struct {
	Fields     []string     // Fields to select (defaults to all fields if empty)
	Sort       []SortOption // Sort options
	Pagination *Pagination  // Pagination options
}

// FindOption is a function that modifies FindOptions.
// This type implements the functional options pattern for configuring
// query parameters in a clean and extensible way.
type FindOption func(*FindOptions)

// FindOptions holds all possible query options for Find method.
// This struct is meant to be configured through FindOption functions,
// not constructed directly.
type FindOptions struct {
	Filter         Filter          // Simple equality-based filter
	AdvancedFilter *AdvancedFilter // Advanced filter with custom operators
	Fields         []string        // Fields to select (defaults to all fields if empty)
	Sort           []SortOption    // Sort options
	Pagination     *Pagination     // Pagination options
}

// WithFilter sets the simple filter option for equality-based filtering.
// Filter is a map where keys are column names and values are the values to match.
// Multiple conditions are combined with AND logic.
//
// Example: WithFilter(Filter{"status": "active", "type": "user"})
func WithFilter(filter Filter) FindOption {
	return func(o *FindOptions) {
		o.Filter = filter
	}
}

// WithAdvancedFilter sets an advanced filter with support for different operators.
// Use this when you need more complex filtering than simple equality checks.
//
// Example:
//
//	WithAdvancedFilter(
//	  NewOrGroup(
//	    NewCondition("status", FilterOpEqual, "active"),
//	    NewCondition("status", FilterOpEqual, "pending"),
//	  ),
//	)
func WithAdvancedFilter(expression FilterExpression) FindOption {
	return func(o *FindOptions) {
		o.AdvancedFilter = &AdvancedFilter{
			Expression: expression,
		}
	}
}

// WithAndGroup creates a filter group with AND logic
//
// Example:
//
//	WithAndGroup(
//	  NewCondition("age", FilterOpGreater, 21),
//	  NewCondition("status", FilterOpEqual, "active"),
//	)
func WithAndGroup(expressions ...FilterExpression) FindOption {
	return WithAdvancedFilter(NewAndGroup(expressions...))
}

// WithOrGroup creates a filter group with OR logic
//
// Example:
//
//	WithOrGroup(
//	  NewCondition("status", FilterOpEqual, "active"),
//	  NewCondition("status", FilterOpEqual, "pending"),
//	)
func WithOrGroup(expressions ...FilterExpression) FindOption {
	return WithAdvancedFilter(NewOrGroup(expressions...))
}

// WithFields sets specific fields/columns to retrieve instead of all columns.
// This is useful for optimizing queries when you don't need all columns.
//
// Example: WithFields("id", "name", "email")
func WithFields(fields ...string) FindOption {
	return func(o *FindOptions) {
		o.Fields = fields
	}
}

// WithSort sets the sort options for ordering the results.
// You can provide multiple sort options which will be applied in the order provided.
//
// Example:
//
//	WithSort(
//	  SortOption{Field: "created_at", Direction: SortDesc},
//	  SortOption{Field: "name", Direction: SortAsc},
//	)
func WithSort(sortOptions ...SortOption) FindOption {
	return func(o *FindOptions) {
		o.Sort = sortOptions
	}
}

// WithPagination sets pagination options using limit and offset values.
// Limit controls how many records to return, offset controls where to start.
//
// Example: WithPagination(10, 30) // 10 records starting at offset 30
func WithPagination(limit, offset int) FindOption {
	return func(o *FindOptions) {
		o.Pagination = &Pagination{
			Limit:  limit,
			Offset: offset,
		}
	}
}
