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

// LogicOp represents a logical operator for combining filter conditions
type LogicOp string

const (
	LogicOpAnd LogicOp = "AND"
	LogicOpOr  LogicOp = "OR"
)

// FilterExpressionType represents the type of filter expression
type FilterExpressionType string

const (
	FilterExprTypeCondition FilterExpressionType = "condition"
	FilterExprTypeGroup     FilterExpressionType = "group"
)

// FilterExpression is the interface for filter conditions and groups
type FilterExpression interface {
	Type() FilterExpressionType
}

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field string
	Op    FilterOp
	Value any
}

// Type returns the type of the expression
func (c FilterCondition) Type() FilterExpressionType {
	return FilterExprTypeCondition
}

// FilterGroup represents a group of filter expressions combined with a logical operator
type FilterGroup struct {
	Logic       LogicOp
	Expressions []FilterExpression
}

// Type returns the type of the expression
func (g FilterGroup) Type() FilterExpressionType {
	return FilterExprTypeGroup
}

// NewCondition creates a new filter condition
func NewCondition(field string, op FilterOp, value any) FilterCondition {
	return FilterCondition{
		Field: field,
		Op:    op,
		Value: value,
	}
}

// NewAndGroup creates a new filter group with AND logic
func NewAndGroup(expressions ...FilterExpression) FilterGroup {
	return FilterGroup{
		Logic:       LogicOpAnd,
		Expressions: expressions,
	}
}

// NewOrGroup creates a new filter group with OR logic
func NewOrGroup(expressions ...FilterExpression) FilterGroup {
	return FilterGroup{
		Logic:       LogicOpOr,
		Expressions: expressions,
	}
}

// AdvancedFilter represents a complex filter with nested expressions
type AdvancedFilter struct {
	Expression FilterExpression
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
