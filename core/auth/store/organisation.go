package store

import "encoding/json"

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
