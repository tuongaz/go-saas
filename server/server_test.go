package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tuongaz/go-saas/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		ServerPort: "8080",
		BaseURL:    "http://localhost:8080",
	}

	server := New(cfg)
	assert.NotNil(t, server)
	assert.Equal(t, "http://localhost:8080", server.BaseURL())

	// Check that the HTTP server is properly configured
	assert.Equal(t, ":8080", server.server.Addr)
	assert.Equal(t, 120*time.Second, server.server.ReadTimeout)
	assert.Equal(t, 120*time.Second, server.server.WriteTimeout)
	assert.Equal(t, 120*time.Second, server.server.IdleTimeout)
}

func TestServer_Router(t *testing.T) {
	cfg := &config.Config{
		ServerPort: "8080",
		BaseURL:    "http://localhost:8080",
	}

	server := New(cfg)
	assert.NotNil(t, server.Router())
}

func TestServer_Shutdown(t *testing.T) {
	cfg := &config.Config{
		ServerPort: "8888", // Use a different port to avoid conflicts
		BaseURL:    "http://localhost:8888",
	}

	server := New(cfg)

	// Start the server in a goroutine
	go func() {
		err := server.Serve()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("server.Serve() error = %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err, "server.Shutdown() should not error")
}
