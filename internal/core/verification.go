package core

import (
	"context"
	"time"
)

type (
	VerificationService interface {
		RegisterEmailService(host string, port int, username, password, sender string)
		RegisterSMSService(smsSender string)
		SendEmail(ctx context.Context, opt Option, ip string) error
		SendSMS(ctx context.Context, opt Option, ip string) error
	}

	VerificationStore interface {
		SetVerificationCode(ctx context.Context, key string, val VerificationCode, expiresIn time.Duration) error
		GetVerificationCode(ctx context.Context, key string) (*VerificationCode, error)
		DeleteVerificationCode(ctx context.Context, key string) error
	}

	Option func(context.Context, UserStore) (*User, error)

	VerificationCode struct {
		Value  string               `json:"value"`
		Type   VerificationCodeType `json:"type"`
		UserID int                  `json:"user_id"`
		IP     string               `json:"ip"`
	}
)

type VerificationCodeType int

const (
	Email VerificationCodeType = iota
	Telephone
)

func (vct VerificationCodeType) ToString() string {
	if vct == 0 {
		return "email"
	}
	return "telephone"
}
