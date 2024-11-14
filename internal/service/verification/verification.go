package verification

import (
	"context"
	"strconv"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/go-mail/mail/v2"
)

type service struct {
	verificationStore core.VerificationStore
	userStore         core.UserStore
	emailDialer       *mail.Dialer
	emailSender       string
	smsSender         string
}

func New(
	verificationStore core.VerificationStore,
	userStore core.UserStore,
) core.VerificationService {
	return &service{
		verificationStore: verificationStore,
		userStore:         userStore,
	}
}

func (s *service) RegisterEmailService(
	host string,
	port int,
	username,
	password,
	sender string,
) {
	s.emailDialer = mail.NewDialer(host, port, username, password)
	s.emailSender = sender
	s.emailDialer.Timeout = 5 * time.Second
}

func (s *service) RegisterSMSService(smsSender string) {
	s.smsSender = smsSender
}

func (s *service) Verify(ctx context.Context, code string) (*core.User, error) {
	userIDStr, err := s.verificationStore.GetVerificationCode(ctx, code)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	userID, err := strconv.Atoi(userIDStr[1:])
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	var isEmailVerified *bool
	var isTelephoneVerified *bool
	if userIDStr[0] == 'e' {
		isEmailVerified = new(bool)
		*isEmailVerified = true
	} else {
		isTelephoneVerified = new(bool)
		*isTelephoneVerified = true
	}

	err = s.verificationStore.DeleteVerificationCode(ctx, code)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	user := core.UpdateUser{
		ID:                  userID,
		IsEmailVerified:     isEmailVerified,
		IsTelephoneVerified: isTelephoneVerified,
	}

	retUser, err := s.userStore.UpdateUser(ctx, user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return retUser, nil
}

func (s *service) SendEmail(ctx context.Context, opt core.Option) error {
	user, err := opt(ctx, s.userStore)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	if err := setCodeAndSendMail(ctx, s.verificationStore, s.emailDialer, s.emailSender, *user); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return err
}

func (s *service) SendSMS(ctx context.Context, opt core.Option) error {
	user, err := opt(ctx, s.userStore)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	if err = setCodeAndSendSMS(ctx, s.verificationStore, s.smsSender, *user); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return err
}
