package form

import (
	"fmt"
	"strings"

	"go.uber.org/multierr"

	"github.com/tuongaz/go-saas/pkg/errors"
	"github.com/tuongaz/go-saas/pkg/types"
)

const (
	String  = "string"
	Integer = "integer"
	Number  = "number"
	Boolean = "boolean"
	Secret  = "secret"

	Text     = "text"
	Select   = "select"
	TextArea = "textarea"
)

type Option struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type Validation struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
}

type Field struct {
	Name        string      `json:"name" validate:"required"`
	Label       string      `json:"label" validate:"required"`
	Type        string      `json:"type" validate:"required,oneof=string integer number boolean secret"`
	Description string      `json:"description" validate:"required"`
	InputMethod string      `json:"input_method" validate:"required,oneof=text select textarea"`
	Options     []Option    `json:"options"`
	HelpText    string      `json:"help_text"`
	Dynamic     bool        `json:"dynamic"`
	Multiple    bool        `json:"multiple"`
	Required    bool        `json:"required"`
	Validation  *Validation `json:"validation"`
	Depends     []string    `json:"depends"`
}

func ValidateFields(data types.M, fields []Field) error {
	var allErrors error

	appendError := func(err error) {
		allErrors = multierr.Append(allErrors, err)
	}

	// make sure only defined fields are passed
	for k := range data {
		var found bool
		for _, field := range fields {
			if field.Name == k {
				found = true
				break
			}
		}

		if !found {
			return errors.NewValidationError(fmt.Errorf("field %s is not defined", k))
		}
	}

	for _, field := range fields {
		if field.Type == "" {
			field.Type = "string"
		}

		value, ok := data[field.Name]
		if !ok && field.Required {
			allErrors = multierr.Append(
				allErrors,
				fmt.Errorf("missing required field %s", field.Name),
			)
			continue
		}

		// if it is empty, but the field is not required, then skip
		if value == nil && !field.Required {
			continue
		}
		if v, ok := value.(string); ok && field.Required && strings.TrimSpace(v) == "" {
			appendError(fmt.Errorf("field %s must not be empty", field.Name))
			continue
		}

		var values []any

		//validate field type
		switch field.Type {
		case String, Secret:
			if field.Multiple {
				v, ok := value.([]any)
				if ok {
					if err := validateStrings(v); err != nil {
						appendError(fmt.Errorf("field %s must be array of strings: %s", field.Name, err.Error()))
						continue
					}
					values = v
					continue
				}

				_, ok = value.([]string)
				if ok {
					values = value.([]any)
					continue
				}

				appendError(fmt.Errorf("field %s must be array of strings", field.Name))
				continue
			} else {
				v, ok := value.(string)
				if !ok {
					appendError(fmt.Errorf("field %s must be string", field.Name))
					continue
				}
				values = append(values, v)
			}
		case Number:
			if field.Multiple {
				v, ok := value.([]any)
				if !ok {
					appendError(fmt.Errorf("field %s must be array of numbers", field.Name))
					continue
				} else {
					if err := validateNumbers(v); err != nil {
						appendError(fmt.Errorf("field %s must be array of number: %s", field.Name, err))
						continue
					}
					values = v
				}
			} else {
				if !isNumber(value) {
					appendError(fmt.Errorf("field %s must be number", field.Name))
					continue
				}
				values = []any{value}
			}
		case Integer:
			if field.Multiple {
				v, ok := value.([]any)
				if !ok {
					appendError(fmt.Errorf("field %s must be array of integers", field.Name))
					continue
				} else {
					if err := validateIntegers(v); err != nil {
						appendError(fmt.Errorf("field %s must be array of integers: %s", field.Name, err))
						continue
					}
					values = v
				}
			} else {
				if !isInteger(value) {
					appendError(fmt.Errorf("field %s must be integer", field.Name))
					continue
				}
				values = []any{value}
			}
		case Boolean:
			if field.Multiple {
				v, ok := value.([]any)
				if !ok {
					appendError(fmt.Errorf("field %s must be array of booleans", field.Name))
					continue
				} else {
					for _, item := range v {
						if _, ok := item.(bool); !ok {
							appendError(fmt.Errorf("field %s must be array of booleans", field.Name))
							continue
						}
					}
					values = v
				}
			} else {
				if _, ok := value.(bool); !ok {
					appendError(fmt.Errorf("field %s must be boolean", field.Name))
					continue
				}
				values = []any{value}
			}
		default:
			appendError(errors.NewValidationError(fmt.Errorf("field %s has invalid type: %s", field.Name, field.Type)))
			return allErrors
		}

		if err := valuesInOptions(field.Name, field, values...); err != nil {
			appendError(err)
			continue
		}

		// validate custom validator
		if err := validateCustom(field, values); err != nil {
			appendError(err)
			continue
		}

		if !field.Multiple {
			value = values[0]
		} else {
			value = values
		}

		data[field.Name] = value
	}

	if allErrors != nil {
		return errors.NewValidationError(allErrors)
	}

	return nil
}

