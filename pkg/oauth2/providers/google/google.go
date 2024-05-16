package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	oauth22 "github.com/tuongaz/go-saas/pkg/oauth2"
	goauth "golang.org/x/oauth2"
)

const Name = "google"

type Google struct {
	oauth2 *oauth22.OAuth2
}

func New(cfg oauth22.Config) *Google {
	return &Google{
		oauth2: oauth22.New(&goauth.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint: goauth.Endpoint{
				AuthURL:   "https://accounts.google.com/o/oauth2/auth",
				TokenURL:  "https://oauth2.googleapis.com/token",
				AuthStyle: goauth.AuthStyleInParams,
			},
		}),
	}
}

func (g *Google) LoginHandler(w http.ResponseWriter, r *http.Request, state map[string]any) {
	g.oauth2.LoginHandler(w, r, state)
}

func (g *Google) CallbackHandler(w http.ResponseWriter, r *http.Request) (*oauth22.AuthDetail, error) {
	return g.oauth2.CallbackHandler(w, r)
}

func (g *Google) GetUser(ctx context.Context, token *goauth.Token) (*oauth22.User, error) {
	client := goauth.NewClient(ctx, goauth.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &oauth22.User{
		UserID:       fmt.Sprint(data["id"]),
		Name:         fmt.Sprint(data["name"]),
		Email:        fmt.Sprint(data["email"]),
		FirstName:    fmt.Sprint(data["given_name"]),
		LastName:     fmt.Sprint(data["family_name"]),
		AvatarURL:    fmt.Sprint(data["picture"]),
		Location:     fmt.Sprint(data["locale"]),
		RawData:      data,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      token.Extra("id_token").(string),
		ExpiresAt:    token.Expiry,
		Provider:     Name,
	}, nil
}
