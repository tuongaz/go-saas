package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNotFoundErr(t *testing.T) {
	// Create a base error
	baseErr := errors.New("record not found")

	// Create a not found error
	notFoundErr := NewNotFoundErr(baseErr)

	// Assert the error type is correct
	assert.IsType(t, &NotFoundErr{}, notFoundErr)

	// Assert the error message is correct
	assert.Equal(t, "record not found", notFoundErr.Error())

	// Assert the error is wrapped correctly
	assert.True(t, errors.Is(notFoundErr, baseErr))

	// Test IsNotFoundError function
	assert.True(t, IsNotFoundError(notFoundErr))
	assert.False(t, IsNotFoundError(baseErr))
}

func TestNewDuplicateKeyErr(t *testing.T) {
	// Create a base error
	baseErr := errors.New("duplicate key value violates unique constraint")

	// Create a duplicate key error
	duplicateKeyErr := NewDuplicateKeyErr(baseErr)

	// Assert the error type is correct
	assert.IsType(t, &DuplicateKeyErr{}, duplicateKeyErr)

	// Assert the error message is correct
	assert.Equal(t, "duplicate key violation: duplicate key value violates unique constraint", duplicateKeyErr.Error())

	// Assert the error is wrapped correctly
	assert.True(t, errors.Is(duplicateKeyErr, baseErr))

	// Test IsDuplicateKeyError function
	assert.True(t, IsDuplicateKeyError(duplicateKeyErr))
	assert.False(t, IsDuplicateKeyError(baseErr))
}

func TestNewForeignKeyErr(t *testing.T) {
	// Create a base error
	baseErr := errors.New("foreign key constraint violation")

	// Create a foreign key error
	foreignKeyErr := NewForeignKeyErr(baseErr)

	// Assert the error type is correct
	assert.IsType(t, &ForeignKeyErr{}, foreignKeyErr)

	// Assert the error message is correct
	assert.Equal(t, "foreign key violation: foreign key constraint violation", foreignKeyErr.Error())

	// Assert the error is wrapped correctly
	assert.True(t, errors.Is(foreignKeyErr, baseErr))

	// Test IsForeignKeyError function
	assert.True(t, IsForeignKeyError(foreignKeyErr))
	assert.False(t, IsForeignKeyError(baseErr))
}

func TestNewNotNullErr(t *testing.T) {
	// Create a base error
	baseErr := errors.New("null value in column violates not-null constraint")

	// Create a not null error
	notNullErr := NewNotNullErr(baseErr)

	// Assert the error type is correct
	assert.IsType(t, &NotNullErr{}, notNullErr)

	// Assert the error message is correct
	assert.Equal(t, "not-null constraint violation: null value in column violates not-null constraint", notNullErr.Error())

	// Assert the error is wrapped correctly
	assert.True(t, errors.Is(notNullErr, baseErr))

	// Test IsNotNullError function
	assert.True(t, IsNotNullError(notNullErr))
	assert.False(t, IsNotNullError(baseErr))
}

func TestNewDBError(t *testing.T) {
	// Create a base error
	baseErr := errors.New("database error")

	// Create a DB error
	dbErr := NewDBError(baseErr)

	// Assert the error type is correct
	assert.IsType(t, &DBError{}, dbErr)

	// Assert the error message is correct
	assert.Equal(t, "database error", dbErr.Error())

	// Assert the error is wrapped correctly
	assert.True(t, errors.Is(dbErr, baseErr))
}
