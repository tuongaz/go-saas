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

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var e *NotFoundErr
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
