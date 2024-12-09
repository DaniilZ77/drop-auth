package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	helper "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/grpc"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model/auth"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model/validator"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/service/verification"
	authv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type server struct {
	authv1.UnimplementedAuthServiceServer
	authService         core.AuthService
	authConfig          core.AuthConfig
	userService         core.UserService
	verificationService core.VerificationService
}

func Register(
	gRPCServer *grpc.Server,
	authService core.AuthService,
	authConfig core.AuthConfig,
	userService core.UserService,
	verificationService core.VerificationService,
) {
	authv1.RegisterAuthServiceServer(
		gRPCServer,
		&server{
			authService:         authService,
			authConfig:          authConfig,
			userService:         userService,
			verificationService: verificationService,
		})
}

func (s *server) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := s.authService.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if helper.OneOf(err, core.ErrAlreadyDeleted, core.ErrRefreshTokenNotValid) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return auth.ToRefreshTokenResponse(*accessToken, *refreshToken), nil
}

func (s *server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	v := validator.New()
	model.ValidateLoginRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	user := auth.FromLoginRequest(req)

	accessToken, refreshToken, err := s.authService.Login(ctx, *user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if helper.OneOf(err, core.ErrInvalidCredentials, core.ErrUserNotFound, core.ErrAlreadyExists, core.ErrAlreadyDeleted) {
			return nil, status.Error(codes.Unauthenticated, core.ErrInvalidCredentials.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return auth.ToLoginResponse(*accessToken, *refreshToken), nil
}

func (s *server) Signup(ctx context.Context, req *authv1.SignupRequest) (*authv1.SignupResponse, error) {
	v := validator.New()
	model.ValidateSignupRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Log().Error(ctx, core.ErrInternal.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	emailCode, telephoneCode, user := auth.FromSignupRequest(req)

	retUser, err := s.authService.Signup(ctx, emailCode, telephoneCode, *user, p.Addr.String())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if helper.OneOf(err, core.ErrEmailAlreadyExists, core.ErrUsernameAlreadyExists, core.ErrTelephoneAlreadyExists, core.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else if errors.Is(err, core.ErrVerificationCodeNotValid) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return auth.ToSignupResponse(*retUser), nil
}

func (s *server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, err := helper.ValidToken(ctx, req.GetToken(), s.authConfig.Secret)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		if errors.Is(err, core.ErrUnauthorized) {
			return auth.ToValidateTokenResponse(false, 0), nil
		}
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return auth.ToValidateTokenResponse(true, *userID), nil
}

func (s *server) SendEmail(ctx context.Context, req *authv1.SendEmailRequest) (*authv1.SendEmailResponse, error) {
	v := validator.New()
	model.ValidateSendEmailRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Log().Error(ctx, core.ErrInternal.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	go func() {
		ctx := context.Background()

		opt := verification.WithUser(&core.User{Email: &req.Email})
		if req.GetIsVerified() {
			opt = verification.WithEmail(req.GetEmail())
		}

		if err := s.verificationService.SendEmail(ctx, opt, p.Addr.String()); err != nil {
			logger.Log().Error(ctx, err.Error())
		}
	}()

	return &authv1.SendEmailResponse{}, nil
}

func (s *server) SendSMS(ctx context.Context, req *authv1.SendSMSRequest) (*authv1.SendSMSResponse, error) {
	v := validator.New()
	model.ValidateSendSMSRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Log().Error(ctx, core.ErrInternal.Error())
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	go func() {
		ctx := context.Background()

		opt := verification.WithUser(&core.User{Telephone: &req.Telephone})
		if req.GetIsVerified() {
			opt = verification.WithTelephone(req.GetTelephone())
		}

		if err := s.verificationService.SendSMS(ctx, opt, p.Addr.String()); err != nil {
			logger.Log().Error(ctx, err.Error())
		}
	}()

	return &authv1.SendSMSResponse{}, nil
}

func (s *server) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	v := validator.New()
	model.ValidateResetPassword(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	user, err := s.authService.ResetPassword(ctx, req.GetCode(), req.GetPassword())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrVerificationCodeNotValid) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	logger.Log().Debug(ctx, "user: %v", user)

	return auth.ToResetPasswordResponse(*user), nil
}

func (s *server) LoginTelegram(ctx context.Context, req *authv1.LoginTelegramRequest) (*authv1.LoginResponse, error) {
	v := validator.New()
	model.ValidateLoginTelegramRequest(v, req)

	user, externalUser, err := helper.GetInitDataFromContext(ctx)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, status.Error(codes.Unauthenticated, core.ErrUnauthorized.Error())
	}
	auth.FromLoginTelegramRequest(req, user)

	accessToken, refreshToken, err := s.authService.LoginExternal(ctx, *user, *externalUser, core.TelegramAuthProvider, v.Valid())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrValidationFailed) {
			return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
		} else if helper.OneOf(err, core.ErrUserIDAuthProviderAlreadyExists, core.ErrUsernameAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return auth.ToLoginResponse(*accessToken, *refreshToken), nil
}
