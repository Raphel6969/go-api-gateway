package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (body []byte, statusCode int, found bool)
	Set(ctx context.Context, key string, statusCode int, body []byte, ttl time.Duration) error
}
