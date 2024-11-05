package auth

import (
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	authv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/auth"
)

func ToSignupResponse(user core.User) *authv1.SignupResponse {
	var email, telephone, middleName string
	if user.Email != nil {
		email = *user.Email
	}
	if user.Telephone != nil {
		telephone = *user.Telephone
	}
	if user.MiddleName != nil {
		middleName = *user.MiddleName
	}

	return &authv1.SignupResponse{
		UserId:     int64(user.ID),
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: middleName,
		Telephone:  telephone,
		Email:      email,
		Pseudonym:  user.Pseudonym,
	}
}

func ToRefreshTokenResponse(accessToken, refreshToken string) *authv1.RefreshTokenResponse {
	return &authv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func ToLoginResponse(accessToken, refreshToken string) *authv1.LoginResponse {
	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func FromLoginRequest(req *authv1.LoginRequest) *core.User {
	var email, telephone *string
	if req.GetEmail() != "" {
		email = &req.Email
	}
	if req.GetTelephone() != "" {
		telephone = &req.Telephone
	}
	return &core.User{
		Email:        email,
		Telephone:    telephone,
		PasswordHash: req.GetPassword(),
	}
}

func FromSignupRequest(req *authv1.SignupRequest) *core.User {
	var email, telephone, middleName *string
	if req.GetEmail() != "" {
		email = &req.Email
	}
	if req.GetTelephone() != "" {
		telephone = &req.Telephone
	}
	if req.GetMiddleName() != "" {
		middleName = &req.MiddleName
	}

	return &core.User{
		Username:     req.GetUsername(),
		Email:        email,
		PasswordHash: req.GetPassword(),
		Pseudonym:    req.GetPseudonym(),
		FirstName:    req.GetFirstName(),
		LastName:     req.GetLastName(),
		MiddleName:   middleName,
		Telephone:    telephone,
	}
}

func ToValidateTokenResponse(isValid bool, userID int) *authv1.ValidateTokenResponse {
	return &authv1.ValidateTokenResponse{
		IsValid: isValid,
		UserId:  int64(userID),
	}
}
