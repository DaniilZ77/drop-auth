package core

import (
	"context"
	"time"
)

type (
	VerificationService interface {
		RegisterEmailService(host string, port int, username, password, sender string)
		RegisterSMSService(smsSender string)
		SendEmail(ctx context.Context, opt Option) error
		SendSMS(ctx context.Context, opt Option) error
		Verify(ctx context.Context, code string) (*User, error)
	}

	VerificationStore interface {
		SetVerificationCode(ctx context.Context, key string, val string, expiresIn time.Duration) error
		GetVerificationCode(ctx context.Context, key string) (string, error)
		DeleteVerificationCode(ctx context.Context, key string) error
	}

	Option func(context.Context, UserStore) (*User, error)
)
