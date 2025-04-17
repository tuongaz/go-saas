package organisation

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tuongaz/go-saas/core/auth"
	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/httputil"
)

// Service handles Organisation-related operations
type Service struct {
	store Store
}

// New creates a new Organisation service
func New(store Store) *Service {
	return &Service{
		store: store,
	}
}

// SetupAPI sets up the Organisation API routes
func (s *Service) SetupAPI(router chi.Router, authMiddleware func(http.Handler) http.Handler) {
	router.With(authMiddleware).Route("/organisations", func(r chi.Router) {
		r.Get("/", s.ListOrganisationsHandler)
		r.Post("/", s.CreateOrganisationHandler)
		r.Get("/{organisationID}", s.GetOrganisationHandler)
		r.Put("/{organisationID}", s.UpdateOrganisationHandler)
		r.Delete("/{organisationID}", s.DeleteOrganisationHandler)
		r.Post("/{organisationID}/members", s.AddOrganisationMemberHandler)
		r.Get("/{organisationID}/members", s.ListOrganisationMembersHandler)
		r.Delete("/{organisationID}/members/{accountID}", s.RemoveOrganisationMemberHandler)
		r.Put("/{organisationID}/members/{accountID}/role", s.UpdateOrganisationMemberRoleHandler)
	})
}

// ListOrganisationsHandler returns a list of Organisations for the current authenticated user
func (s *Service) ListOrganisationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := auth.AccountID(ctx)

	organisations, err := s.store.ListOrganisationsByAccountID(ctx, accountID)
	httputil.HandleResponse(ctx, w, organisations, err)
}

// CreateOrganisationHandler creates a new Organisation with the current user as owner
func (s *Service) CreateOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := auth.AccountID(ctx)

	input, err := httputil.ParseRequestBody[CreateOrganisationInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Ensure the current user is set as owner
	input.OwnerID = accountID

	organisation, err := s.store.CreateOrganisation(ctx, *input)
	httputil.HandleResponse(ctx, w, organisation, err)
}

// GetOrganisationHandler returns details of a specific Organisation
func (s *Service) GetOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has access to this Organisation
	if err := s.verifyOrganisationAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	organisation, err := s.store.GetOrganisation(ctx, organisationID)
	httputil.HandleResponse(ctx, w, organisation, err)
}

// UpdateOrganisationHandler updates an existing Organisation
func (s *Service) UpdateOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	input, err := httputil.ParseRequestBody[UpdateOrganisationInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Set the ID from the URL path parameter
	input.ID = organisationID

	organisation, err := s.store.UpdateOrganisation(ctx, *input)
	httputil.HandleResponse(ctx, w, organisation, err)
}

// DeleteOrganisationHandler deletes an Organisation
func (s *Service) DeleteOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	err := s.store.DeleteOrganisation(ctx, organisationID)
	httputil.HandleResponse(ctx, w, nil, err)
}

// AddOrganisationMemberHandler adds a new member to an Organisation
func (s *Service) AddOrganisationMemberHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	input, err := httputil.ParseRequestBody[AddOrganisationMemberInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Set the Organisation ID from the URL path parameter
	input.OrganisationID = organisationID

	member, err := s.store.AddOrganisationMember(ctx, *input)
	httputil.HandleResponse(ctx, w, member, err)
}

// ListOrganisationMembersHandler lists all members of an Organisation
func (s *Service) ListOrganisationMembersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has access to this Organisation
	if err := s.verifyOrganisationAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	members, err := s.store.ListOrganisationMembers(ctx, organisationID)
	httputil.HandleResponse(ctx, w, members, err)
}

// RemoveOrganisationMemberHandler removes a member from an Organisation
func (s *Service) RemoveOrganisationMemberHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")
	accountID := chi.URLParam(r, "accountID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	err := s.store.RemoveOrganisationMember(ctx, organisationID, accountID)
	httputil.HandleResponse(ctx, w, nil, err)
}

// UpdateOrganisationMemberRoleHandler updates the role of a member in an Organisation
func (s *Service) UpdateOrganisationMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")
	accountID := chi.URLParam(r, "accountID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	input, err := httputil.ParseRequestBody[UpdateOrganisationMemberRoleInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Set the Organisation ID and Account ID from the URL path parameters
	input.OrganisationID = organisationID
	input.AccountID = accountID

	member, err := s.store.UpdateOrganisationMemberRole(ctx, *input)
	httputil.HandleResponse(ctx, w, member, err)
}

// Helper functions for Organisation access control

// verifyOrganisationAccess verifies that the current user has access to the Organisation
func (s *Service) verifyOrganisationAccess(ctx context.Context, organisationID string) error {
	accountID := auth.AccountID(ctx)

	// Check if the user is a member of the Organisation
	_, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, organisationID, accountID)
	if err != nil {
		return apierror.NewForbiddenError("you do not have access to this Organisation", err)
	}

	return nil
}

// verifyOrganisationOwnerAccess verifies that the current user has owner access to the Organisation
func (s *Service) verifyOrganisationOwnerAccess(ctx context.Context, organisationID string) error {
	accountID := auth.AccountID(ctx)

	// Check if the user is an owner of the Organisation
	accRole, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, organisationID, accountID)
	if err != nil {
		return apierror.NewForbiddenError("you do not have access to this Organisation", err)
	}

	if model.Role(accRole.Role) != model.RoleOwner {
		return apierror.NewForbiddenError("you do not have owner access to this Organisation", nil)
	}

	return nil
}
