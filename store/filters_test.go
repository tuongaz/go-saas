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
	conditions := []FilterCondition{
		{
			Field: "name",
			Op:    FilterOpEqual,
			Value: "John",
		},
		{
			Field: "age",
			Op:    FilterOpGreaterEqual,
			Value: 30,
		},
	}

	// Create an advanced filter
	advancedFilter := AdvancedFilter{
		Conditions: conditions,
	}

	// Assert the filter conditions
	assert.Len(t, advancedFilter.Conditions, 2)
	assert.Equal(t, "name", advancedFilter.Conditions[0].Field)
	assert.Equal(t, FilterOpEqual, advancedFilter.Conditions[0].Op)
	assert.Equal(t, "John", advancedFilter.Conditions[0].Value)
	assert.Equal(t, "age", advancedFilter.Conditions[1].Field)
	assert.Equal(t, FilterOpGreaterEqual, advancedFilter.Conditions[1].Op)
	assert.Equal(t, 30, advancedFilter.Conditions[1].Value)
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
