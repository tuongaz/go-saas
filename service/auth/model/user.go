package model

import (
	"time"
)

type User struct {
	ID                                string    `json:"id"`
	Email                             string    `json:"email"`
	Name                              string    `json:"name"`
	Password                          string    `json:"-"`
	ResetPasswordCode                 string    `json:"-"`
	ResetPasswordCodeExpiredTimestamp int64     `json:"-"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}
