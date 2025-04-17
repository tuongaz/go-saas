package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tuongaz/go-saas/pkg/uid"
	"github.com/tuongaz/go-saas/service/payment/model"
	"github.com/tuongaz/go-saas/store"
	"github.com/tuongaz/go-saas/store/types"
)

const (
	tablePaymentMethod  = "payment_method"
	tableInvoice        = "invoice"
	tablePayment        = "payment"
	tableStripeCustomer = "stripe_customer"
)

//go:embed postgres.sql
var postgresSchema string

var _ Interface = (*Store)(nil)

type Interface interface {
	CreateInvoice(ctx context.Context, input model.CreateInvoiceInput) (*model.Invoice, error)
	UpdateInvoice(ctx context.Context, id string, input model.UpdateInvoiceInput) error
	GetInvoice(ctx context.Context, id string) (*model.Invoice, error)

	CreatePayment(ctx context.Context, input model.CreatePaymentInput) (*model.Payment, error)
	CreatePaymentMethod(ctx context.Context, input CreatePaymentMethodInput) (*model.PaymentMethod, error)
	GetPaymentMethods(ctx context.Context, accountID string) ([]*model.PaymentMethod, error)
	UpdatePayment(ctx context.Context, id string, input model.UpdatePaymentInput) error

	CreateStripeCustomer(ctx context.Context, accountID, customerID string) (*model.StripeCustomer, error)
	GetStripeCustomer(ctx context.Context, accountID string) (*model.StripeCustomer, error)
}

type Store struct {
	store store.Interface
}

func (s *Store) CreateStripeCustomer(ctx context.Context, accountID, customerID string) (*model.StripeCustomer, error) {
	record := types.Record{
		"account_id":  accountID,
		"customer_id": customerID,
	}

	if _, err := s.store.Collection(tableStripeCustomer).CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("create stripe customer record: %w", err)
	}

	var stripeCustomer model.StripeCustomer
	if err := record.Decode(&stripeCustomer); err != nil {
		return nil, fmt.Errorf("decode to stripe customer: %w", err)
	}

	return &stripeCustomer, nil
}

func (s *Store) GetStripeCustomer(ctx context.Context, accountID string) (*model.StripeCustomer, error) {
	record, err := s.store.Collection(tableStripeCustomer).FindOne(ctx, store.Filter{
		"account_id": accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("find stripe customer record: %w", err)
	}

	var stripeCustomer model.StripeCustomer
	if err := record.Decode(&stripeCustomer); err != nil {
		return nil, fmt.Errorf("decode to stripe customer: %w", err)
	}

	return &stripeCustomer, nil
}

func New(store store.Interface) (*Store, error) {
	if err := store.Exec(context.Background(), postgresSchema); err != nil {
		return nil, fmt.Errorf("failed to create payment schema: %w", err)
	}
	return &Store{
		store: store,
	}, nil
}

func (s *Store) CreateInvoice(ctx context.Context, input model.CreateInvoiceInput) (*model.Invoice, error) {
	record := types.Record{
		"id":              uid.ID(),
		"account_id":      input.AccountID,
		"reference_id":    input.ReferenceID,
		"amount_in_cents": input.AmountInCents,
		"currency":        input.Currency,
		"status":          input.Status,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}

	if _, err := s.store.Collection(tableInvoice).CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("create invoice record: %w", err)
	}

	var invoice model.Invoice
	if err := record.Decode(&invoice); err != nil {
		return nil, fmt.Errorf("decode to invoice: %w", err)
	}

	return &invoice, nil
}

func (s *Store) GetInvoice(ctx context.Context, id string) (*model.Invoice, error) {
	record, err := s.store.Collection(tableInvoice).FindOne(ctx, store.Filter{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("find invoice record: %w", err)
	}

	var invoice model.Invoice
	if err := record.Decode(&invoice); err != nil {
		return nil, fmt.Errorf("decode to invoice: %w", err)
	}

	return &invoice, nil
}

func (s *Store) UpdateInvoice(ctx context.Context, id string, input model.UpdateInvoiceInput) error {
	_, err := s.store.Collection(tableInvoice).FindOne(ctx, store.Filter{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("find invoice record: %w", err)
	}

	var record = types.Record{}

	if input.AccountID != nil {
		record["account_id"] = *input.AccountID
	}

	if input.AmountInCents != nil {
		record["amount_in_cents"] = *input.AmountInCents
	}

	if input.Currency != nil {
		record["currency"] = *input.Currency
	}

	if input.Status != nil {
		record["status"] = *input.Status
	}

	record["updated_at"] = time.Now()

	if _, err := s.store.Collection(tableInvoice).UpdateRecord(ctx, id, record); err != nil {
		return fmt.Errorf("update invoice record: %w", err)
	}

	return nil
}

func (s *Store) CreatePayment(ctx context.Context, input model.CreatePaymentInput) (*model.Payment, error) {
	record := types.Record{
		"id":                uid.ID(),
		"invoice_id":        input.InvoiceID,
		"payment_method_id": input.PaymentMethodID,
		"amount_in_cents":   input.AmountInCents,
		"currency":          input.Currency,
		"status":            input.Status,
		"created_at":        time.Now(),
		"updated_at":        time.Now(),
	}

	if _, err := s.store.Collection(tablePayment).CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("create payment record: %w", err)
	}

	var payment model.Payment
	if err := record.Decode(&payment); err != nil {
		return nil, fmt.Errorf("decode to payment: %w", err)
	}

	return &payment, nil
}

func (s *Store) UpdatePayment(ctx context.Context, id string, input model.UpdatePaymentInput) error {
	_, err := s.store.Collection(tablePayment).FindOne(ctx, store.Filter{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("find payment record: %w", err)
	}

	var record = types.Record{}

	if input.Status != nil {
		record["status"] = *input.Status
	}

	if input.ChargeData != nil {
		record["charge_data"] = *input.ChargeData
	}

	record["updated_at"] = time.Now()

	if _, err := s.store.Collection(tablePayment).UpdateRecord(ctx, id, record); err != nil {
		return fmt.Errorf("update payment record: %w", err)
	}

	return nil
}

type CreatePaymentMethodInput struct {
	AccountID          string
	Provider           string
	ProviderCustomerID string
	Data               map[string]any
}

func (s *Store) CreatePaymentMethod(ctx context.Context, input CreatePaymentMethodInput) (*model.PaymentMethod, error) {
	data, err := json.Marshal(input.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	record := types.Record{
		"id":                   uid.ID(),
		"account_id":           input.AccountID,
		"provider":             input.Provider,
		"provider_customer_id": input.ProviderCustomerID,
		"data":                 string(data),
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
	}

	if _, err := s.store.Collection(tablePaymentMethod).CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("create payment method record: %w", err)
	}

	var paymentMethod model.PaymentMethod
	if err := record.Decode(&paymentMethod); err != nil {
		return nil, fmt.Errorf("decode to payment method: %w", err)
	}

	return &paymentMethod, nil
}

func (s *Store) GetPaymentMethods(ctx context.Context, accountID string) ([]*model.PaymentMethod, error) {
	records, err := s.store.Collection(tablePaymentMethod).Find(ctx, store.WithFilter(store.Filter{
		"account_id": accountID,
	}))
	if err != nil {
		return nil, fmt.Errorf("find payment methods: %w", err)
	}

	var paymentMethods []*model.PaymentMethod
	if err := records.Decode(&paymentMethods); err != nil {
		return nil, fmt.Errorf("decode payment methods: %w", err)
	}

	return paymentMethods, nil
}
