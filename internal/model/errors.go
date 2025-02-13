package model

import "errors"

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrRefreshTokenNotValid = errors.New("refresh token not valid")
	ErrUserNotFound         = errors.New("user not found")
)
