package model

import (
	"time"
)

type Organisation struct {
	ID        string    `json:"id" mapstructure:"id"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
