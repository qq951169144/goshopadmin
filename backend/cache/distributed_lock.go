package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// DistributedLock 分布式锁结构体
type DistributedLock struct {
	client *redis.Client
	key    string
	value  string
}

// NewDistributedLock 创建分布式锁实例
func NewDistributedLock(client *redis.Client, key string, value string) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		value:  value,
	}
}

// NewDistributedLockFromCacheUtil 从CacheUtil创建分布式锁实例
func NewDistributedLockFromCacheUtil(cacheUtil *CacheUtil, key string, value string) *DistributedLock {
	return &DistributedLock{
		client: cacheUtil.Redis,
		key:    key,
		value:  value,
	}
}

// Acquire 获取分布式锁
func (dl *DistributedLock) Acquire(ctx context.Context, expiration time.Duration) (bool, error) {
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("acquire distributed lock error: %w", err)
	}
	return result, nil
}

// Release 释放分布式锁
func (dl *DistributedLock) Release(ctx context.Context) error {
	// 使用 Lua 脚本保证原子性操作
	luaScript := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
	_, err := dl.client.Eval(ctx, luaScript, []string{dl.key}, dl.value).Result()
	if err != nil {
		return fmt.Errorf("release distributed lock error: %w", err)
	}
	return nil
}

// TryAcquireWithTimeout 尝试获取分布式锁，带超时时间
func (dl *DistributedLock) TryAcquireWithTimeout(ctx context.Context, expiration time.Duration, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		success, err := dl.Acquire(ctx, expiration)
		if err != nil {
			return false, err
		}
		if success {
			return true, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false, nil
}