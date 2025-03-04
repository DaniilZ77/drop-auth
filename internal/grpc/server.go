package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/domain/model"
	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/bufbuild/protovalidate-go"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserModifier interface {
	UpdateUser(ctx context.Context, user generated.UpdateUserParams) (*generated.User, error)
	AddAdmin(ctx context.Context, username string, scale generated.AdminScale) (*model.Admin, error)
	DeleteAdmin(ctx context.Context, id uuid.UUID, scale generated.AdminScale) error
	InitAdmin(ctx context.Context, username string) (*model.Admin, error)
}

type UserProvider interface {
	GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *uint64, err error)
	GetUser(ctx context.Context, id uuid.UUID) (*generated.User, error)
	GetAdmins(ctx context.Context, params generated.GetAdminsParams) (admins []generated.GetAdminsRow, total *uint64, err error)
}

type AuthProvider interface {
	Login(ctx context.Context, user generated.SaveUserParams) (accessToken, refreshToken *string, err error)
	RefreshToken(ctx context.Context, token string) (accessToken, refreshToken *string, err error)
}

type server struct {
	userv1.UnimplementedUserServiceServer
	userModifier UserModifier
	userProvider UserProvider
	authProvider AuthProvider
	log          *slog.Logger
}

func Register(
	gRPCServer *grpc.Server,
	userModifier UserModifier,
	userProvider UserProvider,
	authProvider AuthProvider,
	log *slog.Logger,
) {
	userv1.RegisterUserServiceServer(gRPCServer, &server{userModifier: userModifier, userProvider: userProvider, authProvider: authProvider, log: log})
}

func (s *server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	updateUser := model.ToDomainUpdateUserParams(*id, req)
	user, err := s.userModifier.UpdateUser(ctx, *updateUser)
	if err != nil {
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToUpdateUserResponse(user), nil
}

func (s *server) GetUsers(ctx context.Context, req *userv1.GetUsersRequest) (*userv1.GetUsersResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params := model.ToDomainGetUsersParams(req)
	users, total, err := s.userProvider.GetUsers(ctx, *params)
	if err != nil {
		if errors.Is(err, model.ErrOrderByInvalidField) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToGetUsersResponse(users, *total, params), nil
}

func (s *server) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := s.authProvider.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, model.ErrRefreshTokenNotValid) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.RefreshTokenResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (s *server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := getInitDataFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if req.Pseudonym != nil {
		user.Pseudonym = *req.Pseudonym
	}

	accessToken, refreshToken, err := s.authProvider.Login(ctx, *user)
	if err != nil {
		if errors.Is(err, model.ErrEmptyPseudonym) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.LoginResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (s *server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id must be uuid")
	}

	user, err := s.userProvider.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
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

func (s *server) AddAdmin(ctx context.Context, req *userv1.AddAdminRequest) (*userv1.AddAdminResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	admin := getAdminFromContext(ctx)
	if admin == nil {
		return nil, status.Error(codes.Unauthenticated, "must be admin")
	}

	res, err := s.userModifier.AddAdmin(ctx, req.Username, generated.AdminScale(*admin))
	if err != nil {
		if errors.Is(err, model.ErrAdminAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else if errors.Is(err, model.ErrAdminNotMajor) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		} else if errors.Is(err, model.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.AddAdminResponse{
		UserId:     res.ID.String(),
		Username:   res.Username,
		AdminScale: string(res.Scale),
		CreatedAt:  timestamppb.New(res.CreatedAt),
	}, nil
}

func (s *server) DeleteAdmin(ctx context.Context, req *userv1.DeleteAdminRequest) (*userv1.DeleteAdminResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id must be uuid")
	}

	admin := getAdminFromContext(ctx)
	if admin == nil {
		return nil, status.Error(codes.Unauthenticated, "must be admin")
	}

	if err := s.userModifier.DeleteAdmin(ctx, userID, generated.AdminScale(*admin)); err != nil && !errors.Is(err, model.ErrUserNotFound) {
		if errors.Is(err, model.ErrAdminNotMajor) || errors.Is(err, model.ErrCannotDeleteMajorAdmin) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.DeleteAdminResponse{}, nil
}

func (s *server) InitAdmin(ctx context.Context, req *userv1.InitAdminRequest) (*userv1.InitAdminResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		s.log.Error("internal error", slog.String("error", "could not get peer info"))
		return nil, status.Error(codes.Internal, "could not get peer info")
	}

	if !isLocalhost(p.Addr.String()) {
		s.log.Debug("request must be from localhost", slog.String("from", p.Addr.String()))
		return nil, status.Error(codes.PermissionDenied, "request must be from localhost")
	}

	res, err := s.userModifier.InitAdmin(ctx, req.Username)
	if err != nil {
		if errors.Is(err, model.ErrAdminAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else if errors.Is(err, model.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.InitAdminResponse{
		UserId:     res.ID.String(),
		Username:   res.Username,
		AdminScale: string(res.Scale),
		CreatedAt:  timestamppb.New(res.CreatedAt),
	}, nil
}

func (s *server) Health(context.Context, *userv1.HealthRequest) (*userv1.HealthResponse, error) {
	return &userv1.HealthResponse{
		Message: "OK",
	}, nil
}

func (s *server) GetAdmins(ctx context.Context, req *userv1.GetAdminsRequest) (*userv1.GetAdminsResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	params, err := model.ToDomainGetAdminsParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	admins, total, err := s.userProvider.GetAdmins(ctx, *params)
	if err != nil {
		s.log.Error("internal error", sl.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ToGetAdminsResponse(admins, *total, req), nil
}
