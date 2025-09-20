package limiter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"workshop/internal/service/workshop"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	redisClient *redis.Client
	keyPrefix   string

	mu     sync.RWMutex
	groups map[string]workshop.LimitConfig
}

func NewRedisLimiter(redisClient *redis.Client, keyPrefix string) *RedisLimiter {
	return &RedisLimiter{
		redisClient: redisClient,
		keyPrefix:   keyPrefix,
		groups:      make(map[string]workshop.LimitConfig),
	}
}

func (r *RedisLimiter) RegisterGroup(groupKey string, cfg workshop.LimitConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[groupKey] = cfg
}

func (r *RedisLimiter) getGroupConfig(groupKey string) workshop.LimitConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg := r.groups[groupKey]
	return cfg
}

func (r *RedisLimiter) Check(ctx context.Context, groupKey, userID string) (workshop.RateLimitInfo, error) {
	cfg := r.getGroupConfig(groupKey)
	if cfg.Limit == 0 {
		return workshop.RateLimitInfo{
			Limit:     cfg.Limit,
			Current:   0,
			Remaining: 1,
			ResetAt:   time.Now().Add(cfg.Period),
		}, nil
	}

	counterKey := r.buildKey(groupKey, userID)

	pipe := r.redisClient.Pipeline()
	getCmd := pipe.Get(ctx, counterKey)
	ttlCmd := pipe.TTL(ctx, counterKey)

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return workshop.RateLimitInfo{}, fmt.Errorf("redis pipeline exec error: %w", err)
	}

	current, err := getCmd.Uint64()
	if errors.Is(err, redis.Nil) {
		current = 0
	} else if err != nil {
		return workshop.RateLimitInfo{}, fmt.Errorf("failed GET in pipeline: %w", err)
	}

	ttl, err := ttlCmd.Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return workshop.RateLimitInfo{}, fmt.Errorf("failed TTL in pipeline: %w", err)
	}

	if ttl <= 0 {
		ttl = cfg.Period
	}
	resetAt := time.Now().Add(ttl)

	var remaining uint64
	if current >= cfg.Limit {
		remaining = 0
	} else {
		remaining = cfg.Limit - current
	}

	return workshop.RateLimitInfo{
		Limit:     cfg.Limit,
		Current:   current,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}

func (r *RedisLimiter) TriggerIncrease(ctx context.Context, groupKey, userID string) error {
	cfg := r.getGroupConfig(groupKey)
	if cfg.Limit == 0 {
		return nil
	}

	counterKey := r.buildKey(groupKey, userID)

	pipe := r.redisClient.Pipeline()
	incrCmd := pipe.Incr(ctx, counterKey)
	ttlCmd := pipe.TTL(ctx, counterKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec error: %w", err)
	}

	newVal := incrCmd.Val()
	ttl := ttlCmd.Val()

	if (ttl <= 0) && (newVal == 1) {
		if err := r.redisClient.Expire(ctx, counterKey, cfg.Period).Err(); err != nil {
			return fmt.Errorf("failed to set expire: %w", err)
		}
	}

	return nil
}

func (r *RedisLimiter) buildKey(groupKey, userID string) string {
	return fmt.Sprintf("%s:%s:%s", r.keyPrefix, groupKey, userID)
}
