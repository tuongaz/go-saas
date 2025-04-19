package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	EnvironmentProduction = "production"

	defaultEncryptionKey = "must-be-something-else-in-prod"
)

type Config struct {
	Environment     string `mapstructure:"GOS_ENVIRONMENT"`
	BaseURL         string `mapstructure:"GOS_BASE_URL"`
	PublicServerURL string `mapstructure:"GOS_PUBLIC_SERVER_URL"`
	ServerPort      string `mapstructure:"GOS_SERVER_PORT" validate:"required,port"`
	EncryptionKey   string `mapstructure:"GOS_ENCRYPTION_KEY"`

	// Datasource, credentials
	PostgresDataSource string `mapstructure:"GOS_POSTGRES_DATASOURCE"`

	// CORS
	CORSAllowedOrigins   []string `mapstructure:"GOS_CORS_ALLOWED_ORIGINS"`
	CORSAllowedHeaders   []string `mapstructure:"GOS_CORS_ALLOWED_HEADERS"`
	CORSAllowedMethods   []string `mapstructure:"GOS_CORS_ALLOWED_METHODS"`
	CORSExposedHeaders   []string `mapstructure:"GOS_CORS_EXPOSED_HEADERS"`
	CORSAllowCredentials bool     `mapstructure:"GOS_CORS_ALLOW_CREDENTIALS"`
	CORSMaxAge           int      `mapstructure:"GOS_CORS_MAX_AGE"`

	// Auth
	JWTSigningSecret                  string `mapstructure:"GOS_JWT_SIGNING_SECRET"`
	JWTIssuer                         string `mapstructure:"GOS_JWT_ISSUER"`
	JWTTokenLifetimeSeconds           uint   `mapstructure:"GOS_JWT_TOKEN_LIFETIME_SECONDS"`
	Oauth2AuthProviders               map[string]OAuth2ProviderConfig
	ResetPasswordRequestExpiryMinutes uint `mapstructure:"GOS_RESET_PASSWORD_REQUEST_EXPIRY_MINUTES"`

	// Emailer
	ResendAPIKey string `mapstructure:"GOS_RESEND_API_KEY"`
	EmailFrom    string `mapstructure:"GOS_EMAIL_FROM"`
}

func SetDefault(key string, value any) {
	viper.SetDefault(key, value)
}

func New() (*Config, error) {
	viper.AutomaticEnv()

	// Set default values
	SetDefault("GOS_ENVIRONMENT", "")
	SetDefault("GOS_SERVER_PORT", "8080")
	SetDefault("GOS_BASE_URL", "http://localhost:5173")
	SetDefault("GOS_PUBLIC_SERVER_URL", "http://localhost:"+viper.GetString("GOS_SERVER_PORT"))
	SetDefault("GOS_ENCRYPTION_KEY", defaultEncryptionKey)

	SetDefault("GOS_POSTGRES_DATASOURCE", "")
	SetDefault("GOS_EMAIL_FROM", "")

	// CORS
	SetDefault("GOS_CORS_ALLOWED_ORIGINS", []string{"https://*", "http://*"})
	SetDefault("GOS_CORS_ALLOWED_HEADERS", []string{"*"})
	SetDefault("GOS_CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"})
	SetDefault("GOS_CORS_EXPOSED_HEADERS", []string{"Link"})
	SetDefault("GOS_CORS_ALLOW_CREDENTIALS", false)
	SetDefault("GOS_CORS_MAX_AGE", 300)

	SetDefault("GOS_JWT_SIGNING_SECRET", 300)
	SetDefault("GOS_JWT_ISSUER", 300)
	SetDefault("GOS_JWT_TOKEN_LIFETIME_SECONDS", 60*60) // 1 hour
	SetDefault("GOS_RESET_PASSWORD_REQUEST_EXPIRY_MINUTES", 60)

	// Mailer
	SetDefault("GOS_RESEND_API_KEY", "")
	SetDefault("GOS_BREVO_API_KEY", "")

	// Unmarshal environment variables into Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	cfg.Oauth2AuthProviders = make(map[string]OAuth2ProviderConfig)

	if cfg.IsProduction() {
		if cfg.EncryptionKey == defaultEncryptionKey {
			return nil, fmt.Errorf("encryption key is required in production environment")
		}
	}

	return &cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Environment == EnvironmentProduction
}
