package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/core/auth/store"
	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/oauth2"
	"github.com/tuongaz/go-saas/pkg/oauth2/providers"
	coreStore "github.com/tuongaz/go-saas/store"

	"github.com/tuongaz/go-saas/pkg/httputil"
)

func (s *service) SetupAPI(router *chi.Mux) {
	authMiddleware := s.NewMiddleware()
	deviceMiddleware := s.NewDeviceMiddleware()

	router.Use(deviceMiddleware)
	router.Route("/auth", func(r chi.Router) {
		// public routes
		r.Get("/oauth2-providers", s.Oauth2EnabledProvidersHandler)
		r.Post("/signup", s.SignupHandler)
		r.Post("/reset-password", s.ResetPasswordHandler)
		r.Get("/reset-password", s.GetResetPasswordHandler)
		r.Post("/reset-password-confirm", s.ResetPasswordConfirmHandler)
		r.Post("/login", s.LoginHandler)
		r.Post("/token", s.RefreshTokenHandler) // deprecated
		r.Get("/token", s.RefreshTokenHandler)
		r.Get("/{provider}", s.Oauth2AuthenticateHandler)
		r.Get("/{provider}/callback", s.Oauth2LoginSignupCallbackHandler)

		// private routes
		r.With(authMiddleware).Get("/me", s.MeHandler)
		r.With(authMiddleware).Post("/change-password", s.ChangePasswordHandler)
		r.With(authMiddleware).Put("/account", s.UpdateAccountHandler)

		// Organisation routes - use lowercase in URLs
		r.With(authMiddleware).Route("/organisations", func(r chi.Router) {
			r.Get("/", s.ListOrganisationsHandler)
			r.Post("/", s.CreateOrganisationHandler)
			r.Get("/{organisationID}", s.GetOrganisationHandler)
			r.Put("/{organisationID}", s.UpdateOrganisationHandler)
			r.Post("/{organisationID}/members", s.AddOrganisationMemberHandler)
			r.Get("/{organisationID}/members", s.ListOrganisationMembersHandler)
			r.Delete("/{organisationID}/members/{accountID}", s.RemoveOrganisationMemberHandler)
			r.Put("/{organisationID}/members/{accountID}/role", s.UpdateOrganisationMemberRoleHandler)
		})
	})
}

// MeHandler returns the account information of the current authenticated user.
func (s *service) MeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out, err := s.store.GetAccount(ctx, AccountID(ctx))
	httputil.HandleResponse(ctx, w, out, err)
}

func (s *service) Oauth2EnabledProvidersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	enabledProviders := make([]string, 0, len(s.providers))
	for provider := range s.providers {
		enabledProviders = append(enabledProviders, provider)
	}
	httputil.HandleResponse(ctx, w, map[string]any{
		"providers": enabledProviders,
	}, nil)
}

// Oauth2AuthenticateHandler redirects the user to the OAuth2 provider's login page.
func (s *service) Oauth2AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	oauth2Config, _, err := s.getOauth2Config(r)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	provider := providers.GetProvider(chi.URLParam(r, "provider"), *oauth2Config)
	if provider == nil {
		httputil.HandleResponse(r.Context(), w, nil, fmt.Errorf("provider not found"))
		return
	}

	provider.LoginHandler(w, r, nil)
}

// Oauth2LoginSignupCallbackHandler handles the callback from the OAuth2 provider after the user has authenticated.
func (s *service) Oauth2LoginSignupCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauth2Config, oauth2Provider, err := s.getOauth2Config(r)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	provider := providers.GetProvider(chi.URLParam(r, "provider"), *oauth2Config)
	if provider == nil {
		httputil.HandleResponse(r.Context(), w, nil, fmt.Errorf("provider not found"))
		return
	}

	authDetail, err := provider.CallbackHandler(w, r)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	user, err := provider.GetUser(r.Context(), &authDetail.Token)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	s.oauth2SignupLogin(w, r, *oauth2Provider, *user)
}

// SignupHandler creates a new account with a username and password.
func (s *service) SignupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[SignupInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	authInfo, err := s.signupUsernamePasswordAccount(ctx, input)
	httputil.HandleResponse(ctx, w, authInfo, err)
}

