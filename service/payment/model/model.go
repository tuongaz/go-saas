package model

import (
	"time"
)

const (
	InvoiceStatusPending = "pending"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusFailed  = "failed"

	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"
)

type Invoice struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"account_id"`
	AmountInCents int64     `json:"amount_in_cents"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PaymentMethod struct {
	ID                 string    `json:"id"`
	AccountID          string    `json:"account_id"`
	ProviderCustomerID string    `json:"provider_customer_id"`
	Provider           string    `json:"provider"`
	Data               string    `json:"data"`
	IsDefault          bool      `json:"is_default"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Payment struct {
	ID              string    `json:"id"`
	InvoiceID       string    `json:"invoice_id"`
	PaymentMethodID string    `json:"payment_method_id"`
	AmountInCents   int64     `json:"amount_in_cents"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	ProviderData    string    `json:"provider_data"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StripeCustomer struct {
	AccountID  string `json:"account_id"`
	CustomerID string `json:"customer_id"`
}
