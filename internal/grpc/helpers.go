package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
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
	adminContextKey    = contextKey("admin")
)

func AuthMiddleware(secrets map[string]string, requireAuth, requireAdmin map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !requireAuth[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("authorization")) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), "token not provided")
		}

		data := strings.Split(md.Get("authorization")[0], " ")
		if len(data) < 2 {
			return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), "not enough args in header")
		}

		switch token := strings.TrimSpace(data[1]); strings.ToLower(data[0]) {
		case "bearer":
			id, admin, err := validateToken(token, secrets["bearer"])
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), err.Error())
			}

			if requireAdmin[info.FullMethod] && generated.AdminScale(*admin) != generated.AdminScaleMinor && generated.AdminScale(*admin) != generated.AdminScaleMajor {
				return nil, status.Errorf(codes.PermissionDenied, "%s: %s", model.ErrUnauthorized, "must be admin")
			}

			ctx = context.WithValue(ctx, userIDContextKey, *id)
			ctx = context.WithValue(ctx, adminContextKey, admin)
		case "tma":
			if err := initdata.Validate(token, secrets["tma"], -1); err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), err.Error())
			}

			initData, err := initdata.Parse(token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), err.Error())
			}

			ctx = context.WithValue(ctx, initDataContextKey, initData)
		default:
			return nil, status.Errorf(codes.Unauthenticated, "%s: %s", model.ErrUnauthorized.Error(), "invalid header format")
		}

		return handler(ctx, req)
	}
}

func validateToken(token, secret string) (id, admin *string, err error) {
	data, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %s", model.ErrUnauthorized, "unexpected signing method")
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", model.ErrUnauthorized, err)
	}

	if claims, ok := data.Claims.(jwt.MapClaims); ok && data.Valid {
		id, ok := claims["id"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("%w: %s", model.ErrUnauthorized, "invalid id")
		}

		admin, ok := claims["admin"].(string)
		if !ok {
			return &id, nil, nil
		}

		return &id, &admin, nil
	}

	return nil, nil, model.ErrUnauthorized
}

func getUserIDFromContext(ctx context.Context) (*string, error) {
	id, ok := ctx.Value(userIDContextKey).(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s", model.ErrUnauthorized, "user id not provided")
	}

	return &id, nil
}

func getAdminFromContext(ctx context.Context) *string {
	admin, _ := ctx.Value(adminContextKey).(*string)
	return admin
}

func getInitDataFromContext(ctx context.Context) (*generated.SaveUserParams, error) {
	initData, ok := ctx.Value(initDataContextKey).(initdata.InitData)
	if !ok {
		return nil, fmt.Errorf("%w: %s", model.ErrUnauthorized, "init data not provided")
	}

	var user generated.SaveUserParams
	user.Username = initData.User.Username
	user.FirstName = initData.User.FirstName
	user.LastName = initData.User.LastName

	return &user, nil
}

func isLocalhost(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	return host == "127.0.0.1" || host == "::1"
}
