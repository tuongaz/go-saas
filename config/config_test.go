package config

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigDefaults(t *testing.T) {
	// Clear environment variables to test default values
	os.Clearenv()

	cfg, err := New()
	assert.NoError(t, err)
	assert.Equal(t, "8080", cfg.GetServerPort)
}

func TestNewConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("AUTOPUS_SERVER_PORT", "3000")
	os.Setenv("AUTOPUS_BASE_URL", "http://example.com")
	os.Setenv("AUTOPUS_FRONTEND_BASE_URL", "http://example.com")

	cfg, err := New()
	assert.NoError(t, err)
	assert.Equal(t, "3000", cfg.GetServerPort)

	// Clear environment variables after test
	os.Clearenv()
}

func TestConfigValidation(t *testing.T) {
	validate := getValidatorWithPortValidation()

	// Valid configuration
	cfg := &Config{
		ServerPort: "8080",
	}
	assert.NoError(t, validate.Struct(cfg))

	// Invalid Port
	cfg.GetServerPort = "70000" // Port number out of range
	assert.Error(t, validate.Struct(cfg))
}

func TestCustomPortValidator(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("port", validatePort)

	// Valid Port
	assert.NoError(t, validate.Var("8080", "port"))

	// Invalid Port
	assert.Error(t, validate.Var("70000", "port")) // Port number out of range
}

func getValidatorWithPortValidation() *validator.Validate {
	validate := validator.New()
	_ = validate.RegisterValidation("port", validatePort)
	return validate
}
