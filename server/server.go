package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/baseurl"
)

type Server struct {
	config  config.Interface
	r       *chi.Mux
	baseURL string
}

func New(cfg config.Interface) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(baseurl.NewMiddleware(cfg))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // TODO: Fix this
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if cfg.GetBasePath() != "" {
		r.Mount("/"+cfg.GetBasePath(), r)
	}

	return &Server{
		config: cfg,
		r:      r,
	}
}

func (s *Server) Router() *chi.Mux {
	return s.r
}

func (s *Server) Serve() error {
	fmt.Printf("Server started at %s\n", ":"+s.config.GetServerPort())
	if err := http.ListenAndServe(":"+s.config.GetServerPort(), s.r); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
