package store

import (
	"errors"
)

func NewNotFoundErr(err error) error {
	return &NotFoundErr{err: err}
}

type NotFoundErr struct {
	err error
}

func (e NotFoundErr) Error() string {
	return e.err.Error()
}

func (e NotFoundErr) Unwrap() error {
	return e.err
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var e *NotFoundErr
	return errors.As(err, &e)
}

// DuplicateKeyErr represents a unique constraint violation error
type DuplicateKeyErr struct {
	err error
}

func NewDuplicateKeyErr(err error) error {
	return &DuplicateKeyErr{err: err}
}

func (e DuplicateKeyErr) Error() string {
	return "duplicate key violation: " + e.err.Error()
}

func (e DuplicateKeyErr) Unwrap() error {
	return e.err
}

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var e *DuplicateKeyErr
	return errors.As(err, &e)
}

// ForeignKeyErr represents a foreign key constraint violation error
type ForeignKeyErr struct {
	err error
}

func NewForeignKeyErr(err error) error {
	return &ForeignKeyErr{err: err}
}

func (e ForeignKeyErr) Error() string {
	return "foreign key violation: " + e.err.Error()
}

func (e ForeignKeyErr) Unwrap() error {
	return e.err
}

func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	var e *ForeignKeyErr
	return errors.As(err, &e)
}

// NotNullErr represents a not-null constraint violation error
type NotNullErr struct {
	err error
}

func NewNotNullErr(err error) error {
	return &NotNullErr{err: err}
}

func (e NotNullErr) Error() string {
	return "not-null constraint violation: " + e.err.Error()
}

func (e NotNullErr) Unwrap() error {
	return e.err
}

func IsNotNullError(err error) bool {
	if err == nil {
		return false
	}
	var e *NotNullErr
	return errors.As(err, &e)
}

type DBError struct {
	Err error
}

func NewDBError(err error) *DBError {
	return &DBError{Err: err}
}

func (e *DBError) Error() string {
	return e.Err.Error()
}

func (e *DBError) Unwrap() error {
	return e.Err
}
