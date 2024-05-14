package template

import (
	"reflect"
	"testing"

	"github.com/tuongaz/go-saas/pkg/types"
)

func TestReplace(t *testing.T) {
	ctx := types.M{
		"name": "John",
		"age":  30,
		"address": types.M{
			"city": "Melbourne",
		},
	}

	tests := []struct {
		name   string
		input  string
		ctx    types.M
		expect string
	}{
		{"Simple Replacement", "Hello, {{name}}!", ctx, "Hello, John!"},
		{"No Replacement", "Hello, World!", ctx, "Hello, World!"},
		{"Nested Replacement", "City: {{address.city}}", ctx, "City: Melbourne"},
		{"Non-existent Key", "Hello, {{unknown}}!", ctx, "Hello, !"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Replace(test.input, test.ctx)
			if result != test.expect {
				t.Errorf("Expected %s, got %s", test.expect, result)
			}
		})
	}
}

func TestMapReplace(t *testing.T) {
	ctx := types.M{
		"auth": types.M{
			"name": "Alice",
			"age":  28,
		},
	}

	input := types.M{
		"greeting": "Hello, {{auth.name}}!",
		"age":      "{{auth.age}}",
		"static":   "Static value",
	}

	expected := types.M{
		"greeting": "Hello, Alice!",
		"age":      "28",
		"static":   "Static value",
	}

	result := MapReplace(input, ctx)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
