package organisation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tuongaz/go-saas/core/auth/model"
)

// Store defines the interface for Organisation-related data operations
type Store interface {
	// Organisation operations
	ListOrganisationsByAccountID(ctx context.Context, accountID string) ([]model.Organisation, error)
	CreateOrganisation(ctx context.Context, input CreateOrganisationInput) (*model.Organisation, error)
	GetOrganisation(ctx context.Context, organisationID string) (*model.Organisation, error)
	UpdateOrganisation(ctx context.Context, input UpdateOrganisationInput) (*model.Organisation, error)
	DeleteOrganisation(ctx context.Context, organisationID string) error

	// Organisation member operations
	AddOrganisationMember(ctx context.Context, input AddOrganisationMemberInput) (*model.AccountRole, error)
	ListOrganisationMembers(ctx context.Context, organisationID string) ([]model.AccountRole, error)
	RemoveOrganisationMember(ctx context.Context, organisationID, accountID string) error
	UpdateOrganisationMemberRole(ctx context.Context, input UpdateOrganisationMemberRoleInput) (*model.AccountRole, error)
	GetAccountRoleByOrgAndAccountID(ctx context.Context, organisationID, accountID string) (*model.AccountRole, error)
}

// CreateOrganisationInput defines the input for creating a new Organisation
type CreateOrganisationInput struct {
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Avatar      *string          `json:"avatar,omitempty"`
	Metadata    *json.RawMessage `json:"metadata,omitempty"`
	OwnerID     string           `json:"owner_id"`
}

// UpdateOrganisationInput defines the input for updating an Organisation
type UpdateOrganisationInput struct {
	ID          string           `json:"id"`
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Avatar      *string          `json:"avatar,omitempty"`
	Metadata    *json.RawMessage `json:"metadata,omitempty"`
}

// AddOrganisationMemberInput defines the input for adding a member to an Organisation
type AddOrganisationMemberInput struct {
	OrganisationID string `json:"organisation_id"`
	AccountID      string `json:"account_id"`
	Role           string `json:"role"`
}

// UpdateOrganisationMemberRoleInput defines the input for updating a member's role
type UpdateOrganisationMemberRoleInput struct {
	OrganisationID string `json:"organisation_id"`
	AccountID      string `json:"account_id"`
	Role           string `json:"role"`
}

// OrganisationMember represents a member of an Organisation with their role
type OrganisationMember struct {
	AccountID      string    `json:"account_id"`
	OrganisationID string    `json:"organisation_id"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
