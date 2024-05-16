package providers

import (
	oauth22 "github.com/tuongaz/go-saas/pkg/oauth2"
	"github.com/tuongaz/go-saas/pkg/oauth2/providers/google"
)

func GetProvider(name string, cfg oauth22.Config) oauth22.Provider {
	switch name {
	case google.Name:
		return google.New(cfg)
	}

	return nil
}
