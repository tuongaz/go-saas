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
		AllowedOrigins:   cfg.GetCORSAllowedOrigins(),
		AllowedHeaders:   cfg.GetCORSAllowedHeaders(),
		AllowedMethods:   cfg.GetCORSAllowedMethods(),
		ExposedHeaders:   cfg.GetCORSExposedHeaders(),
		AllowCredentials: cfg.GetCORSAllowCredentials(),
		MaxAge:           cfg.GetCORSMaxAge(),
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
