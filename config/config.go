package config

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	EnvironmentProduction = "production"
)

type Config struct {
	Environment string `mapstructure:"AUTOPUS_ENVIRONMENT"`
	ServerPort  string `mapstructure:"AUTOPUS_SERVER_PORT" validate:"required,port"`
	AppsDir     string `mapstructure:"AUTOPUS_APPS_DIR"`

	JWTSigningSecret        string `mapstructure:"AUTOPUS_JWT_SIGNING_SECRET"`
	JWTTokenLifetimeMinutes int    `mapstructure:"AUTOPUS_JWT_TOKEN_LIFETIME_MINUTES"`
	JWTIssuer               string `mapstructure:"AUTOPUS_JWT_ISSUER"`

	//Auth providers
	AuthGoogleClientID     string `mapstructure:"AUTOPUS_AUTH_GOOGLE_CLIENT_ID"`
	AuthGoogleClientSecret string `mapstructure:"AUTOPUS_AUTH_GOOGLE_CLIENT_SECRET"`

	// Datasource, credentials
	SQLiteDatasource   string `mapstructure:"AUTOPUS_SQLITE_DATASOURCE"`
	OpenAPIKey         string `mapstructure:"AUTOPUS_OPENAI_API_KEY"`
	MySQLDataSource    string `mapstructure:"AUTOPUS_MYSQL_DATASOURCE"`
	PostgresDataSource string `mapstructure:"AUTOPUS_POSTGRES_DATASOURCE"`

	// Security
	EncryptionKey string `mapstructure:"AUTOPUS_ENCRYPTION_KEY"`

	BasePath string
}

func New() (*Config, error) {
	viper.AutomaticEnv()

	validate := validator.New()
	if err := validate.RegisterValidation("port", validatePort); err != nil {
		return nil, fmt.Errorf("unable to register port validator: %w", err)
	}

	// Set default values
	viper.SetDefault("AUTOPUS_ENVIRONMENT", "")
	viper.SetDefault("AUTOPUS_SERVER_PORT", "8080")
	viper.SetDefault("AUTOPUS_APPS_DIR", "./dist/apps")
	viper.SetDefault("AUTOPUS_JWT_SIGNING_SECRET", "default-for-evaluation-change-this-for-production")
	viper.SetDefault("AUTOPUS_JWT_TOKEN_LIFETIME_MINUTES", 15)
	viper.SetDefault("AUTOPUS_JWT_ISSUER", "autopus.ai")

	// Auth providers
	viper.SetDefault("AUTOPUS_AUTH_GOOGLE_CLIENT_ID", "")
	viper.SetDefault("AUTOPUS_AUTH_GOOGLE_CLIENT_SECRET", "")
	viper.SetDefault("AUTOPUS_SQLITE_DATASOURCE", "file:autopusdb?cache=shared&_fk=1")
	viper.SetDefault("AUTOPUS_OPENAI_API_KEY", "")
	viper.SetDefault("AUTOPUS_MYSQL_DATASOURCE", "")
	viper.SetDefault("AUTOPUS_POSTGRES_DATASOURCE", "")
	viper.SetDefault("AUTOPUS_ENCRYPTION_KEY", "must-be-something-else-in-prod")

	// Unmarshal environment variables into Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}
	cfg.BasePath = "api"

	if cfg.IsProduction() {
		if cfg.EncryptionKey == "must-be-something-else-in-prod" {
			return nil, fmt.Errorf("encryption key is required in production environment")
		}
	}

	return &cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Environment == EnvironmentProduction
}

// validatePort is a custom validator for the port
func validatePort(fl validator.FieldLevel) bool {
	port, err := strconv.Atoi(fl.Field().String())
	if err != nil {
		return false
	}
	return port > 0 && port <= 65535
}
