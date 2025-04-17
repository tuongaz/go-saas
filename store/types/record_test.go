package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecord_Normalise(t *testing.T) {
	t.Run("normalizes JSON bytes", func(t *testing.T) {
		// Create a JSON byte array
		jsonData := []byte(`{"name":"test","age":30}`)

		// Create a record with the JSON bytes
		record := Record{
			"data": jsonData,
		}

		// Normalize the record
		record.Normalise()

		// Assert that the JSON bytes have been converted to a map
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "JSON data should be converted to a map")
		assert.Equal(t, "test", data["name"])
		assert.Equal(t, float64(30), data["age"])
	})

	t.Run("normalizes time.Time", func(t *testing.T) {
		// Create a time
		now := time.Now()

		// Create a record with the time
		record := Record{
			"created_at": now,
		}

		// Normalize the record
		record.Normalise()

		// Assert that the time has been converted to a string in RFC3339 format
		timeStr, ok := record["created_at"].(string)
		assert.True(t, ok, "Time should be converted to a string")
		assert.Equal(t, now.Format(time.RFC3339), timeStr)
	})
}

func TestRecord_PrepareForDB(t *testing.T) {
	t.Run("prepares basic types", func(t *testing.T) {
		// Create a record with various basic types
		record := Record{
			"name":    "test",
			"age":     30,
			"balance": 100.5,
			"active":  true,
			"nil":     nil,
		}

		// Prepare the record for DB
		keys, values, placeholders, err := record.PrepareForDB()

		// Assert no error
		assert.NoError(t, err)

		// Assert the keys, values, and placeholders have the correct length
		assert.Len(t, keys, 5)
		assert.Len(t, values, 5)
		assert.Len(t, placeholders, 5)

		// Create maps to check values by key
		valueMap := make(map[string]interface{})
		placeholderMap := make(map[string]string)

		// Map each key to its corresponding value and placeholder
		for i, key := range keys {
			valueMap[key] = values[i]
			placeholderMap[key] = placeholders[i]
		}

		// Assert values by key
		assert.Equal(t, "test", valueMap["name"])
		assert.Equal(t, 30, valueMap["age"])
		assert.Equal(t, 100.5, valueMap["balance"])
		assert.Equal(t, true, valueMap["active"])
		assert.Nil(t, valueMap["nil"])

		// Assert placeholders are in the correct format
		for _, placeholder := range placeholders {
			assert.Regexp(t, `\$\d+`, placeholder, "Placeholder should be in format $n")
		}
	})

	t.Run("handles complex types", func(t *testing.T) {
		// Create a struct
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		user := User{
			Name: "test",
			Age:  30,
		}

		// Create a record with a complex type
		record := Record{
			"user": user,
		}

		// Prepare the record for DB
		keys, values, placeholders, err := record.PrepareForDB()

		// Assert no error
		assert.NoError(t, err)

		// Assert there is one key, value, and placeholder
		assert.Len(t, keys, 1)
		assert.Len(t, values, 1)
		assert.Len(t, placeholders, 1)

		// Assert the key is "user"
		assert.Equal(t, "user", keys[0])

		// Assert the value is a JSON string
		userJSON, ok := values[0].(string)
		assert.True(t, ok, "Complex type should be converted to a JSON string")

		// Unmarshal the JSON string and assert it matches the original struct
		var userFromJSON User
		err = json.Unmarshal([]byte(userJSON), &userFromJSON)
		assert.NoError(t, err)
		assert.Equal(t, user.Name, userFromJSON.Name)
		assert.Equal(t, user.Age, userFromJSON.Age)

		// Assert the placeholder is in the correct format
		assert.Regexp(t, `\$\d+`, placeholders[0], "Placeholder should be in format $n")
	})
}

func TestRecord_Get(t *testing.T) {
	record := Record{
		"name": "test",
		"age":  30,
	}

	assert.Equal(t, "test", record.Get("name"))
	assert.Equal(t, 30, record.Get("age"))
	assert.Nil(t, record.Get("nonexistent"))
}

func TestRecord_Decode(t *testing.T) {
	// Create a struct for decoding
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create a record with matching fields
	record := Record{
		"name": "test",
		"age":  30,
	}

	// Create a variable to decode into
	var user User

	// Decode the record
	err := record.Decode(&user)

	// Assert no error
	assert.NoError(t, err)

	// Assert the decoded values match
	assert.Equal(t, "test", user.Name)
	assert.Equal(t, 30, user.Age)
}
