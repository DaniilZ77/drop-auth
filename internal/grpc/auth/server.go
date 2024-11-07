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
	authv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	authv1.UnimplementedAuthServiceServer
	authService core.AuthService
	authConfig  core.AuthConfig
	mailService core.MailService
	userService core.UserService
	smsService  core.SMSService
}

func Register(
	gRPCServer *grpc.Server,
	authService core.AuthService,
	authConfig core.AuthConfig,
	mailService core.MailService,
	userService core.UserService,
	smsService core.SMSService,
) {
	authv1.RegisterAuthServiceServer(gRPCServer, &server{authService: authService, authConfig: authConfig, mailService: mailService, userService: userService, smsService: smsService})
}

func (s *server) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := s.authService.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrAlreadyDeleted) ||
			errors.Is(err, core.ErrRefreshTokenNotValid) ||
			errors.Is(err, core.ErrEmailAndTelephoneNotVerified) {
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
		return nil, helper.ToGRPCError(v)
	}

	user := auth.FromLoginRequest(req)

	accessToken, refreshToken, err := s.authService.Login(ctx, *user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrInvalidCredentials) ||
			errors.Is(err, core.ErrUserNotFound) ||
			errors.Is(err, core.ErrAlreadyDeleted) ||
			errors.Is(err, core.ErrEmailAndTelephoneNotVerified) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
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
		return nil, helper.ToGRPCError(v)
	}

	user := auth.FromSignupRequest(req)

	retUser, err := s.authService.Signup(ctx, *user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrEmailAlreadyExists) ||
			errors.Is(err, core.ErrUsernameAlreadyExists) ||
			errors.Is(err, core.ErrTelephoneAlreadyExists) ||
			errors.Is(err, core.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	go func() {
		ctx := context.Background()
		if err = s.mailService.Send(ctx, *retUser); err != nil && !errors.Is(err, core.ErrEmailNotProvided) {
			logger.Log().Error(ctx, err.Error())
		}
		if err = s.smsService.Send(ctx, *retUser); err != nil && !errors.Is(err, core.ErrTelephoneNotProvided) {
			logger.Log().Error(ctx, err.Error())
		}
	}()

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

func (s *server) VerifyEmail(ctx context.Context, req *authv1.VerifyEmailRequest) (*authv1.VerifyEmailResponse, error) {
	_, err := s.mailService.Verify(ctx, req.GetCode())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrVerificationCodeNotValid) {
			return nil, status.Error(codes.InvalidArgument, core.ErrVerificationCodeNotValid.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return &authv1.VerifyEmailResponse{}, nil
}

func (s *server) SendEmail(ctx context.Context, req *authv1.SendEmailRequest) (*authv1.SendEmailResponse, error) {
	v := validator.New()
	model.ValidateSendEmailRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.ToGRPCError(v)
	}

	go func() {
		ctx := context.Background()
		if err := s.mailService.Resend(ctx, req.GetEmail()); err != nil {
			logger.Log().Error(ctx, err.Error())
		}
	}()

	return &authv1.SendEmailResponse{}, nil
}

func (s *server) SendTelephone(ctx context.Context, req *authv1.SendTelephoneRequest) (*authv1.SendTelephoneResponse, error) {
	v := validator.New()
	model.ValidateSendTelephonelRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.ToGRPCError(v)
	}

	go func() {
		ctx := context.Background()
		if err := s.smsService.Resend(ctx, req.GetTelephone()); err != nil {
			logger.Log().Error(ctx, err.Error())
		}
	}()

	return &authv1.SendTelephoneResponse{}, nil
}

func (s *server) VerifyTelephone(ctx context.Context, req *authv1.VerifyTelephoneRequest) (*authv1.VerifyTelephoneResponse, error) {
	_, err := s.smsService.Verify(ctx, req.GetCode())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrVerificationCodeNotValid) {
			return nil, status.Error(codes.InvalidArgument, core.ErrVerificationCodeNotValid.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return &authv1.VerifyTelephoneResponse{}, nil
}
