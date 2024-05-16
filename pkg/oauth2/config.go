package oauth2

import (
	"golang.org/x/oauth2"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type AuthDetail struct {
	Token       oauth2.Token `json:"token"`
	State       any          `json:"state"`
	RedirectURL string       `json:"redirect_url"`
}
