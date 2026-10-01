package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var slidingWindowScript = redis.NewScript(`
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local window = tonumber(ARGV[2])
	local limit = tonumber(ARGV[3])

	local clearBefore = now - window

	-- 1. Remove timestamps outside the sliding window
	redis.call('ZREMRANGEBYSCORE', key, '-inf', clearBefore)

	-- 2. Count requests remaining in current window
	local currentRequests = redis.call('ZCARD', key)

	if currentRequests < limit then
		-- 3. Add current request timestamp (using score and member = now)
		local sequenceKey = key .. ':seq'
		redis.call('ZADD', key, now, now .. '-' .. redis.call('INCR', sequenceKey))
		-- Set key expiration slightly beyond the window for auto-cleanup
		redis.call('PEXPIRE', key, window + 1000)
		redis.call('PEXPIRE', sequenceKey, window + 1000)
		return 1
	else
		return 0
	end
`)

type RedisLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRedisLimiter(client *redis.Client, limit int, window time.Duration) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

func (rl *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	windowMillis := rl.window.Milliseconds()

	res, err := slidingWindowScript.Run(
		ctx,
		rl.client,
		[]string{"ratelimit:" + key},
		now,
		windowMillis,
		rl.limit,
	).Int()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}
