package core

import "errors"

var (
	// auth and user
	ErrUserAlreadyExists            = errors.New("user already exists")
	ErrInvalidAuthConfig            = errors.New("invalid secret")
	ErrInvalidCredentials           = errors.New("invalid credentials")
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrUserNotFound                 = errors.New("user not found")
	ErrInternal                     = errors.New("internal error")
	ErrEmailAlreadyExists           = errors.New("email already exists")
	ErrUsernameAlreadyExists        = errors.New("username already exists")
	ErrRefreshTokenNotValid         = errors.New("refresh token not valid")
	ErrAlreadyDeleted               = errors.New("already deleted")
	ErrAlreadyExists                = errors.New("already exists")
	ErrTelephoneAlreadyExists       = errors.New("telephone already exists")
	ErrEmailOrTelephoneNotProvided  = errors.New("email or telephone should be provided")
	ErrVerificationCodeNotValid     = errors.New("verification code not valid")
	ErrEmailNotProvided             = errors.New("email is not provided")
	ErrEmailAndTelephoneNotVerified = errors.New("email and telephone not verified")

	// validation
	ErrValidationFailed = errors.New("validation failed")
)
