package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

// MultiLevelCache 多级缓存结构体
type MultiLevelCache struct {
	localCache  *cache.Cache
	redisClient *redis.Client
}

// NewMultiLevelCache 创建多级缓存实例
func NewMultiLevelCache(defaultExpiration, cleanupInterval time.Duration, redisClient *redis.Client) *MultiLevelCache {
	return &MultiLevelCache{
		localCache:  cache.New(defaultExpiration, cleanupInterval),
		redisClient: redisClient,
	}
}

// Get 从缓存中获取数据
func (mc *MultiLevelCache) Get(ctx context.Context, key string, dest interface{}) error {
	// 先从本地缓存获取
	if val, found := mc.localCache.Get(key); found {
		if err := json.Unmarshal([]byte(val.(string)), dest); err != nil {
			return fmt.Errorf("unmarshal local cache data error: %w", err)
		}
		return nil
	}

	// 本地缓存未命中，从 Redis 获取
	val, err := mc.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss")
	} else if err != nil {
		return fmt.Errorf("get from redis error: %w", err)
	}

	// 将 Redis 数据同步到本地缓存
	mc.localCache.Set(key, val, cache.DefaultExpiration)

	// 反序列化数据
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("unmarshal redis data error: %w", err)
	}

	return nil
}

// Set 设置缓存数据
func (mc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal data error: %w", err)
	}

	// 设置本地缓存
	mc.localCache.Set(key, string(data), expiration)

	// 设置 Redis 缓存
	if err := mc.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("set to redis error: %w", err)
	}

	return nil
}

// Delete 删除缓存数据
func (mc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	// 删除本地缓存
	mc.localCache.Delete(key)

	// 删除 Redis 缓存
	if err := mc.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete from redis error: %w", err)
	}

	return nil
}

// DeletePattern 按模式删除缓存数据
func (mc *MultiLevelCache) DeletePattern(ctx context.Context, pattern string) error {
	// 删除 Redis 缓存
	keys, err := mc.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("get keys by pattern error: %w", err)
	}

	if len(keys) > 0 {
		if err := mc.redisClient.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("delete keys from redis error: %w", err)
		}
	}

	// 本地缓存暂时不支持按模式删除，依赖过期时间自动清理

	return nil
}

// Clear 清空所有缓存
func (mc *MultiLevelCache) Clear(ctx context.Context) error {
	// 清空本地缓存
	mc.localCache.Flush()

	// 清空 Redis 缓存（生产环境慎用）
	if err := mc.redisClient.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flush redis error: %w", err)
	}

	return nil
}