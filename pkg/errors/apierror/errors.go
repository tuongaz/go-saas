package apierror

import (
	"errors"
	"fmt"
	"net/http"
)

type APIError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
	error   error
}

func New(msg string, params ...any) error {
	return fmt.Errorf(msg, params...)
}

func NewUnauthorizedErr(err error) error {
	return &UnauthorizedErr{err: err}
}

type UnauthorizedErr struct {
	err error
}

func (e UnauthorizedErr) Error() string {
	return e.err.Error()
}

func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}

	var vErr *UnauthorizedErr
	return errors.As(err, &vErr)
}

func NewNotFoundErr(err error) error {
	return &NotFoundErr{err: err}
}

type NotFoundErr struct {
	err error
}

func (e NotFoundErr) Error() string {
	return e.err.Error()
}

func IsNotFound1(err error) bool {
	if err == nil {
		return false
	}
	var e *NotFoundErr
	return errors.As(err, &e)
}

func NewValidationError(message string, data map[string]any, err error) *ValidationError {
	return &ValidationError{&APIError{
		Code:    http.StatusBadRequest,
		Message: message,
		Data:    data,
		error:   err,
	}}
}

type ValidationError struct {
	*APIError
}

func (v ValidationError) Error() string {
	if v.error == nil {
		return v.Message
	}

	return v.error.Error()
}

func IsValidation(err error) bool {
	if err == nil {
		return false
	}

	var vErr *ValidationError
	return errors.As(err, &vErr)
}

func NewForbiddenError(err error) *ForbiddenError {
	return &ForbiddenError{err: err}
}

type ForbiddenError struct {
	err error
}

func (f ForbiddenError) Error() string {
	return f.err.Error()
}

func IsForbidden(err error) bool {
	if err == nil {
		return false
	}

	var vErr *ForbiddenError
	return errors.As(err, &vErr)
}
