package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	// Create a filter
	filter := Filter{
		"name":   "John",
		"age":    30,
		"active": true,
	}

	// Assert the filter values
	assert.Equal(t, "John", filter["name"])
	assert.Equal(t, 30, filter["age"])
	assert.Equal(t, true, filter["active"])
}

func TestFilterOp(t *testing.T) {
	// Test all filter operations
	assert.Equal(t, FilterOp("="), FilterOpEqual)
	assert.Equal(t, FilterOp("!="), FilterOpNotEqual)
	assert.Equal(t, FilterOp(">"), FilterOpGreater)
	assert.Equal(t, FilterOp(">="), FilterOpGreaterEqual)
	assert.Equal(t, FilterOp("<"), FilterOpLess)
	assert.Equal(t, FilterOp("<="), FilterOpLessEqual)
	assert.Equal(t, FilterOp("LIKE"), FilterOpLike)
	assert.Equal(t, FilterOp("ILIKE"), FilterOpILike)
	assert.Equal(t, FilterOp("IN"), FilterOpIn)
	assert.Equal(t, FilterOp("NOT IN"), FilterOpNotIn)
	assert.Equal(t, FilterOp("IS NULL"), FilterOpIsNull)
	assert.Equal(t, FilterOp("IS NOT NULL"), FilterOpIsNotNull)
}

func TestFilterCondition(t *testing.T) {
	// Create filter conditions
	equalCondition := FilterCondition{
		Field: "name",
		Op:    FilterOpEqual,
		Value: "John",
	}

	// Assert the condition values
	assert.Equal(t, "name", equalCondition.Field)
	assert.Equal(t, FilterOpEqual, equalCondition.Op)
	assert.Equal(t, "John", equalCondition.Value)
}

func TestAdvancedFilter(t *testing.T) {
	// Create conditions
	name := NewCondition("name", FilterOpEqual, "John")
	age := NewCondition("age", FilterOpGreaterEqual, 30)

	// Create an AND group
	andGroup := NewAndGroup(name, age)

	// Create an advanced filter with AND group
	advancedFilter := AdvancedFilter{
		Expression: andGroup,
	}

	// Assert the filter expression type
	assert.Equal(t, FilterExprTypeGroup, andGroup.Type())

	// Check the group attributes
	group := advancedFilter.Expression.(FilterGroup)
	assert.Equal(t, LogicOpAnd, group.Logic)
	assert.Len(t, group.Expressions, 2)

	// Check the first condition (name = "John")
	cond1 := group.Expressions[0].(FilterCondition)
	assert.Equal(t, "name", cond1.Field)
	assert.Equal(t, FilterOpEqual, cond1.Op)
	assert.Equal(t, "John", cond1.Value)

	// Check the second condition (age >= 30)
	cond2 := group.Expressions[1].(FilterCondition)
	assert.Equal(t, "age", cond2.Field)
	assert.Equal(t, FilterOpGreaterEqual, cond2.Op)
	assert.Equal(t, 30, cond2.Value)

	// Create an OR group
	orGroup := NewOrGroup(
		NewCondition("status", FilterOpEqual, "active"),
		NewCondition("status", FilterOpEqual, "pending"),
	)

	// Assert the OR group
	assert.Equal(t, FilterExprTypeGroup, orGroup.Type())
	assert.Equal(t, LogicOpOr, orGroup.Logic)
	assert.Len(t, orGroup.Expressions, 2)
}

func TestSortDirection(t *testing.T) {
	// Test sort directions
	assert.Equal(t, SortDirection("ASC"), SortAsc)
	assert.Equal(t, SortDirection("DESC"), SortDesc)
}

func TestSortOption(t *testing.T) {
	// Create sort options
	sortOption := SortOption{
		Field:     "name",
		Direction: SortAsc,
	}

	// Assert the sort option values
	assert.Equal(t, "name", sortOption.Field)
	assert.Equal(t, SortAsc, sortOption.Direction)
}
