package store

// Filter is a simple map-based filter where keys are field names and values are the values to match
type Filter map[string]any

// FilterOp represents a filter operation
type FilterOp string

const (
	FilterOpEqual        FilterOp = "="
	FilterOpNotEqual     FilterOp = "!="
	FilterOpGreater      FilterOp = ">"
	FilterOpGreaterEqual FilterOp = ">="
	FilterOpLess         FilterOp = "<"
	FilterOpLessEqual    FilterOp = "<="
	FilterOpLike         FilterOp = "LIKE"
	FilterOpILike        FilterOp = "ILIKE"
	FilterOpIn           FilterOp = "IN"
	FilterOpNotIn        FilterOp = "NOT IN"
	FilterOpIsNull       FilterOp = "IS NULL"
	FilterOpIsNotNull    FilterOp = "IS NOT NULL"
)

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field string
	Op    FilterOp
	Value any
}

// AdvancedFilter represents a collection of filter conditions
type AdvancedFilter struct {
	Conditions []FilterCondition
}

// SortDirection represents the direction of sorting
type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// SortOption represents a sorting option
type SortOption struct {
	Field     string
	Direction SortDirection
}
