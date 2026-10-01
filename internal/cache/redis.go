package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedResponse struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, int, bool) {
	if r == nil || r.client == nil {
		return nil, 0, false
	}
	val, err := r.client.Get(ctx, "cache:"+key).Bytes()
	if err != nil {
		return nil, 0, false
	}

	var resp CachedResponse
	if err := json.Unmarshal(val, &resp); err != nil {
		return nil, 0, false
	}

	return resp.Body, resp.StatusCode, true
}

func (r *RedisCache) Set(ctx context.Context, key string, statusCode int, body []byte, ttl time.Duration) error {
	payload, err := json.Marshal(CachedResponse{
		StatusCode: statusCode,
		Body:       body,
	})
	if err != nil {
		return err
	}

	return r.client.Set(ctx, "cache:"+key, payload, ttl).Err()
}
