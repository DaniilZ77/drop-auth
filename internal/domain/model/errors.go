package model

import "errors"

var (
	ErrUnauthorized           = errors.New("unauthorized")
	ErrRefreshTokenNotValid   = errors.New("refresh token not valid")
	ErrUserNotFound           = errors.New("user not found")
	ErrAdminAlreadyExists     = errors.New("admin already exists")
	ErrAdminNotMajor          = errors.New("admin must be major")
	ErrCannotDeleteMajorAdmin = errors.New("cannot delete major admin")
	ErrOrderByInvalidField    = errors.New("orderBy: invalid field")
	ErrAdminNotFound          = errors.New("admin not found")
	ErrEmptyPseudonym         = errors.New("empty pseudonym")
)
