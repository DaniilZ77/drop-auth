package confirmationcode

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/redis"
	rdb "github.com/redis/go-redis/v9"
)

type storage struct {
	*redis.Redis
}

func New(rdb *redis.Redis) core.ConfirmationCodeStore {
	return &storage{rdb}
}

func (s *storage) DeleteConfirmationCode(ctx context.Context, code string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.Client.Del(ctx, code).Err(); err != nil {
		return err
	}

	return nil
}

func (s *storage) GetConfirmationCode(ctx context.Context, code string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userID, err := s.Client.Get(ctx, code).Int()
	if err == rdb.Nil {
		return 0, core.ErrConfirmationCodeNotValid
	} else if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *storage) SetConfirmationCode(ctx context.Context, userID int, code string, expiresIn time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.Client.Set(ctx, code, userID, expiresIn).Err(); err != nil {
		return err
	}

	return nil
}
