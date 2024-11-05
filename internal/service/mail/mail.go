package mail

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/go-mail/mail/v2"
	"github.com/google/uuid"
)

type service struct {
	confirmationCodeStore core.ConfirmationCodeStore
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
	confirmationCodeStore core.ConfirmationCodeStore,
	userStore core.UserStore) core.MailService {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return &service{
		confirmationCodeStore: confirmationCodeStore,
		dialer:                dialer,
		sender:                sender,
		userStore:             userStore,
	}
}

func (s *service) Send(ctx context.Context, user core.User) error {
	if user.Email == nil {
		logger.Log().Error(ctx, "email is nil")
		return core.ErrEmailNotProvided
	}

	code := uuid.New().String()
	err := s.confirmationCodeStore.SetConfirmationCode(ctx, user.ID, code, 5*time.Minute)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", *user.Email)
	msg.SetHeader("From", s.sender)
	msg.SetHeader("Subject", "Email confirmation")
	msg.SetBody("text/plain", code)

	for range 3 {
		if err = s.dialer.DialAndSend(msg); err == nil {
			break
		}
		logger.Log().Error(ctx, err.Error())
	}

	return err
}

func (s *service) Verify(ctx context.Context, code string) (*core.User, error) {
	userID, err := s.confirmationCodeStore.GetConfirmationCode(ctx, code)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	err = s.confirmationCodeStore.DeleteConfirmationCode(ctx, code)
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
