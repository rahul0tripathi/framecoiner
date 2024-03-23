package integrations

import "context"

type Storage interface {
	Read(ctx context.Context, key string) (string, error)
}
