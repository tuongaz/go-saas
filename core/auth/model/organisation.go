package model

import (
	"encoding/json"
	"time"
)

// Organisation represents an organisation in the system
type Organisation struct {
	ID          string           `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"`
	Description *string          `json:"description" db:"description"`
	Avatar      *string          `json:"avatar" db:"avatar"`
	Metadata    *json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	OwnerID     string           `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

type Metadata map[string]any

func (m *Metadata) Get(key string) (any, bool) {
	if m == nil {
		return nil, false
	}
	value, exists := (*m)[key]
	return value, exists
}

// support GetString, GetInt, GetBool, GetFloat64
func (m *Metadata) GetString(key string) (string, bool) {
	value, exists := m.Get(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

func (m *Metadata) GetInt(key string) (int, bool) {
	value, exists := m.Get(key)
	if !exists {
		return 0, false
	}
	i, ok := value.(int)
	return i, ok
}

func (m *Metadata) GetBool(key string) (bool, bool) {
	value, exists := m.Get(key)
	if !exists {
		return false, false
	}
	b, ok := value.(bool)
	return b, ok
}

func (m *Metadata) GetFloat64(key string) (float64, bool) {
	value, exists := m.Get(key)
	if !exists {
		return 0, false
	}
	f, ok := value.(float64)
	return f, ok
}

func (o *Organisation) GetMetadata() (Metadata, error) {
	metadata := Metadata{}
	if o.Metadata != nil {
		if err := json.Unmarshal(*o.Metadata, &metadata); err != nil {
			return nil, err
		}
	}
	return metadata, nil
}

const (
	// RoleAdmin is the admin role of an organisation
	RoleAdmin Role = "admin"
	// RoleMember is the member role of an organisation
	RoleMember Role = "member"
)
