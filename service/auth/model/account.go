package model

import (
	"time"
)

type Account struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Avatar             string    `json:"avatar"`
	CommunicationEmail string    `json:"communication_email"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AccountRole struct {
	ID             string    `json:"id"`
	AccountID      string    `json:"account_id"`
	OrganisationID string    `json:"organisation_id"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
