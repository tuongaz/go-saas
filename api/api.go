package api

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/autopus/bootstrap/app"
	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/server"
	"github.com/autopus/bootstrap/service/auth"
)

type API struct {
	*privateApp

	cfg     config.Interface
	authSrv *auth.Service
}

type privateApp struct {
	*app.App
}

func New(cfg config.Interface) *API {
	api := &API{
		cfg:        cfg,
		privateApp: &privateApp{app.New(cfg)},
	}

	return api
}

func (a *API) Start() error {
	a.OnBeforeServe().Add(func(ctx context.Context, e *app.OnBeforeServeEvent) error {
		a.authRouterSetup(e.Server)
		return nil
	})

	if err := a.App.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}

func (a *API) authRouterSetup(srv *server.Server) {
	if !a.cfg.IsAuthServiceEnabled() {
		return
	}

	rootRouter := srv.Router()
	authSrv := a.GetAuthService()

	authMiddleware := authSrv.NewMiddleware()

	rootRouter.Route("/auth", func(r chi.Router) {
		// public routes
		r.Post("/signup", authSrv.SignupHandler)
		r.Post("/login", authSrv.LoginHandler)
		r.Post("/token", authSrv.AuthTokenHandler)
		r.Get("/token/authorization", authSrv.TokenAuthorizationHandler)
		r.Get("/{provider}", authSrv.Oauth2AuthenticateHandler)
		r.Get("/{provider}/callback", authSrv.Oauth2LoginSignupCallbackHandler)

		// private routes
		r.With(authMiddleware).Get("/me", authSrv.MeHandler)
	})
}
