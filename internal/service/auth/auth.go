package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/jwt"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	userStorage         core.UserStore
	refreshTokenStorage core.RefreshTokenStore
	verificationStorage core.VerificationStore
	authConfig          core.AuthConfig
}

func NewConfig(secret string, accessTokenTTL, refreshTokenTTL int) core.AuthConfig {
	return core.AuthConfig{
		Secret:          secret,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}

func New(
	userStorage core.UserStore,
	refreshTokenStorage core.RefreshTokenStore,
	authConfig core.AuthConfig,
	verificationStorage core.VerificationStore,
) core.AuthService {
	return &service{
		userStorage:         userStorage,
		refreshTokenStorage: refreshTokenStorage,
		authConfig:          authConfig,
		verificationStorage: verificationStorage,
	}
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (accesstoken, refreshtoken *string, err error) {
	userID, err := s.refreshTokenStorage.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	err = s.refreshTokenStorage.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	userFromDB, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	if userFromDB.IsDeleted {
		logger.Log().Error(ctx, core.ErrAlreadyDeleted.Error())
		return nil, nil, core.ErrAlreadyDeleted
	}

	accessToken, err := jwt.GenerateToken(userFromDB.ID, s.authConfig)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	newRefreshToken := uuid.New().String()
	err = s.refreshTokenStorage.SetRefreshToken(ctx, userFromDB.ID, newRefreshToken, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return accessToken, &newRefreshToken, nil
}

func (s *service) Login(ctx context.Context, user core.User) (accesstoken, refreshtoken *string, err error) {
	var userFromDB *core.User
	if user.Email != nil {
		userFromDB, err = s.userStorage.GetUserByEmail(ctx, *user.Email)
	} else if user.Telephone != nil {
		userFromDB, err = s.userStorage.GetUserByTelephone(ctx, *user.Telephone)
	} else {
		return nil, nil, core.ErrEmailAndTelephoneNotProvided
	}
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrUserNotFound) {
			return nil, nil, core.ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if userFromDB.IsDeleted {
		logger.Log().Error(ctx, core.ErrAlreadyDeleted.Error())
		return nil, nil, core.ErrAlreadyDeleted
	}

	if user.Telephone != nil && (userFromDB.Telephone == nil || *user.Telephone != *userFromDB.Telephone) {
		logger.Log().Error(ctx, "telephone: %s", core.ErrInvalidCredentials.Error())
		return nil, nil, core.ErrInvalidCredentials
	}

	if userFromDB.Email == nil && user.Email != nil {
		logger.Log().Error(ctx, "email: %s", core.ErrInvalidCredentials.Error())
		return nil, nil, core.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFromDB.PasswordHash), []byte(user.PasswordHash))
	if err != nil {
		logger.Log().Error(ctx, "invalid password: %s", err.Error())
		return nil, nil, core.ErrInvalidCredentials
	}

	accessToken, err := jwt.GenerateToken(userFromDB.ID, s.authConfig)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	refreshToken := uuid.New().String()
	err = s.refreshTokenStorage.SetRefreshToken(ctx, userFromDB.ID, refreshToken, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL))
	if err != nil {
		logger.Log().Info(ctx, err.Error())
		return nil, nil, err
	}

	return accessToken, &refreshToken, nil
}

func (s *service) Signup(ctx context.Context, emailCode, telephoneCode string, user core.User, ip string) (*core.User, error) {
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

		if verificationCode.Value != value || verificationCode.IP != ip {
			logger.Log().Error(ctx, core.ErrVerificationCodeNotValid.Error())
			return fmt.Errorf("%s: %w", verificationCodeType.ToString(), core.ErrVerificationCodeNotValid)
		}

		return nil
	}

	if user.Email != nil {
		err := check(emailCode, *user.Email, core.Email)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}
	}

	if user.Telephone != nil {
		err := check(telephoneCode, *user.Telephone, core.Telephone)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	user.PasswordHash = string(passwordHash)

	userID, err := s.userStorage.AddUser(ctx, user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}
	user.ID = userID
	user.IsDeleted = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return &user, nil
}

func (s *service) ResetPassword(ctx context.Context, code, password string) (*core.User, error) {
	verificationCode, err := s.verificationStorage.GetVerificationCode(ctx, code)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	if err = s.verificationStorage.DeleteVerificationCode(ctx, code); err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	if verificationCode.UserID <= 0 {
		logger.Log().Error(ctx, core.ErrVerificationCodeNotValid.Error())
		return nil, core.ErrVerificationCodeNotValid
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	updateUser := core.UpdateUser{
		ID: verificationCode.UserID,
		Password: &core.UpdatePassword{
			NewPassword: string(hashedPassword),
		},
	}

	user, err := s.userStorage.UpdateUser(ctx, updateUser)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	logger.Log().Debug(ctx, "user: %v", user)

	return user, nil
}

func (s *service) LoginExternal(ctx context.Context, user core.User, externalUser core.ExternalUser, provider core.AuthProvider, isValid bool) (accessToken *string, refreshToken *string, err error) {
	generateTokens := func(userID int) (*string, *string, error) {
		accessToken, err := jwt.GenerateToken(userID, s.authConfig)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, err
		}

		refreshToken := uuid.New().String()
		err = s.refreshTokenStorage.SetRefreshToken(ctx, userID, refreshToken, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL))
		if err != nil {
			logger.Log().Info(ctx, err.Error())
			return nil, nil, err
		}
		return accessToken, &refreshToken, nil
	}

	userFromDB, err := s.userStorage.GetUserByExternalID(ctx, externalUser.ExternalID)
	if err == nil {
		return generateTokens(userFromDB.UserID)
	}

	if errors.Is(err, core.ErrUserNotFound) {
		if !isValid {
			logger.Log().Debug(ctx, err.Error())
			return nil, nil, core.ErrValidationFailed
		}
	} else {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	userID, err := s.userStorage.AddExternalUser(ctx, user, externalUser)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return generateTokens(userID)
}
