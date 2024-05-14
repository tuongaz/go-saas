package auth

import (
	"github.com/tuongaz/go-saas/pkg/hooks"
)

type OnAccountCreatedEvent struct {
	AccountID      string
	OrganisationID string
}

func (s *Service) OnAccountCreated() *hooks.Hook[*OnAccountCreatedEvent] {
	return s.onAccountCreated
}
