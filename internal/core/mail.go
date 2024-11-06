package core

import (
	"context"
	"time"
)

type (
	MailService interface {
		Send(ctx context.Context, user User) error
		Resend(ctx context.Context, email string) error
		Verify(ctx context.Context, code string) (*User, error)
	}

	VerificationCodeStore interface {
		SetVerificationCode(ctx context.Context, key string, val string, expiresIn time.Duration) error
		GetVerificationCode(ctx context.Context, key string) (int, error)
		DeleteVerificationCode(ctx context.Context, key string) error
	}
)
