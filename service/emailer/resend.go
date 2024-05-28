package emailer

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

func New(apiKey string) *Resend {
	return &Resend{
		client: resend.NewClient(apiKey),
	}
}

type Resend struct {
	client *resend.Client
}

func (r *Resend) Send(ctx context.Context, input SendEmailInput) (*SendEmailOutput, error) {
	params := &resend.SendEmailRequest{
		From:    input.From,
		To:      input.To,
		Html:    input.HTML,
		Subject: input.Subject,
		Cc:      input.Cc,
		Bcc:     input.Bcc,
		ReplyTo: input.ReplyTo,
	}

	sent, err := r.client.Emails.Send(params)
	if err != nil {
		return nil, fmt.Errorf("send email with Resend: %w", err)
	}

	return &SendEmailOutput{
		ID: sent.Id,
	}, nil
}
