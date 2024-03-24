package repo

import (
	"context"
	"time"
)

type Storage interface {
	Read(ctx context.Context, key string) (string, error)
	Write(ctx context.Context, key string, data string, expiration time.Duration) error
}