func validateCustom(field Field, values []any) error {
	if field.Validation == nil {
		return nil
	}

	if field.Validation.Min != nil {
		switch field.Type {
		case Integer, Number:
			for _, v := range values {
				if value, ok := v.(float64); ok && value < *field.Validation.Min {
					return fmt.Errorf("field %s must be greater than or equal to %v", field.Name, *field.Validation.Min)
				}

				if value, ok := v.(int); ok && float64(value) < *field.Validation.Min {
					return fmt.Errorf("field %s must be greater than or equal to %v", field.Name, *field.Validation.Min)
				}
			}
		case String, Secret:
			for _, v := range values {
				if float64(len(v.(string))) < *field.Validation.Min {
					return fmt.Errorf("field %s must be longer than or equal to %v characters", field.Name, *field.Validation.Min)
				}
			}
		}
	}

	if field.Validation.Max != nil {
		switch field.Type {
		case Integer, Number:
			for _, v := range values {
				if value, ok := v.(int); ok && float64(value) > *field.Validation.Max {
					return fmt.Errorf("field %s must be less than or equal to %v", field.Name, *field.Validation.Max)
				}
				if value, ok := v.(float64); ok && value > *field.Validation.Max {
					return fmt.Errorf("field %s must be less than or equal to %v", field.Name, *field.Validation.Max)
				}
			}
		case String, Secret:
			for _, v := range values {
				if float64(len(v.(string))) > *field.Validation.Max {
					return fmt.Errorf("field %s must be shorter than or equal to %v characters", field.Name, *field.Validation.Max)
				}
			}
		}
	}

	return nil
}

func valuesInOptions(fieldName string, field Field, values ...any) error {
	if field.Options == nil {
		return nil
	}

	for _, val := range values {
		found := false
		for _, option := range field.Options {
			if option.Value == val {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field %s has invalid value: \"%v\" not found in options", fieldName, val)
		}
	}
	return nil
}

func validateStrings(list []any) error {
	for _, item := range list {
		if _, ok := item.(string); !ok {
			return errors.NewValidationError(fmt.Errorf("the value \"%v\" is not a string", item))
		}
	}
	return nil
}

func validateIntegers(list []any) error {
	for _, item := range list {
		if !isInteger(item) {
			return errors.NewValidationError(fmt.Errorf("the value \"%v\" is not an integer", item))
		}
	}
	return nil
}

func validateNumbers(list []any) error {
	for _, item := range list {
		if !isNumber(item) {
			return errors.NewValidationError(fmt.Errorf("the value \"%v\" is not a number", item))
		}
	}
	return nil
}

func isInteger(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return float64(int(v.(float64))) == v.(float64)
	default:
		return false
	}
}

func isNumber(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}
