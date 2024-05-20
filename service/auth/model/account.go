package model

import (
	"time"
)

type Account struct {
	ID                 string    `json:"id" mapstructure:"id"`
	Name               string    `json:"name" mapstructure:"name"`
	FirstName          string    `json:"first_name" mapstructure:"first_name"`
	LastName           string    `json:"last_name" mapstructure:"last_name"`
	Avatar             string    `json:"avatar" mapstructure:"avatar"`
	CommunicationEmail string    `json:"communication_email" mapstructure:"communication_email"`
	CreatedAt          time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" mapstructure:"updated_at"`
}

type AccountRole struct {
	ID             string    `json:"id" mapstructure:"id"`
	AccountID      string    `json:"account_id" mapstructure:"account_id"`
	OrganisationID string    `json:"organisation_id" mapstructure:"organisation_id"`
	Role           string    `json:"role" mapstructure:"role"`
	CreatedAt      time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" mapstructure:"updated_at"`
}
