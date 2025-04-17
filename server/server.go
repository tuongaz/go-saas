package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/pkg/log"
)

type Server struct {
	config      *config.Config
	r           *chi.Mux
	server      *http.Server
	baseURL     string
	middlewares []func(http.Handler) http.Handler
}

func New(cfg *config.Config) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowedMethods:   cfg.CORSAllowedMethods,
		ExposedHeaders:   cfg.CORSExposedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}))

	// Get base URL from config
	baseURL := cfg.BaseURL

	// Create HTTP server with proper timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		config:      cfg,
		r:           r,
		server:      srv,
		baseURL:     baseURL,
		middlewares: []func(http.Handler) http.Handler{},
	}
}

func (s *Server) Router() *chi.Mux {
	return s.r
}

func (s *Server) Serve() error {
	for _, m := range s.middlewares {
		s.r.Use(m)
	}

	log.Info(fmt.Sprintf("Server started at http://127.0.0.1:%s", s.config.ServerPort))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("shutting down server")
	return s.server.Shutdown(ctx)
}

// BaseURL returns the base URL of the server
func (s *Server) BaseURL() string {
	return s.baseURL
}

func (s *Server) AddMiddleware(m ...func(http.Handler) http.Handler) {
	s.middlewares = append(s.middlewares, m...)
}

func (s *Server) PrintRoutes() {
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("%s %s\n", method, route)
		return nil
	}

	if err := chi.Walk(s.Router(), walkFunc); err != nil {
		fmt.Printf("Logging err: %s\n", err.Error())
	}
}