// LoginHandler logs in an account with a username and password.
func (s *service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[LoginInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	authInfo, err := s.loginUsernamePasswordAccount(ctx, input)
	httputil.HandleResponse(ctx, w, authInfo, err)
}

func (s *service) GetResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	if code == "" {
		httputil.HandleResponse(ctx, w, nil, apierror.NewValidationError("Missing reset password code", nil))
		return
	}

	resetPasswordReq, err := s.store.GetResetPasswordRequest(ctx, code)
	if err != nil {
		if coreStore.IsNotFoundError(err) {
			httputil.HandleResponse(ctx, w, nil, apierror.NewValidationError("Invalid reset password code", nil))
			return
		}
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	if resetPasswordReq.IsExpired(s.cfg.ResetPasswordRequestExpiryMinutes) {
		httputil.HandleResponse(ctx, w, nil, apierror.NewValidationError("reset password request expired", nil))
		return
	}

	httputil.HandleResponse(ctx, w, map[string]any{}, err)
}

func (s *service) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[ResetPasswordRequestInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	err = s.resetPasswordRequest(ctx, input)
	httputil.HandleResponse(ctx, w, nil, err)
}

func (s *service) ResetPasswordConfirmHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[ResetPasswordConfirmInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	err = s.resetPasswordConfirm(ctx, input)
	httputil.HandleResponse(ctx, w, nil, err)
}

func (s *service) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	refreshToken := r.URL.Query().Get("refresh_token")

	authInfo, err := s.RefreshToken(ctx, refreshToken)
	httputil.HandleResponse(ctx, w, authInfo, err)
}

type UpdateAccountInput struct {
	Name               string `json:"name" validate:"required"`
	CommunicationEmail string `json:"communication_email" validate:"required,email"`
	Avatar             string `json:"avatar"`
}

func (s *service) getOauth2Config(r *http.Request) (*oauth2.Config, *config.OAuth2ProviderConfig, error) {
	providerName := chi.URLParam(r, "provider")
	oauthProvider, ok := s.providers[providerName]
	if !ok {
		return nil, nil, fmt.Errorf("provider not found: %s", providerName)
	}

	return &oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Scopes:       oauthProvider.Scopes,
	}, &oauthProvider, nil
}

// ListOrganisationsHandler returns a list of Organisations for the current authenticated user
func (s *service) ListOrganisationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := AccountID(ctx)

	organisations, err := s.store.ListOrganisationsByAccountID(ctx, accountID)
	httputil.HandleResponse(ctx, w, organisations, err)
}

// CreateOrganisationHandler creates a new Organisation with the current user as owner
func (s *service) CreateOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := AccountID(ctx)

	input, err := httputil.ParseRequestBody[store.CreateOrganisationInput](r)
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
func (s *service) GetOrganisationHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *service) UpdateOrganisationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	input, err := httputil.ParseRequestBody[store.UpdateOrganisationInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Set the ID from the URL path parameter
	input.ID = organisationID

	organisation, err := s.store.UpdateOrganisation(ctx, *input)
	httputil.HandleResponse(ctx, w, organisation, err)
}

// AddOrganisationMemberHandler adds a new member to an Organisation
func (s *service) AddOrganisationMemberHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	input, err := httputil.ParseRequestBody[store.AddOrganisationMemberInput](r)
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
func (s *service) ListOrganisationMembersHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *service) RemoveOrganisationMemberHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *service) UpdateOrganisationMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organisationID := chi.URLParam(r, "organisationID")
	accountID := chi.URLParam(r, "accountID")

	// Verify the user has owner access to this Organisation
	if err := s.verifyOrganisationOwnerAccess(ctx, organisationID); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	input, err := httputil.ParseRequestBody[store.UpdateOrganisationMemberRoleInput](r)
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
func (s *service) verifyOrganisationAccess(ctx context.Context, organisationID string) error {
	accountID := AccountID(ctx)

	// Check if the user is a member of the Organisation
	_, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, organisationID, accountID)
	if err != nil {
		return apierror.NewForbiddenError("you do not have access to this Organisation", err)
	}

	return nil
}

// verifyOrganisationOwnerAccess verifies that the current user has owner access to the Organisation
func (s *service) verifyOrganisationOwnerAccess(ctx context.Context, organisationID string) error {
	accountID := AccountID(ctx)

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

// ChangePasswordHandler handles the change password request for authenticated users
func (s *service) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the account ID from the context
	accountID := AccountID(ctx)
	if accountID == "" {
		httputil.HandleResponse(ctx, w, nil, apierror.NewUnauthorizedErr("not authenticated", nil))
		return
	}

	// Parse request body
	input, err := httputil.ParseRequestBody[ChangePasswordInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Change password
	err = s.changePassword(ctx, accountID, input)
	httputil.HandleResponse(ctx, w, map[string]any{"success": err == nil}, err)
}

// UpdateAccountHandler updates the authenticated user's account information
func (s *service) UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[UpdateAccountInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	accountID := AccountID(ctx)
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	// Update the account fields
	account.Name = input.Name
	account.CommunicationEmail = input.CommunicationEmail
	account.Avatar = input.Avatar

	updatedAccount, err := s.store.UpdateAccount(ctx, accountID, account)
	httputil.HandleResponse(ctx, w, updatedAccount, err)
}
