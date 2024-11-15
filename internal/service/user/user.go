package user

import (
	"context"
	"fmt"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	userStorage         core.UserStore
	verificationStorage core.VerificationStore
}

func New(userStorage core.UserStore, verificationStore core.VerificationStore) core.UserService {
	return &service{userStorage: userStorage, verificationStorage: verificationStore}
}

func (s *service) DeleteUser(ctx context.Context, userID int) error {
	err := s.userStorage.DeleteUser(ctx, userID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return nil
}

func (s *service) UpdateUser(ctx context.Context, user core.UpdateUser, updateCodes core.UpdateCodes) (*core.User, error) {
	userFromDB, err := s.userStorage.GetUserByID(ctx, user.ID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	if userFromDB.IsDeleted {
		logger.Log().Debug(ctx, core.ErrAlreadyDeleted.Error())
		return nil, core.ErrAlreadyDeleted
	}

	if user.Password != nil {
		newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		user.Password.NewPassword = string(newPasswordHash)

		if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.PasswordHash), []byte(user.Password.OldPassword)); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, core.ErrInvalidCredentials
		}
	}

	check := func(userCode string, value string, verificationCodeType core.VerificationCodeType) error {
		verificationCode, err := s.verificationStorage.GetVerificationCode(ctx, userCode)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return fmt.Errorf("%s: %w", verificationCodeType.ToString(), err)
		}

		err = s.verificationStorage.DeleteVerificationCode(ctx, userCode)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return err
		}

		if verificationCode.Value != value {
			logger.Log().Error(ctx, core.ErrVerificationCodeNotValid.Error())
			return fmt.Errorf("%s: %w", verificationCodeType.ToString(), core.ErrVerificationCodeNotValid)
		}

		return nil
	}

	isVerified := true
	if user.Email != nil {
		if updateCodes.EmailCode == nil {
			logger.Log().Error(ctx, "email code is nil")
			return nil, core.ErrInternal
		}

		if err = check(*updateCodes.EmailCode, *user.Email, core.Email); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		user.IsEmailVerified = &isVerified
	}
	if user.Telephone != nil {
		if updateCodes.TelephoneCode == nil {
			logger.Log().Error(ctx, "telephone code is nil")
			return nil, core.ErrInternal
		}

		if err = check(*updateCodes.TelephoneCode, *user.Telephone, core.Telephone); err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}

		user.IsTelephoneVerified = &isVerified
	}

	retUser, err := s.userStorage.UpdateUser(ctx, user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return retUser, nil
}

func (s *service) GetUser(ctx context.Context, user core.User) (*core.User, error) {
	retUser, err := s.userStorage.GetUserByID(ctx, user.ID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return retUser, nil
}
