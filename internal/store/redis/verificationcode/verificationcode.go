package verificationcode

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/redis"
	rdblib "github.com/redis/go-redis/v9"
)

type storage struct {
	*redis.Redis
}

func New(rdb *redis.Redis) core.VerificationStore {
	return &storage{rdb}
}

func (s *storage) DeleteVerificationCode(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.Client.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

func (s *storage) GetVerificationCode(ctx context.Context, key string) (*core.VerificationCode, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := s.Client.Get(ctx, key).Bytes()
	if err == rdblib.Nil {
		return nil, core.ErrVerificationCodeNotValid
	} else if err != nil {
		return nil, err
	}

	var val core.VerificationCode
	err = json.Unmarshal(res, &val)
	if err != nil {
		return nil, err
	}

	return &val, nil
}

func (s *storage) SetVerificationCode(ctx context.Context, key string, val core.VerificationCode, expiresIn time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	valJSON, err := json.Marshal(val)
	if err != nil {
		return err
	}

	if err := s.Client.Set(ctx, key, valJSON, expiresIn).Err(); err != nil {
		return err
	}

	return nil
}
