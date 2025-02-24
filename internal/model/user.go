package model

import (
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		Limit     uint64
		Offset    uint64
	}

	AuthConfig struct {
		Secret          string
		AccessTokenTTL  int
		RefreshTokenTTL int
	}

	Admin struct {
		ID        uuid.UUID
		Username  string
		Scale     generated.AdminScale
		CreatedAt time.Time
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
			Order: params.OrderBy.Order,
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
		Limit:     params.Limit,
		Offset:    params.Offset,
	}
}

func ToGetUsersResponse(users []generated.User, total uint64, params *GetUsersParams) *userv1.GetUsersResponse {
	var resp userv1.GetUsersResponse
	for _, user := range users {
		resp.Users = append(resp.Users, &userv1.User{
			UserId:    user.ID.String(),
			Username:  user.Username,
			Pseudonym: user.Pseudonym,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: timestamppb.New(user.CreatedAt.Time),
		})
	}
	resp.Pagination = &userv1.Pagination{
		Records:        total,
		RecordsPerPage: params.Limit,
		Pages:          (total + params.Limit - 1) / params.Limit,
		CurPage:        params.Offset/params.Limit + 1,
	}
	return &resp
}

func ToModelGetAdminsParams(params *userv1.GetAdminsRequest) (*generated.GetAdminsParams, error) {
	var userID pgtype.UUID
	if params.UserId != nil {
		if err := userID.Scan(*params.UserId); err != nil {
			return nil, err
		}
	}

	var adminScale generated.NullAdminScale
	if params.AdminScale != nil {
		if err := adminScale.Scan(*params.AdminScale); err != nil {
			return nil, err
		}
	}

	return &generated.GetAdminsParams{
		UserID:     userID,
		Username:   params.Username,
		AdminScale: adminScale,
		Limit:      int32(params.Limit),
		Offset:     int32(params.Offset),
	}, nil
}

func ToGetAdminsResponse(admins []generated.GetAdminsRow, total uint64, params *userv1.GetAdminsRequest) *userv1.GetAdminsResponse {
	var resp userv1.GetAdminsResponse

	for _, v := range admins {
		resp.Admins = append(resp.Admins, &userv1.Admin{
			UserId:     v.ID.String(),
			Username:   v.Username,
			AdminScale: string(v.Scale),
			CreatedAt:  timestamppb.New(v.CreatedAt.Time),
		})
	}

	resp.Pagination = &userv1.Pagination{
		Records:        total,
		RecordsPerPage: params.Limit,
		Pages:          (total + params.Limit - 1) / params.Limit,
		CurPage:        params.Offset/params.Limit + 1,
	}
	return &resp
}
