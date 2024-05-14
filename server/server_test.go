package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tuongaz/go-saas/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		ServerPort: "8080",
	}

	server := New(cfg)
	assert.NotNil(t, server)
}

func TestServer_Router(t *testing.T) {
	cfg := &config.Config{
		ServerPort: "8080",
	}

	server := New(cfg)
	assert.NotNil(t, server.Router())
}
