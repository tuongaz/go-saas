package emailer

import (
	"context"
	"io"
)

var _ Interface = (*Resend)(nil)

type Attachment struct {
	Filename    string
	ContentType string
	Content     io.Reader
}

type SendEmailInput struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	ReplyTo     string
	Subject     string
	HTML        string
	Attachments []Attachment
}

type SendEmailOutput struct {
	ID string
}

type Interface interface {
	Send(ctx context.Context, request SendEmailInput) (*SendEmailOutput, error)
}
