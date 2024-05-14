package auth

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/autopus/bootstrap/pkg/auth/oauth2"
	"github.com/autopus/bootstrap/pkg/auth/oauth2/google"
	"github.com/autopus/bootstrap/pkg/httputil"
)

func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out, err := s.store.GetAccount(ctx, AccountID(ctx))
	httputil.HandleResponse(ctx, w, out, err)
}

func (s *Service) Oauth2AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	oauth2Config, _, err := s.getOauth2Config(r)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	// TODO: Support other providers than Google
	google.New(*oauth2Config).LoginHandler(w, r, nil)
}

func (s *Service) Oauth2LoginSignupCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauth2Config, oauth2Provider, err := s.getOauth2Config(r)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	gauth := google.New(*oauth2Config)
	detail, err := gauth.CallbackHandler(w, r)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	user, err := gauth.GetUser(r.Context(), &detail.Token)
	if err != nil {
		httputil.HandleResponse(r.Context(), w, nil, err)
		return
	}

	// TODO: Support other providers than Google
	s.oauth2SignupLogin(w, r, *oauth2Provider, *user)
}

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

type TokenRequestInput struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) getOauth2Config(r *http.Request) (*oauth2.Config, *provider, error) {
	providerName := chi.URLParam(r, "provider")
	oauthProvider, ok := s.providers[providerName]
	if !ok {
		return nil, nil, fmt.Errorf("provider not found: %s", providerName)
	}

	return &oauth2.Config{
		ClientID:     oauthProvider.clientID,
		ClientSecret: oauthProvider.clientSecret,
		RedirectURL:  oauthProvider.redirectURL,
		Scopes:       oauthProvider.scopes,
	}, &oauthProvider, nil
}
