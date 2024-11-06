package mail

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/go-mail/mail/v2"
)

type service struct {
	verificationCodeStore core.VerificationCodeStore
	userStore             core.UserStore
	dialer                *mail.Dialer
	sender                string
}

func New(
	host string,
	port int,
	username,
	password,
	sender string,
	verificationCodeStore core.VerificationCodeStore,
	userStore core.UserStore,
) core.MailService {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return &service{
		verificationCodeStore: verificationCodeStore,
		dialer:                dialer,
		sender:                sender,
		userStore:             userStore,
	}
}

func (s *service) Send(ctx context.Context, user core.User) error {
	if err := setCodeAndSendMail(ctx, s.verificationCodeStore, s.dialer, s.sender, user); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

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

	isEmailVerified := true
	user := core.UpdateUser{
		ID:              userID,
		IsEmailVerified: &isEmailVerified,
	}

	retUser, err := s.userStore.UpdateUser(ctx, user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return retUser, nil
}

func (s *service) Resend(ctx context.Context, email string) error {
	user, err := s.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	if err := setCodeAndSendMail(ctx, s.verificationCodeStore, s.dialer, s.sender, *user); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return err
}
