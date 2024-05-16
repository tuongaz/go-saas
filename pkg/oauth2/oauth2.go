package oauth2

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	goauth "golang.org/x/oauth2"
)

const (
	stateKey        = "oauth2_state"
	codeVerifierKey = "oauth2_code_verifier"
	nonceKey        = "oauth2_nonce"
	redirectURLKey  = "oauth2_redirect_url"
)

type Provider interface {
	GetUser(ctx context.Context, token *goauth.Token) (*User, error)
	LoginHandler(w http.ResponseWriter, r *http.Request, state map[string]any)
	CallbackHandler(w http.ResponseWriter, r *http.Request) (*AuthDetail, error)
}

type OAuth2 struct {
	cfg *oauth2.Config
}

func New(cfg *oauth2.Config) *OAuth2 {
	return &OAuth2{
		cfg: cfg,
	}
}

func (o *OAuth2) LoginHandler(w http.ResponseWriter, r *http.Request, state map[string]any) {
	nonce := randomString(16)
	verifier := generateCodeVerifier()
	challenge := generateCodeChallenge(verifier)
	encodedState := (State{
		Code:  randomString(16),
		State: state,
	}).Encode()
	http.SetCookie(w, &http.Cookie{
		Name:  redirectURLKey,
		Value: r.URL.Query().Get("redirect"),
	})
	http.SetCookie(w, &http.Cookie{
		Name:  stateKey,
		Value: encodedState,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  nonceKey,
		Value: nonce,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  codeVerifierKey,
		Value: verifier,
	})

	url := o.cfg.AuthCodeURL(
		encodedState,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.ApprovalForce, oauth2.SetAuthURLParam("prompt", "consent"),
	)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (o *OAuth2) CallbackHandler(w http.ResponseWriter, r *http.Request) (*AuthDetail, error) {
	ctx := r.Context()
	code := r.FormValue("code")
	stateStr := r.FormValue("state")

	stateCookie, err := r.Cookie(stateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state cookie: %w", err)
	}

	if stateCookie.Value != stateStr {
		return nil, fmt.Errorf("state is not valid")
	}

	state, err := DecodeState(stateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state: %w", err)
	}

	codeVerifierCookie, err := r.Cookie(codeVerifierKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get code verifier cookie: %w", err)
	}

	token, err := o.cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifierCookie.Value))
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token field in oauth2 token")
	}

	// TODO: verify the token
	tok, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse id token: %w", err)
	}

	if claims, ok := tok.Claims.(jwt.MapClaims); ok {
		nonceCookie, err := r.Cookie(nonceKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce cookie: %w", err)
		}

		if claims["nonce"] != nonceCookie.Value {
			return nil, fmt.Errorf("nonce is not valid")
		}
	} else {
		return nil, fmt.Errorf("failed to parse claims")
	}

	redirectCookie, err := r.Cookie(redirectURLKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get redirect cookie: %w", err)
	}

	// Delete cookies
	for _, cookie := range []string{stateKey, codeVerifierKey, nonceKey, redirectURLKey} {
		http.SetCookie(w, &http.Cookie{
			Name:   cookie,
			Value:  "",
			MaxAge: -1,
		})
	}

	return &AuthDetail{
		Token:       *token,
		State:       state.State,
		RedirectURL: redirectCookie.Value,
	}, nil
}

type State struct {
	Code  string         `json:"code"`
	State map[string]any `json:"state"`
}

func (s State) Encode() string {
	data, _ := json.Marshal(s)
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeState(encoded string) (State, error) {
	var s State
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(decoded, &s)
	return s, err
}

func generateCodeVerifier() string {
	randomBytes := make([]byte, 32)
	_, _ = rand.Read(randomBytes)
	verifier := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
	return verifier
}

func generateCodeChallenge(verifier string) string {
	sha := sha256.New()
	sha.Write([]byte(verifier))
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sha.Sum(nil))
	return challenge
}

func randomString(len int) string {
	b := make([]byte, len/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
