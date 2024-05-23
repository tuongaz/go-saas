package model

import (
	"time"
)

type LoginCredentialsUser struct {
	ID                                string    `json:"id"`
	Email                             string    `json:"email"`
	Name                              string    `json:"name"`
	Password                          string    `json:"password"`
	ResetPasswordCode                 string    `json:"reset_password_code"`
	ResetPasswordCodeExpiredTimestamp int64     `json:"reset_password_code_expired_timestamp"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}
