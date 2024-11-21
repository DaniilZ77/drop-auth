package user

import (
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func FromUpdateUserRequest(req *userv1.UpdateUserRequest, userID int) (*core.UpdateUser, *core.UpdateCodes) {
	user := new(core.UpdateUser)
	user.ID = userID

	updateCodes := new(core.UpdateCodes)

	for _, path := range req.UpdateMask.Paths {
		switch path {
		case "username":
			user.Username = &req.GetUser().Username
		case "email":
			user.Email = &req.GetUser().GetEmail().Email
			updateCodes.EmailCode = &req.GetUser().GetEmail().Code
		case "password":
			user.Password = new(core.UpdatePassword)
			user.Password.OldPassword = req.GetUser().GetPassword().GetOldPassword()
			user.Password.NewPassword = req.GetUser().GetPassword().GetNewPassword()
		case "firstName":
			user.FirstName = &req.GetUser().FirstName
		case "lastName":
			user.LastName = &req.GetUser().LastName
		case "middleName":
			user.MiddleName = &req.GetUser().MiddleName
		case "telephone":
			user.Telephone = &req.GetUser().GetTelephone().Telephone
			updateCodes.TelephoneCode = &req.GetUser().GetTelephone().Code
		case "pseudonym":
			user.Pseudonym = &req.GetUser().Pseudonym
		}
	}

	return user, updateCodes
}

func FromGetUserRequest(req *userv1.GetUserRequest) *core.User {
	user := new(core.User)
	user.ID = int(req.GetUserId())

	return user
}

func ToGetUserResponse(user core.User) *userv1.GetUserResponse {
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

	return &userv1.GetUserResponse{
		UserId:     int64(user.ID),
		Username:   user.Username,
		Email:      email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: middleName,
		Telephone:  telephone,
		Pseudonym:  user.Pseudonym,
		IsDeleted:  user.IsDeleted,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
	}
}

func ToUpdateUserResponse(user core.User) *userv1.UpdateUserResponse {
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

	return &userv1.UpdateUserResponse{
		UserId:     int64(user.ID),
		Username:   user.Username,
		Email:      email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: middleName,
		Telephone:  telephone,
		Pseudonym:  user.Pseudonym,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
	}
}
