package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/tuongaz/go-saas/config"
	"github.com/tuongaz/go-saas/core/auth"
	"github.com/tuongaz/go-saas/pkg/encrypt"
	"github.com/tuongaz/go-saas/pkg/hooks"
	"github.com/tuongaz/go-saas/pkg/log"
	"github.com/tuongaz/go-saas/server"
	"github.com/tuongaz/go-saas/service/emailer"
	"github.com/tuongaz/go-saas/store"
)

type Router chi.Router

var _ AppInterface = (*App)(nil)

type AppInterface interface {
	Store() store.Interface

	Auth() auth.Interface

	Emailer() emailer.Interface

	SetEmailer(emailer emailer.Interface)

	Config() *config.Config

	// OnBeforeBootstrap returns the hook that is triggered before the app is bootstrapped.
	OnBeforeBootstrap() *hooks.Hook[*OnBeforeBootstrapEvent]

	// OnAfterBootstrap returns the hook that is triggered after the app is bootstrapped.
	OnAfterBootstrap() *hooks.Hook[*OnAfterBootstrapEvent]

	// OnTerminate returns the hook that is triggered when the app is terminated.
	OnTerminate() *hooks.Hook[*OnTerminateEvent]

	// OnBeforeServe returns the hook that is triggered before the app starts serving.
	OnBeforeServe() *hooks.Hook[*OnBeforeServeEvent]

	// OnDatabaseBootstrap returns the hook that is triggered when the database is bootstrapped.
	OnDatabaseReady() *hooks.Hook[*OnDatabaseReadyEvent]

	// Database event hooks
	OnBeforeRecordCreated() *hooks.Hook[*OnBeforeRecordCreatedEvent]
	OnAfterRecordCreated() *hooks.Hook[*OnAfterRecordCreatedEvent]
	OnBeforeRecordUpdated() *hooks.Hook[*OnBeforeRecordUpdatedEvent]
	OnAfterRecordUpdated() *hooks.Hook[*OnAfterRecordUpdatedEvent]
	OnBeforeRecordDeleted() *hooks.Hook[*OnBeforeRecordDeletedEvent]
	OnAfterRecordDeleted() *hooks.Hook[*OnAfterRecordDeletedEvent]

	// Collection event hooks
	OnBeforeOrganisationCreated() *hooks.Hook[*OnBeforeOrganisationCreatedEvent]
	OnAfterOrganisationCreated() *hooks.Hook[*OnAfterOrganisationCreatedEvent]
	OnBeforeOrganisationUpdated() *hooks.Hook[*OnBeforeOrganisationUpdatedEvent]
	OnAfterOrganisationUpdated() *hooks.Hook[*OnAfterOrganisationUpdatedEvent]
	OnBeforeOrganisationDeleted() *hooks.Hook[*OnBeforeOrganisationDeletedEvent]
	OnAfterOrganisationDeleted() *hooks.Hook[*OnAfterOrganisationDeletedEvent]

	OnBeforeAccountCreated() *hooks.Hook[*OnBeforeAccountCreatedEvent]
	OnAfterAccountCreated() *hooks.Hook[*OnAfterAccountCreatedEvent]
	OnBeforeAccountUpdated() *hooks.Hook[*OnBeforeAccountUpdatedEvent]
	OnAfterAccountUpdated() *hooks.Hook[*OnAfterAccountUpdatedEvent]
	OnBeforeAccountDeleted() *hooks.Hook[*OnBeforeAccountDeletedEvent]
	OnAfterAccountDeleted() *hooks.Hook[*OnAfterAccountDeletedEvent]

	// Start the app
	Start() error

	PublicRoute(pattern string, fn func(r Router))
	PrivateRoute(pattern string, fn func(r Router))
}

