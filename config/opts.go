package config

type OAuth2ProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	Scopes       []string
	RedirectURL  string
	FailureURL   string
	SuccessURL   string
}

func WithOauth2Provider(name, clientID, clientSecret, redirectURL, failureURL, successURL string, scopes []string) func(*Config) {
	provider := OAuth2ProviderConfig{
		Name:         name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		RedirectURL:  redirectURL,
		FailureURL:   failureURL,
		SuccessURL:   successURL,
	}

	return func(cfg *Config) {
		cfg.Oauth2AuthProviders[name] = provider
	}
}
