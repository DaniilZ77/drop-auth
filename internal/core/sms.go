package core

import "context"

type (
	SMSService interface {
		Send(ctx context.Context, user User) error
		Resend(ctx context.Context, telephone string) error
		Verify(ctx context.Context, code string) (*User, error)
	}
)
