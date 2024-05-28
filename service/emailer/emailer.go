package emailer

import (
	"context"
)

var _ Interface = (*Resend)(nil)

type SendEmailInput struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	ReplyTo string
	Subject string
	HTML    string
}

type SendEmailOutput struct {
	ID string
}

type Interface interface {
	Send(ctx context.Context, request SendEmailInput) (*SendEmailOutput, error)
}
