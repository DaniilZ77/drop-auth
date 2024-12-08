package model

import (
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model/validator"
	authv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/auth"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
)

func ValidateLoginRequest(v *validator.Validator, req *authv1.LoginRequest) {
	validateEmailOrTelephone(v, req.GetEmail(), req.GetTelephone())
	if req.GetEmail() != "" {
		validateEmail(v, req.GetEmail())
	}
	if req.GetTelephone() != "" {
		validatePhone(v, req.GetTelephone())
	}
	validatePassword(v, req.GetPassword())
}

func ValidateSignupRequest(v *validator.Validator, req *authv1.SignupRequest) {
	validateEmailOrTelephone(v, req.GetEmail().GetEmail(), req.GetTelephone().GetTelephone())
	if req.GetEmail().GetEmail() != "" {
		validateEmail(v, req.GetEmail().GetEmail())
		validateCode(v, req.GetEmail().GetCode())
	}
	if req.GetTelephone().GetTelephone() != "" {
		validatePhone(v, req.GetTelephone().GetTelephone())
		validateCode(v, req.GetTelephone().GetCode())
	}
	if req.GetMiddleName() != "" {
		validateName(v, "middle_name", req.GetMiddleName())
	}
	validateUsername(v, req.GetUsername())
	v.Check(validator.Between(len(req.GetFirstName()), 2, 128), "first_name", "length must be between 2 and 128")
	v.Check(validator.Between(len(req.GetLastName()), 2, 128), "last_name", "length must be between 2 and 128")
	validatePseudonym(v, req.GetPseudonym())
	validatePassword(v, req.GetPassword())
}

func ValidateUpdateUserRequest(v *validator.Validator, req *userv1.UpdateUserRequest) {
	v.Check(req.GetUpdateMask() != nil, "mask", "mask is required")
	for _, path := range req.GetUpdateMask().GetPaths() {
		switch path {
		case "username":
			validateUsername(v, req.User.GetUsername())
		case "email":
			validateEmail(v, req.GetUser().GetEmail().GetEmail())
			validateCode(v, req.GetUser().GetEmail().GetCode())
		case "telephone":
			validatePhone(v, req.GetUser().GetTelephone().GetTelephone())
			validateCode(v, req.GetUser().GetTelephone().GetCode())
		case "firstName":
			validateName(v, path, req.GetUser().GetFirstName())
		case "lastName":
			validateName(v, path, req.GetUser().GetLastName())
		case "middleName":
			validateName(v, path, req.GetUser().GetMiddleName())
		case "password":
			validatePassword(v, req.User.GetPassword().GetOldPassword())
			validatePassword(v, req.User.GetPassword().GetNewPassword())
		}
	}
}

func ValidateLoginTelegramRequest(v *validator.Validator, req *authv1.LoginTelegramRequest) {
	v.Check(validator.Between(len(req.GetPseudonym()), 2, 32), "pseudonym", "length must be between 2 and 32")
	if req.GetMiddleName() != "" {
		validateName(v, "middle_name", req.GetMiddleName())
	}
}

func ValidateResetPassword(v *validator.Validator, req *authv1.ResetPasswordRequest) {
	validateCode(v, req.GetCode())
	validatePassword(v, req.GetPassword())
}

func ValidateGetUserRequest(v *validator.Validator, req *userv1.GetUserRequest) {
	v.Check(req.GetUserId() > 0, "user_id", "must be positive")
}

func ValidateSendEmailRequest(v *validator.Validator, req *authv1.SendEmailRequest) {
	validateEmail(v, req.GetEmail())
}

func ValidateSendSMSRequest(v *validator.Validator, req *authv1.SendSMSRequest) {
	validatePhone(v, req.GetTelephone())
}

func validatePseudonym(v *validator.Validator, pseudonym string) {
	v.Check(validator.Between(len(pseudonym), 2, 32), "pseudonym", "length must be between 2 and 32")
}

func validateCode(v *validator.Validator, code string) {
	v.Check(code != "", "code", "must be not empty")
}

func validateName(v *validator.Validator, key, name string) {
	v.Check(validator.Between(len(name), 2, 128), key, "length must be between 2 and 128")
}

func validateEmailOrTelephone(v *validator.Validator, email, telephone string) {
	v.Check(email != "" || telephone != "", "email_or_telephone", "email or telephone must be provided")
}

func validateEmail(v *validator.Validator, email string) {
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be valid")
}

func validateUsername(v *validator.Validator, username string) {
	v.Check(validator.Between(len(username), 2, 32), "username", "length must be between 2 and 32")
}

func validatePassword(v *validator.Validator, password string) {
	v.Check(validator.AtLeast(len(password), 8), "password", "must contain at least 8 characters")
}

func validatePhone(v *validator.Validator, phone string) {
	v.Check(validator.Matches(phone, validator.PhoneRX), "telephone", "must be valid")
}
