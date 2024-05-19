package model

import (
	"time"
)

type LoginCredentialsUser struct {
	ID                                string    `json:"id" mapstructure:"id"`
	Email                             string    `json:"email" mapstructure:"email"`
	Name                              string    `json:"name" mapstructure:"name"`
	Password                          string    `json:"-" mapstructure:"password"`
	ResetPasswordCode                 string    `json:"-" mapstructure:"reset_password_code"`
	ResetPasswordCodeExpiredTimestamp int64     `json:"-" mapstructure:"reset_password_code_expired_timestamp"`
	CreatedAt                         time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at" mapstructure:"updated_at`
}
