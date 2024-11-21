package gprc

import (
	"context"
	"errors"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/golang-jwt/jwt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidToken(ctx context.Context, tokenString, secret string) (*int, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Log().Error(ctx, "unexpected signing method")
			return nil, core.ErrUnauthorized
		}

		return []byte(secret), nil
	})
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		return nil, core.ErrUnauthorized
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["id"].(float64)
		if !ok {
			return nil, core.ErrUnauthorized
		}

		idInt := int(id)
		return &idInt, nil
	}

	return nil, core.ErrUnauthorized
}

func GetUserIDFromContext(ctx context.Context) (int, error) {
	id, ok := ctx.Value(userIDContextKey).(int)
	if !ok {
		logger.Log().Debug(ctx, "user id is not provided")
		return 0, core.ErrUnauthorized
	}

	return id, nil
}

func WithDetails(code codes.Code, err error, details map[string]string) error {
	st := status.New(code, err.Error())
	var violations []*errdetails.QuotaFailure_Violation
	for k, v := range details {
		violations = append(violations, &errdetails.QuotaFailure_Violation{
			Subject:     k,
			Description: v,
		})
	}
	ds, err := st.WithDetails(&errdetails.QuotaFailure{Violations: violations})
	if err != nil {
		return st.Err()
	}
	return ds.Err()
}

func OneOf(e error, errs ...error) bool {
	for _, err := range errs {
		if errors.Is(e, err) {
			return true
		}
	}

	return false
}
