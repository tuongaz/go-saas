package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/autopus/bootstrap/pkg/auth/oauth2"
	"github.com/autopus/bootstrap/pkg/auth/oauth2/google"
	"github.com/autopus/bootstrap/pkg/baseurl"
	"github.com/autopus/bootstrap/pkg/httputil"
)

func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out, err := s.store.GetAccount(ctx, AccountID(ctx))
	httputil.HandleResponse(ctx, w, out, err)
}

func (s *Service) Oauth2AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	google.New(s.getGoogleAuthConfig(r.Context())).LoginHandler(w, r, nil)
}

func (s *Service) Oauth2LoginSignupCallbackHandler(w http.ResponseWriter, r *http.Request) {
	gauth := google.New(s.getGoogleAuthConfig(r.Context()))
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

	s.oauth2SignupLogin(w, r, *user)
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

func (s *Service) getGoogleAuthConfig(ctx context.Context) oauth2.Config {
	return oauth2.Config{
		ClientID:     s.cfg.AuthGoogleClientID,
		ClientSecret: s.cfg.AuthGoogleClientSecret,
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", baseurl.GetBaseAPI(ctx)),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
	}
}
