package auth

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/oauth2"
	"github.com/tuongaz/go-saas/pkg/oauth2/providers"
	"github.com/tuongaz/go-saas/store"

	"github.com/tuongaz/go-saas/pkg/httputil"
)

func (s *Service) SetupAPI(router *chi.Mux) {
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
	})
}

// MeHandler returns the account information of the current authenticated user.
func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out, err := s.store.GetAccount(ctx, AccountID(ctx))
	httputil.HandleResponse(ctx, w, out, err)
}

func (s *Service) Oauth2EnabledProvidersHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *Service) Oauth2AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *Service) Oauth2LoginSignupCallbackHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *Service) SignupHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[LoginInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	authInfo, err := s.loginUsernamePasswordAccount(ctx, input)
	httputil.HandleResponse(ctx, w, authInfo, err)
}

func (s *Service) GetResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	if code == "" {
		httputil.HandleResponse(ctx, w, nil, apierror.NewValidationError("code is required", nil, nil))
		return
	}

	req, err := s.store.GetResetPasswordRequest(ctx, code)
	if store.IsNotFoundError(err) {
		httputil.HandleResponse(ctx, w, nil, apierror.NewValidationError("reset password request not found", nil))
		return
	}
	if req.IsExpired(s.cfg.ResetPasswordRequestExpiryMinutes) {
		httputil.HandleResponse(ctx, w, nil, apierror.NewValidationError("reset password request expired", nil))
		return
	}

	httputil.HandleResponse(ctx, w, map[string]any{}, err)
}

func (s *Service) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[ResetPasswordRequestInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	err = s.resetPasswordRequest(ctx, input)
	httputil.HandleResponse(ctx, w, nil, err)
}

func (s *Service) ResetPasswordConfirmHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[ResetPasswordConfirmInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	err = s.resetPasswordConfirm(ctx, input)
	httputil.HandleResponse(ctx, w, nil, err)
}

func (s *Service) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	refreshToken := r.URL.Query().Get("refresh_token")

	authInfo, err := s.RefreshToken(ctx, refreshToken)
	httputil.HandleResponse(ctx, w, authInfo, err)
}

func (s *Service) getOauth2Config(r *http.Request) (*oauth2.Config, *config.OAuth2ProviderConfig, error) {
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
