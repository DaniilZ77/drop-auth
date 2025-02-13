package model

import (
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	OrderBy struct {
		Order string
		Field string
	}

	GetUsersParams struct {
		UserID    *string
		Username  *string
		Pseudonym *string
		FirstName *string
		LastName  *string
		OrderBy   *OrderBy
		Limit     int
		Offset    int
	}

	AuthConfig struct {
		Secret          string
		AccessTokenTTL  int
		RefreshTokenTTL int
	}
)

func ToModelUpdateUserParams(id uuid.UUID, user *userv1.UpdateUserRequest) *generated.UpdateUserParams {
	return &generated.UpdateUserParams{
		ID:        id,
		Pseudonym: user.Pseudonym,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func ToUpdateUserResponse(user *generated.User) *userv1.UpdateUserResponse {
	return &userv1.UpdateUserResponse{
		UserId:    user.ID.String(),
		Username:  user.Username,
		Pseudonym: user.Pseudonym,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func ToModelGetUsersParams(params *userv1.GetUsersRequest) *GetUsersParams {
	var orderBy *OrderBy
	if params.OrderBy != nil {
		orderBy = &OrderBy{
			Order: params.OrderBy.Order.String(),
			Field: params.OrderBy.Field,
		}
	}

	return &GetUsersParams{
		UserID:    params.UserId,
		Username:  params.Username,
		Pseudonym: params.Pseudonym,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		OrderBy:   orderBy,
		Limit:     int(params.Limit),
		Offset:    int(params.Offset),
	}
}

func ToGetUsersResponse(users []generated.User, total int, params *GetUsersParams) *userv1.GetUsersResponse {
	var response userv1.GetUsersResponse
	for _, user := range users {
		response.Users = append(response.Users, &userv1.User{
			UserId:    user.ID.String(),
			Username:  user.Username,
			Pseudonym: user.Pseudonym,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: timestamppb.New(user.CreatedAt.Time),
		})
	}
	response.Pagination = &userv1.Pagination{
		Records:        int64(total),
		RecordsPerPage: int64(params.Limit),
		Pages:          (int64(total) + int64(params.Limit) - 1) / int64(params.Limit),
		CurPage:        int64(params.Offset)/int64(params.Limit) + 1,
	}
	return &response
}
