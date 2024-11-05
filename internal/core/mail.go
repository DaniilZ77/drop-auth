package core

import "context"

type MailService interface {
	Send(ctx context.Context, user User) error
	Verify(ctx context.Context, code string) (*User, error)
}
