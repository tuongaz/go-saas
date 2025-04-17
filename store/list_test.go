package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuongaz/go-saas/store/types"
)

func TestList_Decode(t *testing.T) {
	// Create a struct for decoding
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create test records
	records := []types.Record{
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25},
	}

	// Create a list with the records
	list := List{
		Records: records,
		Meta: Metadata{
			Total:  2,
			Limit:  10,
			Offset: 0,
		},
	}

	// Create a slice to decode into
	var users []User

	// Decode the list
	err := list.Decode(&users)

	// Assert no error
	assert.NoError(t, err)

	// Assert the length of the decoded slice
	assert.Len(t, users, 2)

	// Assert the decoded values match
	assert.Equal(t, "John", users[0].Name)
	assert.Equal(t, 30, users[0].Age)
	assert.Equal(t, "Jane", users[1].Name)
	assert.Equal(t, 25, users[1].Age)
}

func TestList_Decode_Empty(t *testing.T) {
	// Create a struct for decoding
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create an empty list
	list := List{
		Records: []types.Record{},
		Meta: Metadata{
			Total:  0,
			Limit:  10,
			Offset: 0,
		},
	}

	// Create a slice to decode into
	var users []User

	// Decode the list
	err := list.Decode(&users)

	// Assert no error
	assert.NoError(t, err)

	// Assert the length of the decoded slice is 0
	assert.Len(t, users, 0)
}

func TestList_Decode_InvalidDestination(t *testing.T) {
	// Create a struct for decoding
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create test records
	records := []types.Record{
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25},
	}

	// Create a list with the records
	list := List{
		Records: records,
	}

	// Create a non-slice variable to decode into (this should fail)
	var user User

	// Decode the list
	err := list.Decode(&user)

	// Assert error
	assert.Error(t, err)
}
