package auth

import (
	"fmt"
	"time"

	"github.com/autopus/bootstrap/pkg/auth/signer"
	"github.com/autopus/bootstrap/pkg/encrypt"
	"github.com/autopus/bootstrap/pkg/hooks"
	"github.com/autopus/bootstrap/service/auth/store"
)

type provider struct {
	clientID     string
	clientSecret string
	scopes       []string
	redirectURL  string
	failureURL   string
	successURL   string
}

type builder struct {
	store       store.Interface
	encryptor   encrypt.Interface
	signer      signer.Interface
	jwtIssuer   string
	jwtLifeTime int
	redirectURL string

	providers map[string]provider
}

func NewBuilder() *builder {
	return &builder{
		providers: make(map[string]provider),
	}
}

func (b *builder) Store(store store.Interface) *builder {
	b.store = store
	return b
}

func (b *builder) Encryptor(encryptor encrypt.Interface) *builder {
	b.encryptor = encryptor
	return b
}

func (b *builder) JWTSecret(secret string) *builder {
	b.signer = signer.NewHS256Signer([]byte(secret))
	return b
}

func (b *builder) JWTIssuer(jwtIssuer string) *builder {
	b.jwtIssuer = jwtIssuer
	return b
}

func (b *builder) JWTLifeTime(lifetimeMinutes int) *builder {
	b.jwtLifeTime = lifetimeMinutes
	return b
}

func (b *builder) RedirectURL(redirectURL string) *builder {
	b.redirectURL = redirectURL
	return b
}

func (b *builder) AddProvider(
	name,
	clientID,
	clientSecret,
	failureURL,
	successURL string,
	scopes []string,
) *builder {
	b.providers[name] = provider{
		clientID:     clientID,
		clientSecret: clientSecret,
		scopes:       scopes,
		redirectURL:  b.redirectURL,
		failureURL:   failureURL,
		successURL:   successURL,
	}
	return b
}

func (b *builder) AddGoogleProvider(clientID, clientSecret, failureURL, successURL string, scopes []string) *builder {
	return b.AddProvider("google", clientID, clientSecret, failureURL, successURL, scopes)
}

func (b *builder) Build() (*Service, error) {
	if b.store == nil {
		return nil, fmt.Errorf("store is required")
	}

	if b.encryptor == nil {
		return nil, fmt.Errorf("encryptor is required")
	}

	if b.signer == nil {
		return nil, fmt.Errorf("signer is required")
	}

	if b.jwtIssuer == "" {
		return nil, fmt.Errorf("jwt issuer is required")
	}

	if b.redirectURL == "" {
		return nil, fmt.Errorf("redirect URL is required")
	}

	if len(b.providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}

	jwtLifeTime := b.jwtLifeTime
	if jwtLifeTime == 0 {
		jwtLifeTime = 15
	}
	return &Service{
		store:                b.store,
		signer:               b.signer,
		encryptor:            b.encryptor,
		jwtIssuer:            b.jwtIssuer,
		tokenLifeTimeMinutes: time.Duration(jwtLifeTime) * time.Minute,
		redirectURL:          b.redirectURL,
		providers:            b.providers,
		onAccountCreated:     &hooks.Hook[*OnAccountCreatedEvent]{},
	}, nil
}
