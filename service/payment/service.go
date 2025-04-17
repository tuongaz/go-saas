package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v78"

	"github.com/tuongaz/go-saas/core"
	"github.com/tuongaz/go-saas/service/payment/model"
	"github.com/tuongaz/go-saas/service/payment/store"
	coreStore "github.com/tuongaz/go-saas/store"
)

var _ ServiceInterface = &Service{}

const (
	providerStripe = "stripe"
)

type ServiceInterface interface {
	CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*model.Invoice, error)
	GetPaymentMethods(ctx context.Context, accountID string) ([]*model.PaymentMethod, error)
	GetDefaultPaymentMethod(ctx context.Context, accountID string) (*model.PaymentMethod, error)
	Charge(ctx context.Context, invoiceID string) (*model.Payment, error)
}

type CreateInvoiceInput struct {
	AccountID     string
	ReferenceID   string
	AmountInCents int64
	Currency      string
	AutoCharge    bool
}

type Provider interface {
	Charge(ctx context.Context, paymentMethod *model.PaymentMethod, input model.ChargeInput) (any, error)
}

type Service struct {
	app                 core.AppInterface
	store               store.Interface
	registeredProviders map[string]bool
}

func MustRegister(appInstance core.AppInterface) *Service {
	cfg, err := newConfig()
	if err != nil {
		panic(fmt.Errorf("new payment config: %w", err))
	}

	s := &Service{
		app:                 appInstance,
		registeredProviders: map[string]bool{},
	}

	if cfg.StripePrivateKey != "" {
		stripe.Key = cfg.StripePrivateKey
		s.registeredProviders[providerStripe] = true
	}

	appInstance.OnAfterBootstrap().Add(func(ctx context.Context, e *core.OnAfterBootstrapEvent) error {
		st, err := store.New(appInstance.Store())
		if err != nil {
			return fmt.Errorf("create payment store: %w", err)
		}
		s.store = st

		return nil
	})

	appInstance.OnBeforeServe().Add(func(ctx context.Context, e *core.OnBeforeServeEvent) error {
		e.App.PrivateRoute("/payment-methods/stripe", func(r core.Router) {
			r.Post("/", s.CreateStripePaymentMethodHandler)
		})

		return nil
	})

	return s
}

func (s *Service) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*model.Invoice, error) {
	invoice, err := s.store.CreateInvoice(ctx, model.CreateInvoiceInput{
		AccountID:     input.AccountID,
		ReferenceID:   input.ReferenceID,
		AmountInCents: input.AmountInCents,
		Currency:      input.Currency,
		Status:        model.InvoiceStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	if input.AutoCharge {
		if _, err := s.Charge(ctx, invoice.ID); err != nil {
			return nil, fmt.Errorf("charge invoice: %w", err)
		}
	}

	return invoice, nil
}

func (s *Service) GetPaymentMethods(ctx context.Context, accountID string) ([]*model.PaymentMethod, error) {
	paymentMethods, err := s.store.GetPaymentMethods(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get payment methods: %w", err)
	}

	return paymentMethods, nil
}

func (s *Service) GetDefaultPaymentMethod(ctx context.Context, accountID string) (*model.PaymentMethod, error) {
	paymentMethods, err := s.GetPaymentMethods(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get payment methods: %w", err)
	}

	if len(paymentMethods) == 0 {
		return nil, coreStore.NewNotFoundErr(fmt.Errorf("payment methods not found"))
	}

	for _, paymentMethod := range paymentMethods {
		if paymentMethod.IsDefault {
			return paymentMethod, nil
		}
	}

	// There's no default, so we'll pick the first one
	return paymentMethods[0], nil
}

func (s *Service) providerRegistered(provider string) bool {
	_, ok := s.registeredProviders[provider]
	return ok
}
