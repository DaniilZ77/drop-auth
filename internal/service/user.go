package service

import (
	"context"
	"errors"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserModifier interface {
	UpdateUser(ctx context.Context, user generated.UpdateUserParams) (*generated.User, error)
	SaveUser(ctx context.Context, user generated.SaveUserParams) (*uuid.UUID, error)
}

type UserProvider interface {
	GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *int, err error)
	GetUserByExternalID(ctx context.Context, id int32) (*generated.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*generated.User, error)
}

type RefreshTokenProvider interface {
	GetRefreshToken(ctx context.Context, tokenID string) (*string, error)
}

type RefreshTokenModifier interface {
	SetRefreshToken(ctx context.Context, userID string, tokenID string, expiry time.Duration) error
	ReplaceRefreshToken(ctx context.Context, oldID, newID, userID string, expiry time.Duration) error
}

type UserService struct {
	userModifier         UserModifier
	userProvider         UserProvider
	refreshTokenProvider RefreshTokenProvider
	refreshTokenModifier RefreshTokenModifier
	authConfig           model.AuthConfig
}

func New(
	userModifier UserModifier,
	userProvider UserProvider,
	refreshTokenProvider RefreshTokenProvider,
	refreshTokenModifier RefreshTokenModifier,
	authConfig model.AuthConfig,
) *UserService {
	return &UserService{
		userModifier:         userModifier,
		userProvider:         userProvider,
		refreshTokenProvider: refreshTokenProvider,
		refreshTokenModifier: refreshTokenModifier,
		authConfig:           authConfig,
	}
}

func (s *UserService) UpdateUser(ctx context.Context, updateUser generated.UpdateUserParams) (*generated.User, error) {
	user, err := s.userModifier.UpdateUser(ctx, updateUser)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *int, err error) {
	return s.userProvider.GetUsers(ctx, params)
}

func (s *UserService) generateToken(ctx context.Context, id uuid.UUID, expiry time.Duration) (*string, error) {
	data := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Minute * expiry).Unix(),
	})

	token, err := data.SignedString([]byte(s.authConfig.Secret))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return &token, nil
}

func (s *UserService) Login(ctx context.Context, saveUser generated.SaveUserParams) (accessToken *string, refreshToken *string, err error) {
	user, err := s.userProvider.GetUserByExternalID(ctx, saveUser.ExternalID)
	if err != nil && !errors.Is(err, model.ErrUserNotFound) {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	if errors.Is(err, model.ErrUserNotFound) {
		_, err := s.userModifier.SaveUser(ctx, saveUser)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, err
		}

		user, err = s.userProvider.GetUserByExternalID(ctx, saveUser.ExternalID)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, nil, err
		}
	}

	accessToken, err = s.generateToken(ctx, user.ID, time.Duration(s.authConfig.AccessTokenTTL))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	refreshToken = new(string)
	*refreshToken = uuid.NewString()
	if err := s.refreshTokenModifier.SetRefreshToken(ctx, user.ID.String(), *refreshToken, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL)); err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(ctx context.Context, token string) (accessToken *string, refreshToken *string, err error) {
	userID, err := s.refreshTokenProvider.GetRefreshToken(ctx, token)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	user, err := s.userProvider.GetUserByID(ctx, uuid.MustParse(*userID))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	accessToken, err = s.generateToken(ctx, user.ID, time.Duration(s.authConfig.AccessTokenTTL))
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	refreshToken = new(string)
	*refreshToken = uuid.NewString()
	if err = s.refreshTokenModifier.ReplaceRefreshToken(ctx, token, *refreshToken, *userID, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL)); err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*generated.User, error) {
	return s.userProvider.GetUserByID(ctx, uuid.MustParse(id))
}
