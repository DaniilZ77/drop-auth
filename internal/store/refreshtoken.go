package store

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/redis"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	redislib "github.com/redis/go-redis/v9"
)

type RefreshTokenStore struct {
	*redis.Redis
}

func NewRefreshTokenStore(r *redis.Redis) *RefreshTokenStore {
	return &RefreshTokenStore{r}
}

func (s *RefreshTokenStore) GetRefreshToken(ctx context.Context, tokenID string) (*string, error) {
	userID, err := s.Redis.Get(ctx, tokenID).Result()
	if err == redislib.Nil {
		logger.Log().Debug(ctx, "refresh token does not exists")
		return nil, model.ErrRefreshTokenNotValid
	} else if err != nil {
		logger.Log().Error(ctx, "failed to get refresh token: %w", err)
		return nil, err
	}

	return &userID, nil
}

func (s *RefreshTokenStore) SetRefreshToken(ctx context.Context, userID string, tokenID string, expiry time.Duration) error {
	if err := s.Redis.Set(ctx, tokenID, userID, expiry).Err(); err != nil {
		logger.Log().Error(ctx, "failed to set refresh token: %w", err)
		return err
	}

	return nil
}

func (s *RefreshTokenStore) ReplaceRefreshToken(ctx context.Context, oldID string, newID string, userID string, expiry time.Duration) error {
	_, err := s.Redis.TxPipelined(ctx, func(pipe redislib.Pipeliner) error {
		pipe.Del(ctx, oldID)
		pipe.Set(ctx, newID, userID, expiry)

		return nil
	})
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return err
	}

	return nil
}
