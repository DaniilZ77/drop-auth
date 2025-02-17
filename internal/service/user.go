package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserModifier interface {
	UpdateUser(ctx context.Context, user generated.UpdateUserParams) (*generated.User, error)
	SaveUser(ctx context.Context, user generated.SaveUserParams) (*uuid.UUID, error)
	SaveAdmin(ctx context.Context, params generated.SaveAdminParams) error
	DeleteAdmin(ctx context.Context, userID uuid.UUID) error
}

type UserProvider interface {
	GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *int, err error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*generated.User, error)
	GetUserByUsername(ctx context.Context, username string) (*generated.User, error)
	GetAdminByID(ctx context.Context, id uuid.UUID) (*generated.GetAdminByIDRow, error)
	GetUserByExternalID(ctx context.Context, id int32) (*generated.User, error)
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
	log                  *slog.Logger
}

func New(
	userModifier UserModifier,
	userProvider UserProvider,
	refreshTokenProvider RefreshTokenProvider,
	refreshTokenModifier RefreshTokenModifier,
	authConfig model.AuthConfig,
	log *slog.Logger,
) *UserService {
	return &UserService{
		userModifier:         userModifier,
		userProvider:         userProvider,
		refreshTokenProvider: refreshTokenProvider,
		refreshTokenModifier: refreshTokenModifier,
		authConfig:           authConfig,
		log:                  log,
	}
}

func (s *UserService) UpdateUser(ctx context.Context, updateUser generated.UpdateUserParams) (*generated.User, error) {
	user, err := s.userModifier.UpdateUser(ctx, updateUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *int, err error) {
	return s.userProvider.GetUsers(ctx, params)
}

func (s *UserService) generateToken(id uuid.UUID, scale generated.NullAdminScale, expiry time.Duration) (*string, error) {
	claims := jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Minute * expiry).Unix(),
	}

	if scale.Valid {
		claims["admin"] = scale.AdminScale
	}

	data := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := data.SignedString([]byte(s.authConfig.Secret))
	if err != nil {
		s.log.Error("failed to sign token", sl.Err(err))
		return nil, err
	}

	return &token, nil
}

func (s *UserService) Login(ctx context.Context, saveUser generated.SaveUserParams) (accessToken *string, refreshToken *string, err error) {
	user, err := s.userProvider.GetUserByExternalID(ctx, saveUser.ExternalID)
	if err != nil && !errors.Is(err, model.ErrUserNotFound) {
		s.log.Error("failed to get user", sl.Err(err))
		return nil, nil, err
	}

	if errors.Is(err, model.ErrUserNotFound) {
		_, err := s.userModifier.SaveUser(ctx, saveUser)
		if err != nil {
			s.log.Error("failed to save user", sl.Err(err))
			return nil, nil, err
		}

		user, err = s.userProvider.GetUserByExternalID(ctx, saveUser.ExternalID)
		if err != nil {
			s.log.Error("failed to get user", sl.Err(err))
			return nil, nil, err
		}
	}

	admin, err := s.userProvider.GetAdminByID(ctx, user.ID)
	if err != nil {
		s.log.Error("failed to get admin", sl.Err(err))
		return nil, nil, err
	}

	accessToken, err = s.generateToken(user.ID, admin.Scale, time.Duration(s.authConfig.AccessTokenTTL))
	if err != nil {
		s.log.Error("failed to generate token", sl.Err(err))
		return nil, nil, err
	}

	refreshToken = new(string)
	*refreshToken = uuid.NewString()
	if err := s.refreshTokenModifier.SetRefreshToken(ctx, user.ID.String(), *refreshToken, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL)); err != nil {
		s.log.Error("failed to set refresh token", sl.Err(err))
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(ctx context.Context, token string) (accessToken *string, refreshToken *string, err error) {
	userID, err := s.refreshTokenProvider.GetRefreshToken(ctx, token)
	if err != nil {
		s.log.Error("failed to get refresh token", sl.Err(err))
		return nil, nil, err
	}

	user, err := s.userProvider.GetAdminByID(ctx, uuid.MustParse(*userID))
	if err != nil {
		s.log.Error("failed to get admin", sl.Err(err))
		return nil, nil, err
	}

	accessToken, err = s.generateToken(user.ID, user.Scale, time.Duration(s.authConfig.AccessTokenTTL))
	if err != nil {
		s.log.Error("failed to generate token", sl.Err(err))
		return nil, nil, err
	}

	refreshToken = new(string)
	*refreshToken = uuid.NewString()
	if err = s.refreshTokenModifier.ReplaceRefreshToken(ctx, token, *refreshToken, *userID, time.Minute*time.Duration(s.authConfig.RefreshTokenTTL)); err != nil {
		s.log.Error("failed to replace refresh token", sl.Err(err))
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*generated.User, error) {
	return s.userProvider.GetUserByID(ctx, uuid.MustParse(id))
}

func (s *UserService) AddAdmin(ctx context.Context, username string, scale generated.AdminScale) error {
	if scale != generated.AdminScaleMajor {
		s.log.Debug("scale of admin is not major")
		return model.ErrAdminNotMajor
	}

	user, err := s.userProvider.GetUserByUsername(ctx, username)
	if err != nil {
		s.log.Error("failed to get user", sl.Err(err))
		return err
	}

	return s.userModifier.SaveAdmin(ctx, generated.SaveAdminParams{
		UserID: user.ID,
		Scale:  generated.AdminScaleMinor,
	})
}

func (s *UserService) DeleteAdmin(ctx context.Context, username string, scale generated.AdminScale) error {
	if scale != generated.AdminScaleMajor {
		s.log.Debug("scale of admin is not major")
		return model.ErrAdminNotMajor
	}

	user, err := s.userProvider.GetUserByUsername(ctx, username)
	if err != nil {
		s.log.Error("failed to get user", sl.Err(err))
		return err
	}

	admin, err := s.userProvider.GetAdminByID(ctx, user.ID)
	if err != nil {
		s.log.Error("failed to get admin", sl.Err(err))
		return err
	}

	if admin.Scale.Valid && admin.Scale.AdminScale == generated.AdminScaleMinor {
		s.log.Debug("cannot delete major admin")
		return model.ErrCannotDeleteMajorAdmin
	}

	return s.userModifier.DeleteAdmin(ctx, user.ID)
}

func (s *UserService) InitAdmin(ctx context.Context, username string) error {
	user, err := s.userProvider.GetUserByUsername(ctx, username)
	if err != nil {
		s.log.Error("failed to get user", sl.Err(err))
		return err
	}

	return s.userModifier.SaveAdmin(ctx, generated.SaveAdminParams{
		UserID: user.ID,
		Scale:  generated.AdminScaleMajor,
	})
}
