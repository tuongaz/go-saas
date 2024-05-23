package model

type CreateInvoiceInput struct {
	AccountID     string
	ReferenceID   string
	AmountInCents int64
	Currency      string
	Status        string
}

type UpdateInvoiceInput struct {
	AccountID     *string
	AmountInCents *int64
	Currency      *string
	Status        *string
}

type CreatePaymentInput struct {
	InvoiceID       string
	PaymentMethodID string
	AmountInCents   int64
	Currency        string
	Status          string
}

type UpdatePaymentInput struct {
	ChargeData *string
	Status     *string
}

type ChargeInput struct {
	AmountInCents int64
	Currency      string
}
