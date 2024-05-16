package apierror

import (
	"net/http"
)

type APIError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
	err     error
}

func (e APIError) Error() string {
	if e.err == nil {
		return e.Message
	}

	return e.err.Error()
}

func (e APIError) Unwrap() error {
	return e.err
}

func NewUnauthorizedErr(message string, err error, data ...map[string]any) error {
	return newError(http.StatusUnauthorized, message, err, data...)
}

func NewNotFoundErr(message string, err error, data ...map[string]any) error {
	return newError(http.StatusNotFound, message, err, data...)
}

func NewValidationError(message string, err error, data ...map[string]any) error {
	return newError(http.StatusBadRequest, message, err, data...)
}

func NewForbiddenError(message string, err error, data ...map[string]any) error {
	return newError(http.StatusForbidden, message, err, data...)
}

func newError(code int, message string, err error, data ...map[string]any) error {
	var d map[string]any
	if len(data) > 0 {
		d = data[0]
	}

	return &APIError{
		Code:    code,
		Message: message,
		Data:    d,
		err:     err,
	}
}
