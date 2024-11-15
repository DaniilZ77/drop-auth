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
		if helper.OneOf(err, core.ErrAlreadyDeleted, core.ErrRefreshTokenNotValid, core.ErrEmailAndTelephoneNotVerified) {
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
		if helper.OneOf(err, core.ErrEmailNotVerified, core.ErrTelephoneNotVerified) {
			go func() {
				ctx := context.Background()
				sendEmailOrTelephone(ctx,
					verification.WithEmail(req.GetEmail()),
					verification.WithTelephone(req.GetTelephone()),
					s.verificationService,
				)
			}()

			var key string
			if errors.Is(err, core.ErrEmailNotVerified) {
				key = core.KeyEmailNotVerified
			} else {
				key = core.KeyTelephoneNotVerified
			}

			return nil, helper.WithDetails(
				codes.Unauthenticated,
				err,
				map[string]string{
					key: "code was sent",
				},
			)
		} else if helper.OneOf(err, core.ErrInvalidCredentials, core.ErrUserNotFound, core.ErrAlreadyExists, core.ErrAlreadyDeleted) {
			return nil, helper.WithDetails(
				codes.Unauthenticated,
				core.ErrInvalidCredentials,
				map[string]string{
					core.KeyInvalidCredentials: "invalid email or telephone or password",
				},
			)
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

	user := auth.FromSignupRequest(req)

	retUser, err := s.authService.Signup(ctx, *user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if helper.OneOf(err, core.ErrEmailAlreadyExists, core.ErrUsernameAlreadyExists, core.ErrTelephoneAlreadyExists, core.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	go func() {
		ctx := context.Background()
		sendEmailOrTelephone(ctx,
			verification.WithUser(retUser),
			verification.WithUser(retUser),
			s.verificationService,
		)
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

func (s *server) Verify(ctx context.Context, req *authv1.VerifyRequest) (*authv1.VerifyResponse, error) {
	_, err := s.verificationService.Verify(ctx, req.GetCode())
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, core.ErrVerificationCodeNotValid) {
			return nil, status.Error(codes.InvalidArgument, core.ErrVerificationCodeNotValid.Error())
		}
		return nil, status.Error(codes.Internal, core.ErrInternal.Error())
	}

	return &authv1.VerifyResponse{}, nil
}

func (s *server) SendEmail(ctx context.Context, req *authv1.SendEmailRequest) (*authv1.SendEmailResponse, error) {
	v := validator.New()
	model.ValidateSendEmailRequest(v, req)
	if !v.Valid() {
		logger.Log().Debug(ctx, fmt.Sprintf("%+v", v.Errors))
		return nil, helper.WithDetails(codes.InvalidArgument, core.ErrValidationFailed, v.Errors)
	}

	go func() {
		ctx := context.Background()
		if err := s.verificationService.SendEmail(ctx,
			verification.WithUser(&core.User{Email: &req.Email}),
		); err != nil {
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

	go func() {
		ctx := context.Background()
		if err := s.verificationService.SendSMS(ctx,
			verification.WithUser(&core.User{Telephone: &req.Telephone}),
		); err != nil {
			logger.Log().Error(ctx, err.Error())
		}
	}()

	return &authv1.SendSMSResponse{}, nil
}
