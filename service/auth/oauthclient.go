package auth

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
)

type OAuthClientManager struct{}

func (o *OAuthClientManager) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	list := map[string]oauth2.ClientInfo{
		"openai": &models.Client{
			ID:     "openai",
			Domain: "openai.com",
		},
		"22222222": &models.Client{
			ID:     "22222222",
			Domain: "localhost:9094",
		},
	}

	return list[id], nil
}
