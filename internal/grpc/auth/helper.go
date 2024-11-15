package auth

import (
	"context"
	"errors"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
)

func sendEmailOrTelephone(ctx context.Context, optEmail core.Option, optTelephone core.Option, service core.VerificationService, ip string) {
	if err := service.SendEmail(ctx, optEmail, ip); err != nil &&
		!errors.Is(err, core.ErrEmailNotProvided) {
		logger.Log().Error(ctx, err.Error())
	}
	if err := service.SendSMS(ctx, optTelephone, ip); err != nil &&
		!errors.Is(err, core.ErrTelephoneNotProvided) {
		logger.Log().Error(ctx, err.Error())
	}
}