type App struct {
	cfg               *config.Config
	onBeforeBootstrap *hooks.Hook[*OnBeforeBootstrapEvent]
	onAfterBootstrap  *hooks.Hook[*OnAfterBootstrapEvent]
	onBeforeServe     *hooks.Hook[*OnBeforeServeEvent]
	onTerminate       *hooks.Hook[*OnTerminateEvent]
	onDatabaseReady   *hooks.Hook[*OnDatabaseReadyEvent]
	store             store.Interface
	auth              auth.Interface
	emailer           emailer.Interface
	server            *server.Server
	encryptor         encrypt.Interface

	// Database event hooks
	onBeforeRecordCreated *hooks.Hook[*OnBeforeRecordCreatedEvent]
	onAfterRecordCreated  *hooks.Hook[*OnAfterRecordCreatedEvent]
	onBeforeRecordUpdated *hooks.Hook[*OnBeforeRecordUpdatedEvent]
	onAfterRecordUpdated  *hooks.Hook[*OnAfterRecordUpdatedEvent]
	onBeforeRecordDeleted *hooks.Hook[*OnBeforeRecordDeletedEvent]
	onAfterRecordDeleted  *hooks.Hook[*OnAfterRecordDeletedEvent]

	onBeforeOrganisationCreated *hooks.Hook[*OnBeforeOrganisationCreatedEvent]
	onAfterOrganisationCreated  *hooks.Hook[*OnAfterOrganisationCreatedEvent]
	onBeforeOrganisationUpdated *hooks.Hook[*OnBeforeOrganisationUpdatedEvent]
	onAfterOrganisationUpdated  *hooks.Hook[*OnAfterOrganisationUpdatedEvent]
	onBeforeOrganisationDeleted *hooks.Hook[*OnBeforeOrganisationDeletedEvent]
	onAfterOrganisationDeleted  *hooks.Hook[*OnAfterOrganisationDeletedEvent]

	onBeforeAccountCreated *hooks.Hook[*OnBeforeAccountCreatedEvent]
	onAfterAccountCreated  *hooks.Hook[*OnAfterAccountCreatedEvent]
	onBeforeAccountUpdated *hooks.Hook[*OnBeforeAccountUpdatedEvent]
	onAfterAccountUpdated  *hooks.Hook[*OnAfterAccountUpdatedEvent]
	onBeforeAccountDeleted *hooks.Hook[*OnBeforeAccountDeletedEvent]
	onAfterAccountDeleted  *hooks.Hook[*OnAfterAccountDeletedEvent]

	serverMiddlewares []func(http.Handler) http.Handler
}

func New(opts ...func(cfg *config.Config)) (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, opt := range opts {
		opt(cfg)
	}

	encryptor := encrypt.New(cfg.EncryptionKey)

	var emailService emailer.Interface
	if cfg.ResendAPIKey != "" {
		emailService = emailer.NewResend(cfg.ResendAPIKey)
	} else {
		emailService = emailer.NewLocal()
	}

	return &App{
		cfg:                         cfg,
		onBeforeBootstrap:           &hooks.Hook[*OnBeforeBootstrapEvent]{},
		onAfterBootstrap:            &hooks.Hook[*OnAfterBootstrapEvent]{},
		onBeforeServe:               &hooks.Hook[*OnBeforeServeEvent]{},
		onTerminate:                 &hooks.Hook[*OnTerminateEvent]{},
		onDatabaseReady:             &hooks.Hook[*OnDatabaseReadyEvent]{},
		onBeforeRecordCreated:       &hooks.Hook[*OnBeforeRecordCreatedEvent]{},
		onAfterRecordCreated:        &hooks.Hook[*OnAfterRecordCreatedEvent]{},
		onBeforeRecordUpdated:       &hooks.Hook[*OnBeforeRecordUpdatedEvent]{},
		onAfterRecordUpdated:        &hooks.Hook[*OnAfterRecordUpdatedEvent]{},
		onBeforeRecordDeleted:       &hooks.Hook[*OnBeforeRecordDeletedEvent]{},
		onAfterRecordDeleted:        &hooks.Hook[*OnAfterRecordDeletedEvent]{},
		onBeforeOrganisationCreated: &hooks.Hook[*OnBeforeOrganisationCreatedEvent]{},
		onAfterOrganisationCreated:  &hooks.Hook[*OnAfterOrganisationCreatedEvent]{},
		onBeforeOrganisationUpdated: &hooks.Hook[*OnBeforeOrganisationUpdatedEvent]{},
		onAfterOrganisationUpdated:  &hooks.Hook[*OnAfterOrganisationUpdatedEvent]{},
		onBeforeOrganisationDeleted: &hooks.Hook[*OnBeforeOrganisationDeletedEvent]{},
		onAfterOrganisationDeleted:  &hooks.Hook[*OnAfterOrganisationDeletedEvent]{},
		onBeforeAccountCreated:      &hooks.Hook[*OnBeforeAccountCreatedEvent]{},
		onAfterAccountCreated:       &hooks.Hook[*OnAfterAccountCreatedEvent]{},
		onBeforeAccountUpdated:      &hooks.Hook[*OnBeforeAccountUpdatedEvent]{},
		onAfterAccountUpdated:       &hooks.Hook[*OnAfterAccountUpdatedEvent]{},
		onBeforeAccountDeleted:      &hooks.Hook[*OnBeforeAccountDeletedEvent]{},
		onAfterAccountDeleted:       &hooks.Hook[*OnAfterAccountDeletedEvent]{},
		encryptor:                   encryptor,
		emailer:                     emailService,
		serverMiddlewares:           []func(http.Handler) http.Handler{},
	}, nil
}

