package errors

import (
	goErrors "errors"
	"testing"
)

func TestNotAuthorizedErr(t *testing.T) {
	err := NewUnauthorizedErr(New("access denied"))
	if err.Error() != "access denied" {
		t.Errorf("Expected 'access denied', got '%s'", err.Error())
	}
}

func TestNotFoundErr(t *testing.T) {
	err := NewNotFoundErr(New("item not found"))
	if err.Error() != "item not found" {
		t.Errorf("Expected 'item not found', got '%s'", err.Error())
	}
}

func TestIsNotFound(t *testing.T) {
	notFoundErr := NewNotFoundErr(New("item not found"))
	if !IsNotFound(notFoundErr) {
		t.Errorf("Expected true, got false")
	}

	otherErr := goErrors.New("other error")
	if IsNotFound(otherErr) {
		t.Errorf("Expected false, got true")
	}
}

func TestValidationError(t *testing.T) {
	validationErr := NewValidationError(goErrors.New("invalid input"))
	if validationErr.Error() != "invalid input" {
		t.Errorf("Expected 'invalid input', got '%s'", validationErr.Error())
	}
}

func TestIsValidation(t *testing.T) {
	validationErr := NewValidationError(goErrors.New("invalid input"))
	if !IsValidation(validationErr) {
		t.Errorf("Expected true, got false")
	}

	otherErr := goErrors.New("other error")
	if IsValidation(otherErr) {
		t.Errorf("Expected false, got true")
	}
}
