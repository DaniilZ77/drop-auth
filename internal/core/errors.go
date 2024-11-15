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
	ErrEmailAndTelephoneNotProvided = errors.New("email and telephone should be provided")
	ErrVerificationCodeNotValid     = errors.New("verification code not valid")
	ErrEmailNotProvided             = errors.New("email is not provided")
	ErrTelephoneNotProvided         = errors.New("telephone is not provided")
	ErrEmailAndTelephoneNotVerified = errors.New("email and telephone not verified")
	ErrEmailNotVerified             = errors.New("email not verified")
	ErrTelephoneNotVerified         = errors.New("telephone not verified")

	// validation
	ErrValidationFailed = errors.New("validation failed")
)

const (
	KeyEmailNotVerified     = "EMAIL_NOT_VERIFIED"
	KeyTelephoneNotVerified = "TELEPHONE_NOT_VERIFIED"
	KeyInvalidCredentials   = "INVALID_CREDENTIALS"
)
