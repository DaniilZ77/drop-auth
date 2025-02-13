package grpc

import (
	"context"
	"errors"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/bufbuild/protovalidate-go"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserSaver interface {
	UpdateUser(ctx context.Context, user generated.UpdateUserParams) (*generated.User, error)
}

type UserProvider interface {
	GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *int, err error)
	GetUser(ctx context.Context, id string) (*generated.User, error)
}

type AuthProvider interface {
	Login(ctx context.Context, user generated.SaveUserParams) (accessToken, refreshToken *string, err error)
	RefreshToken(ctx context.Context, token string) (accessToken, refreshToken *string, err error)
}

type server struct {
	userv1.UnimplementedUserServiceServer
	userSaver    UserSaver
	userProvider UserProvider
	authProvider AuthProvider
}

func Register(
	gRPCServer *grpc.Server,
	userSaver UserSaver,
	userProvider UserProvider,
	authProvider AuthProvider,
) {
	userv1.RegisterUserServiceServer(gRPCServer, &server{userSaver: userSaver, userProvider: userProvider, authProvider: authProvider})
}

func (s *server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	updateUser := model.ToModelUpdateUserParams(uuid.MustParse(*id), req)
	user, err := s.userSaver.UpdateUser(ctx, *updateUser)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToUpdateUserResponse(user), nil
}

func (s *server) GetUsers(ctx context.Context, req *userv1.GetUsersRequest) (*userv1.GetUsersResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := model.ToModelGetUsersParams(req)
	users, total, err := s.userProvider.GetUsers(ctx, *params)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToGetUsersResponse(users, *total, params), nil
}

func (s *server) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := s.authProvider.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrRefreshTokenNotValid) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.RefreshTokenResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (s *server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := getInitDataFromContext(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if req.Pseudonym != nil {
		user.Pseudonym = *req.Pseudonym
	}

	accessToken, refreshToken, err := s.authProvider.Login(ctx, *user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.LoginResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (s *server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.userProvider.GetUser(ctx, req.UserId)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.GetUserResponse{
		UserId:    user.ID.String(),
		Username:  user.Username,
		Pseudonym: user.Pseudonym,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
	}, nil
}
