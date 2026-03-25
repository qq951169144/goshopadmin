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
// 多级缓存是一种结合本地缓存和分布式缓存的缓存策略
// 特点：
// 1. 本地缓存：速度快，适合热点数据
// 2. Redis缓存：分布式，适合跨实例共享数据
// 3. 两级缓存结合，提高性能和可靠性
type MultiLevelCache struct {
	localCache   *cache.Cache  // 本地缓存，使用go-cache实现
	redisClient *redis.Client  // Redis客户端，用于分布式缓存
}

// NewMultiLevelCache 创建多级缓存实例
// 参数：
// - defaultExpiration: 本地缓存默认过期时间
// - cleanupInterval: 本地缓存清理间隔
// - redisClient: Redis客户端实例
// 返回：
// - *MultiLevelCache: 多级缓存实例
// 说明：
// 本地缓存用于提高热点数据的访问速度，Redis缓存用于跨实例数据共享
func NewMultiLevelCache(defaultExpiration, cleanupInterval time.Duration, redisClient *redis.Client) *MultiLevelCache {
	return &MultiLevelCache{
		localCache:   cache.New(defaultExpiration, cleanupInterval),
		redisClient: redisClient,
	}
}

// Get 从缓存中获取数据
// 参数：
// - ctx: 上下文，用于控制操作超时
// - key: 缓存键
// - dest: 目标对象，用于存储获取的数据
// 返回：
// - error: 操作错误信息
// 说明：
// 1. 先从本地缓存获取数据
// 2. 本地缓存未命中，从Redis获取
// 3. 将Redis数据同步到本地缓存
// 4. 反序列化数据到目标对象
func (mc *MultiLevelCache) Get(ctx context.Context, key string, dest interface{}) error {
	// 1. 先从本地缓存获取
	if val, found := mc.localCache.Get(key); found {
		// 反序列化本地缓存数据
		if err := json.Unmarshal([]byte(val.(string)), dest); err != nil {
			return fmt.Errorf("unmarshal local cache data error: %w", err)
		}
		return nil
	}

	// 2. 本地缓存未命中，从 Redis 获取
	val, err := mc.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss") // 缓存未命中
	} else if err != nil {
		return fmt.Errorf("get from redis error: %w", err) // Redis错误
	}

	// 3. 将 Redis 数据同步到本地缓存
	mc.localCache.Set(key, val, cache.DefaultExpiration)

	// 4. 反序列化数据
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("unmarshal redis data error: %w", err)
	}

	return nil
}

// Set 设置缓存数据
// 参数：
// - ctx: 上下文，用于控制操作超时
// - key: 缓存键
// - value: 缓存值
// - expiration: 过期时间
// 返回：
// - error: 操作错误信息
// 说明：
// 1. 序列化数据
// 2. 设置本地缓存
// 3. 设置Redis缓存
func (mc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 1. 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal data error: %w", err)
	}

	// 2. 设置本地缓存
	mc.localCache.Set(key, string(data), expiration)

	// 3. 设置 Redis 缓存
	if err := mc.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("set to redis error: %w", err)
	}

	return nil
}

// Delete 删除缓存数据
// 参数：
// - ctx: 上下文，用于控制操作超时
// - key: 缓存键
// 返回：
// - error: 操作错误信息
// 说明：
// 1. 删除本地缓存
// 2. 删除Redis缓存
func (mc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	// 1. 删除本地缓存
	mc.localCache.Delete(key)

	// 2. 删除 Redis 缓存
	if err := mc.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete from redis error: %w", err)
	}

	return nil
}

// DeletePattern 按模式删除缓存数据
// 参数：
// - ctx: 上下文，用于控制操作超时
// - pattern: 键模式，如 "product:list:*"
// 返回：
// - error: 操作错误信息
// 说明：
// 1. 删除Redis中匹配模式的所有键
// 2. 本地缓存暂时不支持按模式删除，依赖过期时间自动清理
func (mc *MultiLevelCache) DeletePattern(ctx context.Context, pattern string) error {
	// 1. 删除 Redis 缓存
	keys, err := mc.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("get keys by pattern error: %w", err)
	}

	if len(keys) > 0 {
		if err := mc.redisClient.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("delete keys from redis error: %w", err)
		}
	}

	// 2. 本地缓存暂时不支持按模式删除，依赖过期时间自动清理

	return nil
}

// Clear 清空所有缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// 返回：
// - error: 操作错误信息
// 说明：
// 1. 清空本地缓存
// 2. 清空Redis缓存（生产环境慎用）
func (mc *MultiLevelCache) Clear(ctx context.Context) error {
	// 1. 清空本地缓存
	mc.localCache.Flush()

	// 2. 清空 Redis 缓存（生产环境慎用）
	if err := mc.redisClient.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flush redis error: %w", err)
	}

	return nil
}
