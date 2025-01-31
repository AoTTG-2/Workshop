package limiter

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workshop/internal/service/workshop"
)

func Test_NewRedisLimiter(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRedisLimiter(client, "test_prefix")
	require.NotNil(t, rl)
	assert.Equal(t, "test_prefix", rl.keyPrefix)
	assert.NotNil(t, rl.groups)
	assert.Equal(t, 0, len(rl.groups))
}

func Test_RedisLimiter_RegisterGroup(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRedisLimiter(client, "test_prefix")
	cfg := workshop.LimitConfig{
		Limit:  10,
		Period: 5 * time.Minute,
	}
	rl.RegisterGroup("posts_limit", cfg)

	require.Equal(t, 1, len(rl.groups))
	storedCfg := rl.getGroupConfig("posts_limit")
	assert.Equal(t, uint64(10), storedCfg.Limit)
	assert.Equal(t, 5*time.Minute, storedCfg.Period)
}

func Test_RedisLimiter_Check(t *testing.T) {
	t.Run("Check with non-existent key", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  3,
			Period: 2 * time.Minute,
		}
		rl.RegisterGroup("test_group", cfg)

		info, err := rl.Check(context.Background(), "test_group", "user123")
		require.NoError(t, err)
		assert.Equal(t, uint64(3), info.Limit)
		assert.Equal(t, uint64(0), info.Current)
		assert.Equal(t, uint64(3), info.Remaining)
		assert.True(t, time.Until(info.ResetAt) <= 2*time.Minute+time.Second)
	})

	t.Run("Check with existent key", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  5,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("test_group", cfg)

		err = client.Set(context.Background(), "prefix:test_group:user123", 2, 0).Err()
		require.NoError(t, err)
		err = client.Expire(context.Background(), "prefix:test_group:user123", 10*time.Minute).Err()
		require.NoError(t, err)

		info, err := rl.Check(context.Background(), "test_group", "user123")
		require.NoError(t, err)
		assert.Equal(t, uint64(5), info.Limit)
		assert.Equal(t, uint64(2), info.Current)
		assert.Equal(t, uint64(3), info.Remaining)
		assert.True(t, time.Until(info.ResetAt) <= 10*time.Minute+time.Second)
	})

	t.Run("Corrupted counter value", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  5,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("test_group", cfg)

		err = client.Set(context.Background(), "prefix:test_group:user123", "not-a-number", 0).Err()
		require.NoError(t, err)

		_, checkErr := rl.Check(context.Background(), "test_group", "user123")
		assert.ErrorIs(t, checkErr, strconv.ErrSyntax)
	})

	t.Run("Zero limit config", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  0,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("zero_group", cfg)

		info, err := rl.Check(context.Background(), "test_group", "user123")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), info.Limit)
		assert.Equal(t, uint64(0), info.Current)
		assert.Equal(t, uint64(1), info.Remaining)
	})

	t.Run("Pipeline exec error", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  2,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("test_group", cfg)

		mr.Close()

		_, checkErr := rl.Check(context.Background(), "test_group", "user123")
		assert.Error(t, checkErr)
		assert.Contains(t, checkErr.Error(), "pipeline exec error")
	})
}

func Test_RedisLimiter_TriggerIncrease(t *testing.T) {
	t.Run("Key does not exist", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  5,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("test_group", cfg)

		err = rl.TriggerIncrease(context.Background(), "test_group", "user123")
		require.NoError(t, err)

		valStr, err := client.Get(context.Background(), "prefix:test_group:user123").Result()
		require.NoError(t, err)
		val, err := strconv.Atoi(valStr)
		require.NoError(t, err)
		assert.Equal(t, 1, val)

		ttl, err := client.TTL(context.Background(), "prefix:test_group:user123").Result()
		require.NoError(t, err)
		assert.True(t, ttl > 0)
	})

	t.Run("Key exists and leaves TTL if > 0", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  5,
			Period: 10 * time.Minute,
		}
		rl.RegisterGroup("test_group", cfg)

		err = client.Set(context.Background(), "prefix:test_group:user123", 2, 5*time.Minute).Err()
		require.NoError(t, err)

		err = rl.TriggerIncrease(context.Background(), "test_group", "user123")
		require.NoError(t, err)

		valStr, err := client.Get(context.Background(), "prefix:test_group:user123").Result()
		require.NoError(t, err)
		val, err := strconv.Atoi(valStr)
		require.NoError(t, err)
		assert.Equal(t, 3, val)

		ttl, err := client.TTL(context.Background(), "prefix:test_group:user123").Result()
		require.NoError(t, err)
		assert.True(t, ttl > 4*time.Minute && ttl <= 5*time.Minute)
	})

	t.Run("Zero limit config", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  0,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("test_group", cfg)

		err = rl.TriggerIncrease(context.Background(), "test_group", "user123")
		require.NoError(t, err)
		err = rl.TriggerIncrease(context.Background(), "test_group", "user123")
		require.NoError(t, err)

		_, err = client.Get(context.Background(), "prefix:zero_group:user_zero").Result()
		assert.Error(t, err)

		info, err := rl.Check(context.Background(), "test_group", "user123")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), info.Current)
		assert.Equal(t, uint64(1), info.Remaining)
	})

	t.Run("Pipeline exec error", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		rl := NewRedisLimiter(client, "prefix")
		cfg := workshop.LimitConfig{
			Limit:  5,
			Period: 1 * time.Hour,
		}
		rl.RegisterGroup("test_group", cfg)

		mr.Close()

		err = rl.TriggerIncrease(context.Background(), "test_group", "user123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pipeline exec error")
	})
}
