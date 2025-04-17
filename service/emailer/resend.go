package emailer

import (
	"context"
	"fmt"
	"io"

	"github.com/resend/resend-go/v2"
)

func NewResend(apiKey string) *Resend {
	return &Resend{
		client: resend.NewClient(apiKey),
	}
}

type Resend struct {
	client *resend.Client
}

func (r *Resend) Send(ctx context.Context, input SendEmailInput) (*SendEmailOutput, error) {
	var attachments []*resend.Attachment
	for _, att := range input.Attachments {
		if att.Content != nil {
			data, err := io.ReadAll(att.Content)
			if err != nil {
				return nil, fmt.Errorf("read attachment content: %w", err)
			}
			attachments = append(attachments, &resend.Attachment{
				Content:     data,
				Filename:    att.Filename,
				ContentType: att.ContentType,
			})
		}
	}

	params := &resend.SendEmailRequest{
		From:        input.From,
		To:          input.To,
		Html:        input.HTML,
		Subject:     input.Subject,
		Cc:          input.Cc,
		Bcc:         input.Bcc,
		ReplyTo:     input.ReplyTo,
		Attachments: attachments,
	}

	sent, err := r.client.Emails.Send(params)
	if err != nil {
		return nil, fmt.Errorf("send email with Resend: %w", err)
	}

	return &SendEmailOutput{
		ID: sent.Id,
	}, nil
}
