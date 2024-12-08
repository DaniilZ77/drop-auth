package core

import (
	"context"
	"time"
)

type (
	AuthService interface {
		Login(ctx context.Context, user User) (accessToken *string, refreshToken *string, err error)
		Signup(ctx context.Context, emailCode, telephoneCode string, user User, ip string) (*User, error)
		RefreshToken(ctx context.Context, refreshToken string) (*string, *string, error)
		ResetPassword(ctx context.Context, code, password string) (*User, error)
		LoginExternal(ctx context.Context, user User, externalUser ExternalUser, provider AuthProvider, isValid bool) (accessToken *string, refreshToken *string, err error)
	}

	AuthConfig struct {
		Secret          string
		AccessTokenTTL  int
		RefreshTokenTTL int
	}

	RefreshTokenStore interface {
		SetRefreshToken(ctx context.Context, userID int, tokenID string, expiresIn time.Duration) error
		GetRefreshToken(ctx context.Context, tokenID string) (int, error)
		DeleteRefreshToken(ctx context.Context, prevTokenID string) error
	}

	AuthProvider string
)

const (
	TelegramAuthProvider AuthProvider = "tg"
)
