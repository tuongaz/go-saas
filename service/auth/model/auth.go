package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthProviderUsernamePassword = "username_password"
)

type AccessToken struct {
	ID             string    `json:"id" mapstructure:"id"`
	AccountRoleID  string    `json:"account_role_id" mapstructure:"account_role_id"`
	RefreshToken   string    `json:"refresh_token" mapstructure:"refresh_token"`
	Device         string    `json:"device" mapstructure:"device"`
	ProviderUserID string    `json:"provider_user_id" mapstructure:"provider_user_id"`
	LastAccessedAt string    `json:"last_access_at" mapstructure:"last_access_at"`
	CreatedAt      time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" mapstructure:"updated_at"`
}

type CustomClaims struct {
	Organisation string `json:"organisation"`
	AccountType  string `json:"account_type"`
	jwt.RegisteredClaims
}

type AuthenticatedInfo struct {
	Token        string `json:"token"`
	Type         string `json:"type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type Provider struct {
	Key         string
	Secret      string
	CallbackURL string
}

type LoginProvider struct {
	ID             string    `json:"id" mapstructure:"id"`
	AccountID      string    `json:"account_id" mapstructure:"account_id"`
	Provider       string    `json:"provider" mapstructure:"provider"`
	Email          string    `json:"email" mapstructure:"email"`
	Name           string    `json:"name" mapstructure:"name"`
	FirstName      string    `json:"first_name" mapstructure:"first_name"`
	LastName       string    `json:"last_name" mapstructure:"last_name"`
	Avatar         string    `json:"avatar" mapstructure:"avatar"`
	ProviderUserID string    `json:"provider_user_id" mapstructure:"provider_user_id"`
	LastLogin      time.Time `json:"last_login" mapstructure:"last_login"`
	CreatedAt      time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" mapstructure:"updated_at"`
}

type OAuthClient struct {
	ID           string
	ClientID     string
	ClientSecret string
	Domain       string
}
