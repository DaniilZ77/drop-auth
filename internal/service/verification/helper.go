package verification

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	helpers "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/service"
	"github.com/go-mail/mail/v2"
)

func setCodeAndSendMail(
	ctx context.Context,
	store core.VerificationStore,
	dialer *mail.Dialer,
	sender string,
	user core.User,
	ip string,
) error {
	if user.Email == nil {
		logger.Log().Debug(ctx, "email is nil")
		return core.ErrEmailNotProvided
	}

	code, err := helpers.GenCode(6)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	val := core.VerificationCode{
		UserID: user.ID,
		Type:   core.Email,
		Value:  *user.Email,
		IP:     ip,
	}

	err = store.SetVerificationCode(ctx, code, val, 5*time.Minute)
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

func setCodeAndSendSMS(
	ctx context.Context,
	store core.VerificationStore,
	sender string,
	user core.User,
	ip string,
) error {
	if user.Telephone == nil {
		logger.Log().Debug(ctx, core.ErrTelephoneNotProvided.Error())
		return core.ErrTelephoneNotProvided
	}

	code, err := helpers.GenCode(6)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	val := core.VerificationCode{
		UserID: user.ID,
		Type:   core.Telephone,
		Value:  *user.Telephone,
		IP:     ip,
	}

	if err = store.SetVerificationCode(ctx, code, val, 5*time.Minute); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	logger.Log().Debug(ctx, "code sended to %s: %s; from %s", *user.Telephone, code, sender)

	return nil
}
