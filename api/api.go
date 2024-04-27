package api

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/autopus/bootstrap/app"
	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/server"
	"github.com/autopus/bootstrap/service/auth"
)

type API struct {
	*privateApp

	authSrv *auth.Service
}

type privateApp struct {
	*app.App
}

func New(cfg config.Interface) *API {
	api := &API{
		privateApp: &privateApp{app.New(cfg)},
	}

	return api
}

func (a *API) Start(_ context.Context) error {
	if err := a.App.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	srv := server.New(a.Cfg)
	a.authRouterSetup(srv)

	if err := a.OnBeforeServe().Trigger(
		context.Background(),
		&app.OnBeforeServeEvent{App: a.App, Server: srv},
	); err != nil {
		return fmt.Errorf("failed to trigger on before serve event: %w", err)
	}

	if err := srv.Serve(); err != nil {
		log.Default().Error("failed to start server", log.ErrorAttr(err))
		panic(err)
	}

	return nil
}

func (a *API) authRouterSetup(srv *server.Server) {
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
