package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"

	"github.com/tuongaz/go-saas/pkg/errors"
	"github.com/tuongaz/go-saas/pkg/httputil"
	"github.com/tuongaz/go-saas/pkg/types"
)

type Client struct {
	ID           string
	Secret       string
	RedirectURIs []string
}

var clients = map[string]Client{
	"openai": {ID: "openai", Secret: "openai"},
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type RefreshToken struct {
	Token     string
	ClientID  string
	ExpiresIn int64
}

func (s *Service) TokenAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state := &authState{}
	stateBase64 := r.URL.Query().Get("state")
	stateStr, err := base64.StdEncoding.DecodeString(stateBase64)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	if err := json.Unmarshal(stateStr, state); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	if _, ok := clients[state.ClientID]; !ok {
		httputil.HandleResponse(ctx, w, nil, fmt.Errorf("invalid client id"))
		return
	}

	claims, err := s.signer.ParseCustomClaims(state.Token)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	accRole, err := s.GetAccountRole(ctx, claims.Organisation, claims.Subject)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	authCode, err := s.createAuthCode(accRole.ID)
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	params := url.Values{}
	params.Add("code", authCode)
	params.Add("state", state.State)

	redirectURL := state.RedirectURI + "?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (s *Service) AuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	grantType := r.FormValue("grant_type")
	if grantType == "" {
		grantType = "refresh_token"
	}

	if grantType == "refresh_token" {
		refreshToken := r.FormValue("refresh_token")
		if refreshToken == "" {
			bM := types.M{}
			if err := json.NewDecoder(r.Body).Decode(&bM); err != nil {
				httputil.HandleResponse(ctx, w, nil, err)
				return
			}
			refreshToken = bM["refresh_token"].(string)
			if refreshToken == "" {
				httputil.HandleResponse(ctx, w, nil, errors.NewValidationError(fmt.Errorf("refresh token is required")))
				return
			}
		}

		authInfo, err := s.RefreshToken(r.Context(), refreshToken)
		httputil.HandleResponse(ctx, w, authInfo, err)
		return
	}

	if grantType == "authorization_code" {
		client, ok := clients[r.FormValue("client_id")]
		if !ok {
			httputil.HandleResponse(ctx, w, nil, fmt.Errorf("invalid client id"))
			return
		}

		if r.FormValue("client_secret") != client.Secret {
			httputil.HandleResponse(ctx, w, nil, fmt.Errorf("invalid client secret"))
			return
		}

		authCodeEncrypted := r.FormValue("code")
		authCode, err := s.encryptor.Decrypt(authCodeEncrypted)
		if err != nil {
			httputil.HandleResponse(ctx, w, nil, err)
			return
		}

		authCodeM := types.M{}
		if err := json.Unmarshal([]byte(authCode), &authCodeM); err != nil {
			httputil.HandleResponse(ctx, w, nil, err)
			return
		}

		accRole, err := s.store.GetAccountRoleByID(ctx, authCodeM["account_role_id"].(string))
		if err != nil {
			httputil.HandleResponse(ctx, w, nil, err)
			return
		}

		authToken, err := s.NewToken(r.Context(), accRole)
		if err != nil {
			httputil.HandleResponse(ctx, w, nil, err)
			return
		}

		token := &oauth2.Token{
			AccessToken:  authToken.Token,
			TokenType:    "Bearer",
			RefreshToken: authToken.RefreshToken,
			Expiry:       time.Now().Add(time.Duration(authToken.ExpiresIn) * time.Second),
		}

		httputil.HandleResponse(ctx, w, token, nil)
		return
	}

	httputil.HandleResponse(ctx, w, nil, errors.NewUnauthorizedErr(fmt.Errorf("invalid grant type")))
}

func (s *Service) createAuthCode(accountRoleID string) (string, error) {
	authCodeM := types.M{
		"account_role_id": accountRoleID,
		"expires_at":      time.Now().Add(time.Minute * 5),
	}

	authCode, err := json.Marshal(authCodeM)
	if err != nil {
		return "", fmt.Errorf("unable to marshal auth code: %w", err)
	}

	code, err := s.encryptor.Encrypt(string(authCode))
	if err != nil {
		return "", fmt.Errorf("unable to encrypt auth code: %w", err)
	}

	return code, nil
}

type authState struct {
	Token        string `json:"token"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	State        string `json:"state"`
}
