package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/paymentmethod"

	"github.com/tuongaz/go-saas/core/auth"
	"github.com/tuongaz/go-saas/pkg/httputil"
	"github.com/tuongaz/go-saas/service/payment/model"
	"github.com/tuongaz/go-saas/service/payment/store"
	coreStore "github.com/tuongaz/go-saas/store"
)

type CreateStripePaymentMethodInput struct {
	PaymentMethodID string `json:"payment_method_id"`
}

func (s *Service) CreateStripePaymentMethodHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input, err := httputil.ParseRequestBody[CreateStripePaymentMethodInput](r)
	if err != nil {
		httputil.HandleResponse(ctx, w, r, err)
		return
	}

	account, err := s.app.Auth().GetAccount(ctx, auth.AccountID(ctx))
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	stripeCustomerID := ""
	stripeCustomer, err := s.store.GetStripeCustomer(ctx, account.ID)
	if err != nil {
		if coreStore.IsNotFoundError(err) {
			cust, stripeErr := customer.New(&stripe.CustomerParams{
				Email: stripe.String(account.CommunicationEmail),
				Name:  stripe.String(account.Name),
			})
			if stripeErr != nil {
				httputil.HandleResponse(ctx, w, nil, stripeErr)
				return
			}
			stripeCustomerID = cust.ID

			if _, err := s.store.CreateStripeCustomer(ctx, account.ID, stripeCustomerID); err != nil {
				httputil.HandleResponse(ctx, w, nil, err)
				return
			}
		} else {
			httputil.HandleResponse(ctx, w, nil, err)
			return
		}
	} else {
		stripeCustomerID = stripeCustomer.CustomerID
	}

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(stripeCustomerID),
	}
	if _, err = paymentmethod.Attach(input.PaymentMethodID, params); err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	pm, err := s.store.CreatePaymentMethod(r.Context(), store.CreatePaymentMethodInput{
		AccountID: account.ID,
		Provider:  providerStripe,
		Data: map[string]any{
			"payment_method_id": input.PaymentMethodID,
		},
		ProviderCustomerID: stripeCustomerID,
	})
	if err != nil {
		httputil.HandleResponse(ctx, w, nil, err)
		return
	}

	httputil.New(w).JSON(pm, http.StatusCreated)
}

func (s *Service) Charge(ctx context.Context, invoiceID string) (*model.Payment, error) {
	invoice, err := s.store.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if invoice.Status == model.InvoiceStatusPaid {
		return nil, fmt.Errorf("invoice %s already paid", invoiceID)
	}

	paymentMethod, err := s.GetDefaultPaymentMethod(ctx, invoice.AccountID)
	if err != nil {
		return nil, fmt.Errorf("get default payment method: %w", err)
	}

	if !s.providerRegistered(paymentMethod.Provider) {
		return nil, fmt.Errorf("provider %s not found", paymentMethod.Provider)
	}

	switch paymentMethod.Provider {
	case providerStripe:
		return s.chargeStripe(ctx, invoiceID, invoice, paymentMethod)
	default:
		return nil, fmt.Errorf("provider %s not found", paymentMethod.Provider)
	}
}

func (s *Service) chargeStripe(
	ctx context.Context,
	invoiceID string,
	invoice *model.Invoice,
	paymentMethod *model.PaymentMethod,
) (*model.Payment, error) {
	// Create a pending payment record
	payment, err := s.store.CreatePayment(ctx, model.CreatePaymentInput{
		InvoiceID:       invoiceID,
		PaymentMethodID: paymentMethod.ID,
		AmountInCents:   invoice.AmountInCents,
		Currency:        invoice.Currency,
		Status:          model.PaymentStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("create payment record: %w", err)
	}

	// charge the payment
	var paymentMethodData map[string]any
	if err := json.Unmarshal([]byte(paymentMethod.Data), &paymentMethodData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment method data: %w", err)
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(invoice.AmountInCents),
		Currency:      stripe.String(invoice.Currency),
		PaymentMethod: stripe.String(paymentMethodData["payment_method_id"].(string)),
		Customer:      stripe.String(paymentMethod.ProviderCustomerID),
		Confirm:       stripe.Bool(true),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}

	charge, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("unable to create payment intent: %w", err)
	}

	chargeData, err := json.Marshal(charge)
	if err != nil {
		return nil, fmt.Errorf("marshal charge data: %w", err)
	}

	if charge.Status != stripe.PaymentIntentStatusSucceeded {
		if err := s.store.UpdatePayment(ctx, payment.ID, model.UpdatePaymentInput{
			Status:     lo.ToPtr(model.PaymentStatusFailed),
			ChargeData: lo.ToPtr(string(chargeData)),
		}); err != nil {
			return nil, fmt.Errorf("update failed payment: %w", err)
		}
	}

	// update payment status to paid
	if err := s.store.UpdatePayment(ctx, payment.ID, model.UpdatePaymentInput{
		Status:     lo.ToPtr(model.PaymentStatusPaid),
		ChargeData: lo.ToPtr(string(chargeData)),
	}); err != nil {
		return nil, fmt.Errorf("update failed payment: %w", err)
	}

	// update invoice status to paid
	if err := s.store.UpdateInvoice(ctx, invoiceID, model.UpdateInvoiceInput{
		Status: lo.ToPtr(model.InvoiceStatusPaid),
	}); err != nil {
		return nil, fmt.Errorf("update failed invoice: %w", err)
	}

	return payment, nil
}
