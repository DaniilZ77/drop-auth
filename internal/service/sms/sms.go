package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/google/uuid"
)

type service struct {
	verificationCodeStore core.VerificationCodeStore
	userStore             core.UserStore
	sender                string
}

func New(
	sender string,
	verificationCodeStore core.VerificationCodeStore,
	userStore core.UserStore,
) core.SMSService {
	return &service{
		verificationCodeStore: verificationCodeStore,
		sender:                sender,
		userStore:             userStore,
	}
}

func (s *service) Resend(ctx context.Context, telephone string) error {
	user, err := s.userStore.GetUserByTelephone(ctx, telephone)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	if user.Telephone == nil {
		logger.Log().Debug(ctx, core.ErrTelephoneNotProvided.Error())
		return core.ErrTelephoneNotProvided
	}

	code := uuid.New().String()
	if err = s.verificationCodeStore.SetVerificationCode(ctx, code, fmt.Sprintf("%d", user.ID), 5*time.Minute); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	logger.Log().Debug(ctx, "code sended to %s: %s; from %s", user.Telephone, code, s.sender)

	return err
}

func (s *service) Send(ctx context.Context, user core.User) error {
	if user.Telephone == nil {
		logger.Log().Debug(ctx, core.ErrTelephoneNotProvided.Error())
		return core.ErrTelephoneNotProvided
	}

	code := uuid.New().String()
	if err := s.verificationCodeStore.SetVerificationCode(ctx, code, fmt.Sprintf("%d", user.ID), 5*time.Minute); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	logger.Log().Debug(ctx, "code sended to %s: %s; from %s", user.Telephone, code, s.sender)

	return nil
}

func (s *service) Verify(ctx context.Context, code string) (*core.User, error) {
	userID, err := s.verificationCodeStore.GetVerificationCode(ctx, code)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	err = s.verificationCodeStore.DeleteVerificationCode(ctx, code)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	isTelephoneVerified := true
	user := core.UpdateUser{
		ID:                  userID,
		IsTelephoneVerified: &isTelephoneVerified,
	}

	retUser, err := s.userStore.UpdateUser(ctx, user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return retUser, nil
}
