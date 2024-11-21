package verification

import (
	"context"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
)

func WithUser(user *core.User) core.Option {
	return func(context.Context, core.UserStore) (*core.User, error) {
		return user, nil
	}
}

func WithEmail(email string) core.Option {
	return func(ctx context.Context, s core.UserStore) (*core.User, error) {
		user, err := s.GetUserByEmail(ctx, email)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}
		return user, nil
	}
}

func WithTelephone(telephone string) core.Option {
	return func(ctx context.Context, s core.UserStore) (*core.User, error) {
		user, err := s.GetUserByTelephone(ctx, telephone)
		if err != nil {
			logger.Log().Error(ctx, err.Error())
			return nil, err
		}
		return user, nil
	}
}
