package config

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"github.com/spf13/viper"

	"github.com/autopus/bootstrap"
)

const (
	EnvironmentProduction = "production"

	AuthService      = "auth"
	SchedulerService = "scheduler"
)

type Interface interface {
	GetEnvironment() string
	IsProduction() bool

	GetServerPort() string
	GetBasePath() string

	GetEncryptionKey() string
	GetJWTSigningSecret() string
	GetJWTTokenLifetimeMinutes() int
	GetJWTIssuer() string

	GetSQLiteDatasource() string
	GetSQLiteSchema() string
	GetMySQLDataSource() string
	GetMySQLSchema() string
	GetPostgresDataSource() string
	GetPostgresSchema() string

	GetAuthGoogleClientID() string
	GetAuthGoogleClientSecret() string

	GetCORSAllowedOrigins() []string
	GetCORSAllowedHeaders() []string
	GetCORSAllowedMethods() []string
	GetCORSExposedHeaders() []string
	GetCORSAllowCredentials() bool
	GetCORSMaxAge() int

	IsAuthServiceEnabled() bool
	IsSchedulerServiceEnabled() bool
}

type Config struct {
	Environment   string `mapstructure:"AUTOPUS_ENVIRONMENT"`
	ServerPort    string `mapstructure:"AUTOPUS_SERVER_PORT" validate:"required,port"`
	EncryptionKey string `mapstructure:"AUTOPUS_ENCRYPTION_KEY"`
	BasePath      string

	// Authentication
	JWTSigningSecret        string `mapstructure:"AUTOPUS_JWT_SIGNING_SECRET"`
	JWTTokenLifetimeMinutes int    `mapstructure:"AUTOPUS_JWT_TOKEN_LIFETIME_MINUTES"`
	JWTIssuer               string `mapstructure:"AUTOPUS_JWT_ISSUER"`

	// Datasource, credentials
	SqliteDatasource   string `mapstructure:"AUTOPUS_SQLITE_DATASOURCE"`
	MySqlDataSource    string `mapstructure:"AUTOPUS_MYSQL_DATASOURCE"`
	PostgresDataSource string `mapstructure:"AUTOPUS_POSTGRES_DATASOURCE"`

	//Auth providers
	AuthGoogleClientID     string `mapstructure:"AUTOPUS_AUTH_GOOGLE_CLIENT_ID"`
	AuthGoogleClientSecret string `mapstructure:"AUTOPUS_AUTH_GOOGLE_CLIENT_SECRET"`

	AppsDir    string `mapstructure:"AUTOPUS_APPS_DIR"`
	OpenAPIKey string `mapstructure:"AUTOPUS_OPENAI_API_KEY"`

	// CORS
	CORSAllowedOrigins   []string `mapstructure:"AUTOPUS_CORS_ALLOWED_ORIGINS"`
	CORSAllowedHeaders   []string `mapstructure:"AUTOPUS_CORS_ALLOWED_HEADERS"`
	CORSAllowedMethods   []string `mapstructure:"AUTOPUS_CORS_ALLOWED_METHODS"`
	CORSExposedHeaders   []string `mapstructure:"AUTOPUS_CORS_EXPOSED_HEADERS"`
	CORSAllowCredentials bool     `mapstructure:"AUTOPUS_CORS_ALLOW_CREDENTIALS"`
	CORSMaxAge           int      `mapstructure:"AUTOPUS_CORS_MAX_AGE"`

	EnabledServices []string `mapstructure:"AUTOPUS_ENABLED_SERVICES"`

	sqliteSchema   string
	mysqlSchema    string
	postgresSchema string
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

	// CORS
	viper.SetDefault("AUTOPUS_CORS_ALLOWED_ORIGINS", []string{"https://*", "http://*"})
	viper.SetDefault("AUTOPUS_CORS_ALLOWED_HEADERS", []string{"*"})
	viper.SetDefault("AUTOPUS_CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("AUTOPUS_CORS_EXPOSED_HEADERS", []string{"Link"})
	viper.SetDefault("AUTOPUS_CORS_ALLOW_CREDENTIALS", false)
	viper.SetDefault("AUTOPUS_CORS_MAX_AGE", 300)

	// Enabled services
	viper.SetDefault("AUTOPUS_ENABLED_SERVICES", []string{AuthService, SchedulerService})

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

func (c *Config) GetEnvironment() string {
	return c.Environment
}

func (c *Config) GetServerPort() string {
	return c.ServerPort
}

func (c *Config) GetEncryptionKey() string {
	return c.EncryptionKey
}

func (c *Config) GetBasePath() string {
	return c.BasePath
}

func (c *Config) GetJWTSigningSecret() string {
	return c.JWTSigningSecret
}

func (c *Config) GetJWTTokenLifetimeMinutes() int {
	return c.JWTTokenLifetimeMinutes
}

func (c *Config) GetJWTIssuer() string {
	return c.JWTIssuer
}

func (c *Config) GetSQLiteDatasource() string {
	return c.SqliteDatasource
}

func (c *Config) SetSQLiteSchema(schema string) {
	c.sqliteSchema = schema
}

func (c *Config) GetSQLiteSchema() string {
	if c.sqliteSchema == "" {
		return app.SqliteSchema
	}

	return c.sqliteSchema
}

func (c *Config) GetMySQLDataSource() string {
	return c.MySqlDataSource
}

func (c *Config) SetMySQLSchema(schema string) {
	c.mysqlSchema = schema
}

func (c *Config) GetMySQLSchema() string {
	if c.mysqlSchema == "" {
		return app.MySQLSchema
	}

	return c.mysqlSchema
}

func (c *Config) GetPostgresDataSource() string {
	return c.PostgresDataSource
}

func (c *Config) SetPostgresSchema(schema string) {
	c.postgresSchema = schema
}

func (c *Config) GetPostgresSchema() string {
	if c.postgresSchema == "" {
		return app.PostgresSchema
	}

	return c.postgresSchema
}

func (c *Config) GetAuthGoogleClientID() string {
	return c.AuthGoogleClientID
}

func (c *Config) GetAuthGoogleClientSecret() string {
	return c.AuthGoogleClientSecret
}

func (c *Config) IsProduction() bool {
	return c.Environment == EnvironmentProduction
}

func (c *Config) GetCORSAllowedOrigins() []string {
	return c.CORSAllowedOrigins
}

func (c *Config) GetCORSAllowedHeaders() []string {
	return c.CORSAllowedHeaders
}

func (c *Config) GetCORSAllowedMethods() []string {
	return c.CORSAllowedMethods
}

func (c *Config) GetCORSExposedHeaders() []string {
	return c.CORSExposedHeaders
}

func (c *Config) GetCORSAllowCredentials() bool {
	return c.CORSAllowCredentials
}

func (c *Config) GetCORSMaxAge() int {
	return c.CORSMaxAge
}

func (c *Config) IsAuthServiceEnabled() bool {
	return lo.Contains(c.EnabledServices, AuthService)
}

func (c *Config) IsSchedulerServiceEnabled() bool {
	return lo.Contains(c.EnabledServices, SchedulerService)
}

// validatePort is a custom validator for the port
func validatePort(fl validator.FieldLevel) bool {
	port, err := strconv.Atoi(fl.Field().String())
	if err != nil {
		return false
	}
	return port > 0 && port <= 65535
}
