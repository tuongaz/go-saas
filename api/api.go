package api

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"

	app2 "github.com/autopus/bootstrap/app"
	"github.com/autopus/bootstrap/config"
	"github.com/autopus/bootstrap/pkg/auth/signer"
	"github.com/autopus/bootstrap/pkg/baseurl"
	"github.com/autopus/bootstrap/pkg/encrypt"
	"github.com/autopus/bootstrap/pkg/log"
	"github.com/autopus/bootstrap/server"
	"github.com/autopus/bootstrap/service/auth"
	"github.com/autopus/bootstrap/ui"
)

type API struct {
	*app2.App

	authSrv *auth.Service
}

func New(cfg *config.Config) *API {
	api := &API{
		App: app2.New(cfg),
	}

	return api
}

func (a *API) Start(ctx context.Context) error {
	if err := a.App.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	encryptor := encrypt.New(a.Cfg.EncryptionKey)
	// setup authentication service
	authSrv, err := auth.New(
		a.Cfg,
		a.Store,
		encryptor,
		signer.NewHS256Signer([]byte(a.Cfg.JWTSigningSecret)),
		auth.WithJWTLifeTimeMinutes(a.Cfg.JWTTokenLifetimeMinutes),
	)
	if err != nil {
		log.Default().Error("failed to init auth service", log.ErrorAttr(err))
		panic(err)
	}

	srv := server.New(a.Cfg)
	routerSetup(
		a.Cfg,
		srv.Router(),
		authSrv,
	)

	if err := srv.Serve(); err != nil {
		log.Default().Error("failed to start server", log.ErrorAttr(err))
		panic(err)
	}

	return nil
}

func routerSetup(
	cfg *config.Config,
	rootRouter chi.Router,
	authSrv *auth.Service,
) {
	authMiddleware := authSrv.NewMiddleware()
	baseURLMw := baseurl.NewMiddleware(cfg)
	rootRouter.Use(baseURLMw)

	rootRouter.Route("/"+cfg.BasePath, func(apiRouter chi.Router) {
		apiRouter.Route("/auth", func(r chi.Router) {
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
	})

	if err := ui.Handler(rootRouter); err != nil {
		log.Default().Error("failed to init frontend", log.ErrorAttr(err))
		panic(err)
	}
}