func (a *App) Store() store.Interface {
	return a.store
}

func (a *App) Auth() auth.Interface {
	return a.auth
}

func (a *App) Emailer() emailer.Interface {
	return a.emailer
}

func (a *App) SetEmailer(emailer emailer.Interface) {
	a.emailer = emailer
}

func (a *App) Config() *config.Config {
	return a.cfg
}

func (a *App) Encryptor() encrypt.Interface {
	return a.encryptor
}

func (a *App) PublicRoute(pattern string, fn func(r Router)) {
	a.server.Router().Route(pattern, func(r chi.Router) {
		fn(r)
	})
}

func (a *App) AddServerMiddleware(m ...func(http.Handler) http.Handler) {
	a.serverMiddlewares = append(a.serverMiddlewares, m...)
}

func (a *App) PrivateRoute(pattern string, fn func(r Router)) {
	a.server.Router().Route(pattern, func(r chi.Router) {
		r.Use(a.auth.NewMiddleware())
		fn(r)
	})
}

func (a *App) Start() error {
	ctx := context.Background()

	if err := a.bootstrap(ctx); err != nil {
		return fmt.Errorf("app bootstrap: %w", err)
	}

	a.server = server.New(a.Config(), a.serverMiddlewares...)

	// Setup the auth API using the existing auth service
	a.auth.SetupAPI(a.server.Router())

	// Validate everything before starting the service
	if err := a.validate(); err != nil {
		return fmt.Errorf("validate app: %w", err)
	}

	if err := a.OnBeforeServe().Trigger(
		ctx,
		&OnBeforeServeEvent{
			App: a,
		},
	); err != nil {
		return fmt.Errorf("trigger hooks on before serve event: %w", err)
	}

	a.server.PrintRoutes()

	done := make(chan bool, 1)
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	go func() {
		if err := a.server.Serve(); err != nil {
			log.Default().Error("failed to serve", "err", err)
			// Signal that we need to terminate due to server error
			done <- true
		}

		// Signal that server has stopped (this happens after Shutdown is called)
		done <- true
	}()

	<-done

	if err := a.OnTerminate().Trigger(
		ctx,
		&OnTerminateEvent{App: a},
		func(ctx context.Context, e *OnTerminateEvent) error {
			return e.App.Shutdown()
		},
	); err != nil {
		return fmt.Errorf("trigger hooks on terminate event: %w", err)
	}

	return nil
}

func (a *App) Shutdown() error {
	log.Default().Warn("shutting down app")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
	}

	return nil
}

