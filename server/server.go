package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/tuongaz/go-saas/config"
)

type Server struct {
	config  *config.Config
	r       *chi.Mux
	baseURL string
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

	return &Server{
		config: cfg,
		r:      r,
	}
}

func (s *Server) Router() *chi.Mux {
	return s.r
}

func (s *Server) Serve() error {
	fmt.Printf("Server started at %s\n", ":"+s.config.ServerPort)
	if err := http.ListenAndServe(":"+s.config.ServerPort, s.r); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
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
