package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRateLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
}

func NewRedisRateLimiter(
	client *redis.Client,
	limit int64,
	window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{client: client, limit: limit, window: window}
}

// Allow implements app.RateLimiterPort
func (r *RedisRateLimiter) Allow(
	ctx context.Context,
	key string,
) (bool, error) {
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		r.client.Expire(ctx, key, r.window)
	}
	return count <= int64(r.limit), nil
}