func (a *App) bootstrap(ctx context.Context) error {
	if err := a.OnBeforeBootstrap().Trigger(ctx, &OnBeforeBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("before bootstrap: %w", err)
	}

	log.Info("bootstrapping database")
	st, err := store.New(a.Config().PostgresDataSource)
	if err != nil {
		return fmt.Errorf("new store: %w", err)
	}
	a.store = st

	// Register database event handler
	st.AddEventHandler(NewAppDatabaseEventHandler(a))

	if err := a.OnDatabaseReady().Trigger(ctx, &OnDatabaseReadyEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("database bootstrap: %w", err)
	}

	// Initialize auth service
	authSrv, err := auth.New(a.Config(), a.Emailer(), a.store)
	if err != nil {
		return fmt.Errorf("new auth service: %w", err)
	}
	a.auth = authSrv

	a.OnTerminate().Add(func(ctx context.Context, e *OnTerminateEvent) error {
		a.store.Close()

		return nil
	})

	if err := a.OnAfterBootstrap().Trigger(ctx, &OnAfterBootstrapEvent{
		App: a,
	}); err != nil {
		return fmt.Errorf("after bootstrap: %w", err)
	}

	return nil
}

func (a *App) validate() error {
	if a.emailer == nil {
		return fmt.Errorf("email service is not set")
	}

	return nil
}

func (a *App) OnBeforeRecordCreated() *hooks.Hook[*OnBeforeRecordCreatedEvent] {
	return a.onBeforeRecordCreated
}

func (a *App) OnAfterRecordCreated() *hooks.Hook[*OnAfterRecordCreatedEvent] {
	return a.onAfterRecordCreated
}

func (a *App) OnBeforeRecordUpdated() *hooks.Hook[*OnBeforeRecordUpdatedEvent] {
	return a.onBeforeRecordUpdated
}

func (a *App) OnAfterRecordUpdated() *hooks.Hook[*OnAfterRecordUpdatedEvent] {
	return a.onAfterRecordUpdated
}

func (a *App) OnBeforeRecordDeleted() *hooks.Hook[*OnBeforeRecordDeletedEvent] {
	return a.onBeforeRecordDeleted
}

func (a *App) OnAfterRecordDeleted() *hooks.Hook[*OnAfterRecordDeletedEvent] {
	return a.onAfterRecordDeleted
}

func (a *App) OnAfterOrganisationCreated() *hooks.Hook[*OnAfterOrganisationCreatedEvent] {
	return a.onAfterOrganisationCreated
}

func (a *App) OnBeforeAccountCreated() *hooks.Hook[*OnBeforeAccountCreatedEvent] {
	return a.onBeforeAccountCreated
}

func (a *App) OnAfterAccountCreated() *hooks.Hook[*OnAfterAccountCreatedEvent] {
	return a.onAfterAccountCreated
}

func (a *App) OnBeforeOrganisationCreated() *hooks.Hook[*OnBeforeOrganisationCreatedEvent] {
	return a.onBeforeOrganisationCreated
}

func (a *App) OnBeforeOrganisationUpdated() *hooks.Hook[*OnBeforeOrganisationUpdatedEvent] {
	return a.onBeforeOrganisationUpdated
}

func (a *App) OnAfterOrganisationUpdated() *hooks.Hook[*OnAfterOrganisationUpdatedEvent] {
	return a.onAfterOrganisationUpdated
}

func (a *App) OnBeforeOrganisationDeleted() *hooks.Hook[*OnBeforeOrganisationDeletedEvent] {
	return a.onBeforeOrganisationDeleted
}

func (a *App) OnAfterOrganisationDeleted() *hooks.Hook[*OnAfterOrganisationDeletedEvent] {
	return a.onAfterOrganisationDeleted
}

func (a *App) OnBeforeAccountUpdated() *hooks.Hook[*OnBeforeAccountUpdatedEvent] {
	return a.onBeforeAccountUpdated
}

func (a *App) OnAfterAccountUpdated() *hooks.Hook[*OnAfterAccountUpdatedEvent] {
	return a.onAfterAccountUpdated
}

func (a *App) OnBeforeAccountDeleted() *hooks.Hook[*OnBeforeAccountDeletedEvent] {
	return a.onBeforeAccountDeleted
}

func (a *App) OnAfterAccountDeleted() *hooks.Hook[*OnAfterAccountDeletedEvent] {
	return a.onAfterAccountDeleted
}
