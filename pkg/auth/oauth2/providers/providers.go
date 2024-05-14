package providers

import (
	"github.com/tuongaz/go-saas/pkg/auth/oauth2"
	"github.com/tuongaz/go-saas/pkg/auth/oauth2/providers/google"
)

func GetProvider(name string, cfg oauth2.Config) oauth2.Provider {
	switch name {
	case google.Name:
		return google.New(cfg)
	}

	return nil
}
