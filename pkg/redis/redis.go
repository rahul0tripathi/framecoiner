package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/rahul0tripathi/framecoiner/entity"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	UserName string
	Password string
}
type Redis struct {
	client *redis.Client
}

func NewRedisDB(cfg RedisConfig) (*Redis, error) {
	fmt.Println(cfg)
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Username: cfg.UserName,
			Password: cfg.Password,
			Addr:     cfg.Addr,
		}),
	}, nil
}

func (r *Redis) Read(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", entity.ErrEmpty
	case err != nil:
		return "", err
	case value == "":
		return "", entity.ErrEmpty
	}

	return value, nil
}
