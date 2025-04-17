package store

import (
	"context"
	"fmt"

	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/pkg/timer"
	"github.com/tuongaz/go-saas/pkg/uid"
	"github.com/tuongaz/go-saas/store"
	"github.com/tuongaz/go-saas/store/types"
)

// ListOrganisationsByAccountID returns all organisations that the account is a member of
func (s *Store) ListOrganisationsByAccountID(ctx context.Context, accountID string) ([]model.Organisation, error) {
	query := `
		SELECT o.* FROM organisation o
		JOIN organisation_account_role oar ON o.id = oar.organisation_id
		WHERE oar.account_id = $1 
		ORDER BY o.created_at DESC
	`
	var organisations []model.Organisation
	if err := s.store.DB().SelectContext(ctx, &organisations, query, accountID); err != nil {
		return nil, fmt.Errorf("list organisations by account id: %w", err)
	}

	return organisations, nil
}

// CreateOrganisation creates a new organisation and adds the owner as a member
func (s *Store) CreateOrganisation(ctx context.Context, input CreateOrganisationInput) (*model.Organisation, error) {
	tx, err := s.store.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Create the organisation
	organisationID := uid.ID()
	now := timer.Now()

	// Create organisation record with required fields
	orgData := types.Record{
		"id":          organisationID,
		"name":        input.Name,
		"owner_id":    input.OwnerID,
		"avatar":      "",
		"description": "",
		"metadata":    "{}",
		"created_at":  now,
		"updated_at":  now,
	}

	// Add optional fields only if they're provided (non-nil pointers)
	if input.Description != nil {
		orgData["description"] = *input.Description
	}

	if input.Avatar != nil {
		orgData["avatar"] = *input.Avatar
	}

	if input.Metadata != nil {
		orgData["metadata"] = *input.Metadata
	}

	orgRecord, err := s.store.Collection(TableOrganisation).CreateRecord(ctx, orgData)
	if err != nil {
		return nil, fmt.Errorf("create organisation: %w", err)
	}

	// Add the owner as a member with owner role
	_, err = s.store.Collection(tableOrganisationAccountRole).CreateRecord(ctx, types.Record{
		"id":              uid.ID(),
		"organisation_id": organisationID,
		"account_id":      input.OwnerID,
		"role":            string(model.RoleOwner),
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return nil, fmt.Errorf("add owner to organisation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	organisation := &model.Organisation{}
	if err := orgRecord.Decode(organisation); err != nil {
		return nil, err
	}

	return organisation, nil
}

// GetOrganisation returns an organisation by ID
func (s *Store) GetOrganisation(ctx context.Context, organisationID string) (*model.Organisation, error) {
	record, err := s.store.Collection(TableOrganisation).FindOne(ctx, store.Filter{"id": organisationID})
	if err != nil {
		return nil, fmt.Errorf("get organisation: %w", err)
	}

	organisation := &model.Organisation{}
	if err := record.Decode(organisation); err != nil {
		return nil, err
	}

	return organisation, nil
}

// UpdateOrganisation updates an organisation's details
func (s *Store) UpdateOrganisation(ctx context.Context, input UpdateOrganisationInput) (*model.Organisation, error) {
	// Get the current organisation to handle partial updates
	_, err := s.GetOrganisation(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("get organisation for update: %w", err)
	}

	// Create update record with only the fields that are provided
	updateRecord := types.Record{
		"updated_at": timer.Now(),
	}

	// Only update fields that are provided (non-nil pointers)
	if input.Name != nil {
		updateRecord["name"] = *input.Name
	}

	if input.Description != nil {
		updateRecord["description"] = *input.Description
	}

	if input.Avatar != nil {
		updateRecord["avatar"] = *input.Avatar
	}

	if input.Metadata != nil {
		updateRecord["metadata"] = *input.Metadata
	}

	record, err := s.store.Collection(TableOrganisation).UpdateRecord(ctx, input.ID, updateRecord)
	if err != nil {
		return nil, fmt.Errorf("update organisation: %w", err)
	}

	updatedOrganisation := &model.Organisation{}
	if err := record.Decode(updatedOrganisation); err != nil {
		return nil, err
	}

	return updatedOrganisation, nil
}

// DeleteOrganisation deletes an organisation and all its members
func (s *Store) DeleteOrganisation(ctx context.Context, organisationID string) error {
	tx, err := s.store.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Delete all members
	err = tx.Exec(ctx, "DELETE FROM organisation_account_role WHERE organisation_id = $1", organisationID)
	if err != nil {
		return fmt.Errorf("delete organisation members: %w", err)
	}

	// Delete the organisation
	if err := s.store.Collection(TableOrganisation).DeleteRecord(ctx, organisationID); err != nil {
		return fmt.Errorf("delete organisation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// AddOrganisationMember adds a member to an organisation
func (s *Store) AddOrganisationMember(ctx context.Context, input AddOrganisationMemberInput) (*model.AccountRole, error) {
	// Check if the member already exists
	exists, err := s.store.Collection(tableOrganisationAccountRole).Exists(ctx, store.Filter{
		"organisation_id": input.OrganisationID,
		"account_id":      input.AccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("check if member exists: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("member already exists in this organisation")
	}

	// Add the member
	now := timer.Now()
	record, err := s.store.Collection(tableOrganisationAccountRole).CreateRecord(ctx, types.Record{
		"id":              uid.ID(),
		"organisation_id": input.OrganisationID,
		"account_id":      input.AccountID,
		"role":            input.Role,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return nil, fmt.Errorf("add member to organisation: %w", err)
	}

	accountRole := &model.AccountRole{}
	if err := record.Decode(accountRole); err != nil {
		return nil, err
	}

	return accountRole, nil
}

// ListOrganisationMembers returns all members of an organisation
func (s *Store) ListOrganisationMembers(ctx context.Context, organisationID string) ([]model.AccountRole, error) {
	records, err := s.store.Collection(tableOrganisationAccountRole).Find(
		ctx,
		store.WithFilter(store.Filter{"organisation_id": organisationID}),
	)
	if err != nil {
		return nil, fmt.Errorf("list organisation members: %w", err)
	}

	var members []model.AccountRole
	if err := records.Decode(&members); err != nil {
		return nil, err
	}

	return members, nil
}

// RemoveOrganisationMember removes a member from an organisation
func (s *Store) RemoveOrganisationMember(ctx context.Context, organisationID, accountID string) error {
	// Check if the member exists
	record, err := s.store.Collection(tableOrganisationAccountRole).FindOne(ctx, store.Filter{
		"organisation_id": organisationID,
		"account_id":      accountID,
	})
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	// Check if the member is the owner
	accountRole := &model.AccountRole{}
	if err := record.Decode(accountRole); err != nil {
		return err
	}

	if model.Role(accountRole.Role) == model.RoleOwner {
		return fmt.Errorf("cannot remove the owner from the organisation")
	}

	// Remove the member
	if err := s.store.Collection(tableOrganisationAccountRole).DeleteRecord(ctx, accountRole.ID); err != nil {
		return fmt.Errorf("remove member from organisation: %w", err)
	}

	return nil
}

// UpdateOrganisationMemberRole updates a member's role in an organisation
func (s *Store) UpdateOrganisationMemberRole(ctx context.Context, input UpdateOrganisationMemberRoleInput) (*model.AccountRole, error) {
	// Check if the member exists
	record, err := s.store.Collection(tableOrganisationAccountRole).FindOne(ctx, store.Filter{
		"organisation_id": input.OrganisationID,
		"account_id":      input.AccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}

	// Check if the member is the owner
	accountRole := &model.AccountRole{}
	if err := record.Decode(accountRole); err != nil {
		return nil, err
	}

	if model.Role(accountRole.Role) == model.RoleOwner {
		return nil, fmt.Errorf("cannot change the role of the owner")
	}

	// Update the member's role
	record, err = s.store.Collection(tableOrganisationAccountRole).UpdateRecord(ctx, accountRole.ID, types.Record{
		"role":       input.Role,
		"updated_at": timer.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("update member role: %w", err)
	}

	if err := record.Decode(accountRole); err != nil {
		return nil, err
	}

	return accountRole, nil
}
