package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("returns error when connection fails", func(t *testing.T) {
		store, err := New("invalid-datasource")
		assert.Error(t, err)
		assert.Nil(t, store)
	})
}

func TestStore_Collection(t *testing.T) {
	t.Run("returns collection with valid table name", func(t *testing.T) {
		store := &Store{}

		collection := store.Collection("users")
		assert.NotNil(t, collection)
		assert.Equal(t, "users", collection.Table())
	})

	t.Run("panics with invalid table name", func(t *testing.T) {
		store := &Store{}
		assert.Panics(t, func() {
			store.Collection("invalid-table-name")
		})
	})
}

func TestStore_Close(t *testing.T) {
	t.Run("handles nil db", func(t *testing.T) {
		store := &Store{db: nil}
		err := store.Close()
		assert.NoError(t, err)
	})
}
