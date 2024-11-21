package verification

import (
	"context"
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

func (s *service) SendEmail(ctx context.Context, opt core.Option, ip string) error {
	user, err := opt(ctx, s.userStore)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	if err := setCodeAndSendMail(ctx, s.verificationStore, s.emailDialer, s.emailSender, *user, ip); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return err
}

func (s *service) SendSMS(ctx context.Context, opt core.Option, ip string) error {
	user, err := opt(ctx, s.userStore)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	if err = setCodeAndSendSMS(ctx, s.verificationStore, s.smsSender, *user, ip); err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return err
}
