package mail

import (
	"context"
	"fmt"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/go-mail/mail/v2"
	"github.com/google/uuid"
)

func setCodeAndSendMail(
	ctx context.Context,
	store core.VerificationCodeStore,
	dialer *mail.Dialer,
	sender string,
	user core.User,
) error {
	if user.Email == nil {
		logger.Log().Error(ctx, "email is nil")
		return core.ErrEmailNotProvided
	}

	code := uuid.New().String()
	err := store.SetVerificationCode(ctx, code, fmt.Sprintf("%d", user.ID), 5*time.Minute)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", *user.Email)
	msg.SetHeader("From", sender)
	msg.SetHeader("Subject", "Email confirmation")
	msg.SetBody("text/plain", code)

	for range 3 {
		if err = dialer.DialAndSend(msg); err == nil {
			break
		}
		logger.Log().Error(ctx, err.Error())
	}

	return nil
}
