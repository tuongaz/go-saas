package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthProviderUsernamePassword = "username_password"
)

type AccessToken struct {
	ID             string    `json:"id"`
	AccountRoleID  string    `json:"account_role_id"`
	RefreshToken   string    `json:"refresh_token"`
	Device         string    `json:"device"`
	ProviderUserID string    `json:"provider_user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	ID             string    `json:"id"`
	AccountID      string    `json:"account_id"`
	Provider       string    `json:"provider"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Avatar         string    `json:"avatar"`
	ProviderUserID string    `json:"provider_user_id"`
	LastLogin      time.Time `json:"last_login"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OAuthClient struct {
	ID           string
	ClientID     string
	ClientSecret string
	Domain       string
}
