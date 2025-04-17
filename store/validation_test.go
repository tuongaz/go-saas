package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidTableName(t *testing.T) {
	validTables := []string{
		"users",
		"users_table",
		"users_2",
		"users_table_2",
		"_users",
		"_users_2",
	}

	invalidTables := []string{
		"invalid-table-name",
		"2_invalid",
		"invalid table",
		"invalid+table",
		"",
		"users-table",
		"users.table",
		"users;table",
	}

	for _, table := range validTables {
		t.Run("valid table: "+table, func(t *testing.T) {
			assert.True(t, ValidTableName(table))
		})
	}

	for _, table := range invalidTables {
		t.Run("invalid table: "+table, func(t *testing.T) {
			assert.False(t, ValidTableName(table))
		})
	}
}

func TestValidIdentifierName(t *testing.T) {
	validIdentifiers := []string{
		"id",
		"user_id",
		"email_address",
		"firstName",
		"last_name_2",
		"_id",
	}

	invalidIdentifiers := []string{
		"invalid-identifier",
		"2_invalid",
		"invalid identifier",
		"invalid+identifier",
		"",
		"user-id",
		"user.id",
		"user;id",
	}

	for _, identifier := range validIdentifiers {
		t.Run("valid identifier: "+identifier, func(t *testing.T) {
			assert.True(t, ValidIdentifierName(identifier))
		})
	}

	for _, identifier := range invalidIdentifiers {
		t.Run("invalid identifier: "+identifier, func(t *testing.T) {
			assert.False(t, ValidIdentifierName(identifier))
		})
	}
}
