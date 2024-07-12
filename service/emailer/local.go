package emailer

import (
	"context"
	"fmt"
)

func NewLocal() *Local {
	return &Local{}
}

type Local struct {
}

func (*Local) Send(ctx context.Context, input SendEmailInput) (*SendEmailOutput, error) {
	fmt.Printf("Sending email: %+v\n", input)
	return &SendEmailOutput{
		ID: "local",
	}, nil
}
