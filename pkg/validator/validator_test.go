package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name string `validate:"required"`
	Age  int    `validate:"min=18"`
}

func TestValidate_Success(t *testing.T) {
	input := testStruct{
		Name: "John Doe",
		Age:  25,
	}

	err := Validate(input)
	assert.NoError(t, err)
}

func TestValidate_Failure(t *testing.T) {
	input := testStruct{
		Name: "", // Name is required
		Age:  17, // Age must be at least 18
	}

	err := Validate(input)
	assert.Error(t, err)
}
