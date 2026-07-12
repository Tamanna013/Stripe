package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBucketLimiter implements a distributed token bucket rate limiter via Redis.
type TokenBucketLimiter struct {
	client *redis.Client
	script *redis.Script
}

func NewTokenBucketLimiter(client *redis.Client) *TokenBucketLimiter {
	// Lua script from Part 9.3 for atomic check-and-decrement
	script := redis.NewScript(`
		local bucket = redis.call('HMGET', KEYS[1], 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or tonumber(ARGV[2])
		local last_refill = tonumber(bucket[2]) or tonumber(ARGV[3])
		local elapsed = tonumber(ARGV[3]) - last_refill
		tokens = math.min(tonumber(ARGV[2]), tokens + elapsed * tonumber(ARGV[1]))
		if tokens < 1 then
			redis.call('HMSET', KEYS[1], 'tokens', tokens, 'last_refill', ARGV[3])
			return 0
		end
		redis.call('HMSET', KEYS[1], 'tokens', tokens - 1, 'last_refill', ARGV[3])
		return 1
	`)
	return &TokenBucketLimiter{
		client: client,
		script: script,
	}
}

// Allow checks if a request is allowed for a given source ID.
func (l *TokenBucketLimiter) Allow(ctx context.Context, sourceID string, refillRate int, capacity int) (bool, error) {
	key := "rl:" + sourceID
	now := time.Now().Unix()

	res, err := l.script.Run(ctx, l.client, []string{key}, refillRate, capacity, now).Result()
	if err != nil {
		// Fail-open strategy: if Redis is down, we allow the request (logged outside).
		return true, err
	}

	return res.(int64) == 1, nil
}
