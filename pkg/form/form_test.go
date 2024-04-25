package form

import (
	"testing"

	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autopus/bootstrap/pkg/types"
)

func TestForm_ValidateFields(t *testing.T) {
	type TestItem struct {
		name         string
		fields       []Field
		input        any
		errorMessage string
	}

	runTests := func(tests []TestItem) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var inputs []types.M
				if input, ok := tt.input.(types.M); ok {
					inputs = append(inputs, input)
				} else if input, ok := tt.input.([]types.M); ok {
					inputs = input
				} else {
					require.Fail(t, "invalid input")
				}

				for _, input := range inputs {
					err := ValidateFields(input, tt.fields)
					if tt.errorMessage == "" {
						assert.NoError(t, err, "expect no error message")
					} else {
						require.Error(t, err)
						assert.Equal(t, tt.errorMessage, err.Error(), "expect error message")
					}
				}
			})
		}
	}

	t.Run("full completed fields", func(t *testing.T) {
		tests := []TestItem{
			{
				name: "valid input",
				fields: []Field{
					{Name: "folder_id", Type: "string", Required: true},
					{Name: "file_id", Type: "number", Required: true},
					{Name: "owner", Type: "string"},
					{Name: "allowed_owners", Type: "string", Multiple: true, Required: true},
					{Name: "size", Type: "number"},
					{Name: "max_retries", Type: "integer"},
					{Name: "allowed_age", Type: "integer", Validation: &Validation{Min: lo.ToPtr(18.0), Max: lo.ToPtr(60.0)}},
					{Name: "size_mb", Type: "number"},
					{Name: "deleted", Type: "boolean"},
					{Name: "password", Type: "secret"},
					{Name: "gender", Type: "string", Options: []Option{{Label: "Male", Value: "male"}, {Label: "Female", Value: "female"}}},
				},
				input: types.M{
					"folder_id":      "123",
					"file_id":        123,
					"owner":          "auth",
					"allowed_owners": []any{"here", "there"},
					"max_retries":    9,
					"size":           123,
					"size_mb":        123.45,
					"deleted":        false,
					"password":       "do-not-tell",
					"gender":         "male",
					"allowed_age":    35,
				},
			},
			{
				name: "missing required field",
				fields: []Field{
					{Name: "folder_id", Type: "string", Required: true},
					{Name: "owner", Type: "string"},
				},
				input: types.M{
					"owner": "auth",
				},
				errorMessage: "missing required field folder_id",
			},
		}

		runTests(tests)
	})

	t.Run("field type string", func(t *testing.T) {
		tests := []TestItem{
			{
				name: "valid multiple string field",
				fields: []Field{
					{Name: "ingredients", Type: "string", Multiple: true},
				},
				input: types.M{
					"ingredients": []any{"salt", "pepper", "garlic"},
				},
			},
			{
				name: "invalid string option value",
				fields: []Field{
					{Name: "gender", Type: "string", Options: []Option{{Label: "Male", Value: "male"}, {Label: "Female", Value: "female"}}},
				},
				input: types.M{
					"gender": "invalid",
				},
				errorMessage: "field gender has invalid value: \"invalid\" not found in options",
			},
			{
				name: "validation: valid min length",
				fields: []Field{
					{Name: "name", Type: "string", Validation: &Validation{Min: lo.ToPtr(3.0)}},
				},
				input: types.M{
					"name": "abc",
				},
			},
			{
				name: "validation: invalid min length",
				fields: []Field{
					{Name: "name", Type: "string", Validation: &Validation{Min: lo.ToPtr(3.0)}},
				},
				input: types.M{
					"name": "ab",
				},
				errorMessage: "field name must be longer than or equal to 3 characters",
			},
			{
				name: "validation: invalid max length",
				fields: []Field{
					{Name: "name", Type: "string", Validation: &Validation{Max: lo.ToPtr(3.0)}},
				},
				input: types.M{
					"name": "hello",
				},
				errorMessage: "field name must be shorter than or equal to 3 characters",
			},
			{
				name: "validation: in range",
				fields: []Field{
					{Name: "name", Type: "string", Validation: &Validation{Min: lo.ToPtr(3.0), Max: lo.ToPtr(6.0)}},
				},
				input: types.M{
					"name": "hello",
				},
			},
		}

		runTests(tests)
	})

	t.Run("field type number", func(t *testing.T) {
		tests := []TestItem{
			{
				name: "invalid optional number(int) field",
				fields: []Field{
					{Name: "age", Type: "number"},
				},
				input: types.M{
					"age": "1",
				},
				errorMessage: "field age must be number",
			},
			{
				name: "invalid required number(int) field",
				fields: []Field{
					{Name: "age", Type: "number", Required: true},
				},
				input: types.M{
					"age": "1",
				},
				errorMessage: "field age must be number",
			},
			{
				name: "valid multiple number(int) field",
				fields: []Field{
					{Name: "sizes", Type: "number", Multiple: true},
				},
				input: types.M{
					"sizes": []any{1, 2, 3},
				},
			},
			{
				name: "invalid multiple number(int) field",
				fields: []Field{
					{Name: "sizes", Type: "number", Multiple: true},
				},
				input: types.M{
					"sizes": []any{1, "two", 3},
				},
				errorMessage: "field sizes must be array of number: the value \"two\" is not a number",
			},
			{
				name: "valid multiple float field",
				fields: []Field{
					{Name: "sizes", Type: "number", Multiple: true},
				},
				input: types.M{
					"sizes": []any{1.25, 2, 3.99},
				},
			},
			{
				name: "invalid required float field",
				fields: []Field{
					{Name: "amount", Type: "number", Required: true},
				},
				input: types.M{
					"amount": "1.33",
				},
				errorMessage: "field amount must be number",
			},
			{
				name: "invalid number option value",
				fields: []Field{
					{Name: "ages", Type: "number", Options: []Option{{Label: "20", Value: 20}, {Label: "30", Value: 30}}},
				},
				input: types.M{
					"ages": 100,
				},
				errorMessage: "field ages has invalid value: \"100\" not found in options",
			},
			{
				name: "valid number option value",
				fields: []Field{
					{Name: "ages", Type: "number", Options: []Option{{Label: "20", Value: 20}, {Label: "30", Value: 30}}},
				},
				input: types.M{
					"ages": 20,
				},
			},
			{
				name: "validation: invalid min value",
				fields: []Field{
					{Name: "age", Type: "number", Validation: &Validation{Min: lo.ToPtr(18.0)}},
				},
				input: types.M{
					"age": 17.9,
				},
				errorMessage: "field age must be greater than or equal to 18",
			},
			{
				name: "validation: invalid max value",
				fields: []Field{
					{Name: "age", Type: "number", Validation: &Validation{Max: lo.ToPtr(18.0)}},
				},
				input: types.M{
					"age": 18.1,
				},
				errorMessage: "field age must be less than or equal to 18",
			},
			{
				name: "validation: valid min/max value",
				fields: []Field{
					{Name: "age", Type: "number", Validation: &Validation{Min: lo.ToPtr(18.0), Max: lo.ToPtr(28.0)}},
				},
				input: types.M{
					"age": 20,
				},
			},
		}

		runTests(tests)
	})

	t.Run("integer field", func(t *testing.T) {
		tests := []TestItem{

			{
				name: "valid single integer field",
				fields: []Field{
					{Name: "age", Type: "integer"},
				},
				input: types.M{
					"age": 20,
				},
			},
			{
				name: "valid single integer field, it is a float with no decimal",
				fields: []Field{
					{Name: "age", Type: "integer"},
				},
				input: types.M{
					"age": 20.0,
				},
			},

			{
				name: "invalid single integer field, it is a float",
				fields: []Field{
					{Name: "age", Type: "integer"},
				},
				input: types.M{
					"age": 20.25,
				},
				errorMessage: "field age must be integer",
			},
			{
				name: "invalid single integer field, it is a string",
				fields: []Field{
					{Name: "age", Type: "integer"},
				},
				input: types.M{
					"age": "20",
				},
				errorMessage: "field age must be integer",
			},
			{
				name: "valid multiple integer field",
				fields: []Field{
					{Name: "ages", Type: "integer", Multiple: true},
				},
				input: types.M{
					"ages": []any{20, 30, 40},
				},
			},
			{
				name: "invalid multiple integer field",
				fields: []Field{
					{Name: "ages", Type: "integer", Multiple: true},
				},
				input: types.M{
					"ages": []any{20, 30.2, 40},
				},
				errorMessage: "field ages must be array of integers: the value \"30.2\" is not an integer",
			},
			{
				name: "validation: invalid min value",
				fields: []Field{
					{Name: "age", Type: "number", Validation: &Validation{Min: lo.ToPtr(18.0)}},
				},
				input: types.M{
					"age": 17,
				},
				errorMessage: "field age must be greater than or equal to 18",
			},
			{
				name: "validation: invalid max value",
				fields: []Field{
					{Name: "age", Type: "number", Validation: &Validation{Max: lo.ToPtr(18.0)}},
				},
				input: types.M{
					"age": 19,
				},
				errorMessage: "field age must be less than or equal to 18",
			},
			{
				name: "validation: valid min/max value",
				fields: []Field{
					{Name: "age", Type: "number", Validation: &Validation{Min: lo.ToPtr(18.0), Max: lo.ToPtr(28.0)}},
				},
				input: types.M{
					"age": 20,
				},
			},
		}

		runTests(tests)
	})

	t.Run("boolean field", func(t *testing.T) {
		tests := []TestItem{
			{
				name: "invalid boolean field",
				fields: []Field{
					{Name: "smoking", Type: "boolean"},
				},
				input: types.M{
					"smoking": "yes",
				},
				errorMessage: "field smoking must be boolean",
			},
			{
				name: "valid multiple boolean fields",
				fields: []Field{
					{Name: "preferences", Type: "boolean", Multiple: true},
				},
				input: types.M{
					"preferences": []any{true, false, true},
				},
			},
			{
				name: "invalid multiple boolean fields",
				fields: []Field{
					{Name: "preferences", Type: "boolean", Multiple: true},
				},
				input: types.M{
					"preferences": []any{true, "yes", true},
				},
				errorMessage: "field preferences must be array of booleans",
			},
			{
				name: "invalid boolean option value",
				fields: []Field{
					{Name: "preferences", Type: "boolean", Options: []Option{{Label: "false", Value: false}}},
				},
				input: types.M{
					"preferences": true,
				},
				errorMessage: "field preferences has invalid value: \"true\" not found in options",
			},
			{
				name: "valid boolean option value",
				fields: []Field{
					{Name: "preferences", Type: "boolean", Options: []Option{{Label: "true", Value: true}, {Label: "false", Value: false}}},
				},
				input: types.M{
					"preferences": true,
				},
			},
		}

		runTests(tests)
	})
}
