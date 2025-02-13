package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	"github.com/golang-jwt/jwt"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	initDataContextKey = contextKey("init-data")
	userIDContextKey   = contextKey("user-id")
)

func AuthMiddleware(secrets map[string]string, requireAuth map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !requireAuth[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("authorization")) == 0 {
			logger.Log().Debug(ctx, "token not provided")
			return nil, status.Error(codes.Unauthenticated, model.ErrUnauthorized.Error())
		}

		data := strings.Split(md.Get("authorization")[0], " ")
		if len(data) < 2 {
			logger.Log().Debug(ctx, "not enough args: %v", data)
			return nil, status.Error(codes.Unauthenticated, model.ErrUnauthorized.Error())
		}

		switch token := strings.TrimSpace(data[1]); data[0] {
		case "bearer":
			id, err := validateToken(ctx, token, secrets["bearer"])
			if err != nil {
				logger.Log().Debug(ctx, err.Error())
				return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("%s: %s", model.ErrUnauthorized.Error(), err.Error()))
			}

			ctx = context.WithValue(ctx, userIDContextKey, *id)
		case "tma":
			if err := initdata.Validate(token, secrets["tma"], -1); err != nil {
				logger.Log().Debug(ctx, err.Error())
				return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("%s: %s", model.ErrUnauthorized.Error(), err.Error()))
			}

			initData, err := initdata.Parse(token)
			if err != nil {
				logger.Log().Error(ctx, err.Error())
				return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("%s: %s", model.ErrUnauthorized.Error(), err.Error()))
			}

			ctx = context.WithValue(ctx, initDataContextKey, initData)
		}

		return handler(ctx, req)
	}
}

func validateToken(ctx context.Context, token, secret string) (*string, error) {
	data, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Log().Error(ctx, "unexpected signing method")
			return nil, model.ErrUnauthorized
		}

		return []byte(secret), nil
	})
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, fmt.Errorf("%w: %w", model.ErrUnauthorized, err)
	}

	if claims, ok := data.Claims.(jwt.MapClaims); ok && data.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return nil, model.ErrUnauthorized
		}

		return &id, nil
	}

	return nil, model.ErrUnauthorized
}

func getUserIDFromContext(ctx context.Context) (*string, error) {
	id, ok := ctx.Value(userIDContextKey).(string)
	if !ok {
		logger.Log().Debug(ctx, "user id not provided")
		return nil, model.ErrUnauthorized
	}

	return &id, nil
}

func getInitDataFromContext(ctx context.Context) (*generated.SaveUserParams, error) {
	initData, ok := ctx.Value(initDataContextKey).(initdata.InitData)
	if !ok {
		logger.Log().Debug(ctx, "init data not provided")
		return nil, model.ErrUnauthorized
	}

	var user generated.SaveUserParams
	user.ExternalID = int32(initData.User.ID)
	user.Username = initData.User.Username
	user.FirstName = initData.User.FirstName
	user.LastName = initData.User.LastName

	return &user, nil
}
