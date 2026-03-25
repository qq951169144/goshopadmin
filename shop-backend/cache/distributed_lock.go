package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// DistributedLock 分布式锁结构体
// 分布式锁是一种用于在分布式系统中协调多个进程对共享资源访问的机制
// 使用Redis的SETNX命令实现，具有以下特点：
// 1. 互斥性：同一时间只有一个进程可以获取锁
// 2. 安全性：锁只能被持有它的进程释放
// 3. 容错性：即使持有锁的进程崩溃，锁也会在一定时间后自动释放
type DistributedLock struct {
	client *redis.Client // Redis客户端，用于执行锁操作
	key    string        // 锁的键名
	value  string        // 锁的值，用于验证锁的持有者
}

// NewDistributedLock 创建分布式锁实例
// 参数：
// - client: Redis客户端实例
// - key: 锁的键名
// - value: 锁的值，通常使用UUID等唯一标识符
// 返回：
// - *DistributedLock: 分布式锁实例
// 说明：
// 锁的value应该是唯一的，用于在释放锁时验证锁的持有者
func NewDistributedLock(client *redis.Client, key string, value string) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		value:  value,
	}
}

// NewDistributedLockFromCacheUtil 从CacheUtil创建分布式锁实例
// 参数：
// - cacheUtil: 缓存工具实例
// - key: 锁的键名
// - value: 锁的值
// 返回：
// - *DistributedLock: 分布式锁实例
// 说明：
// 便捷方法，从现有的CacheUtil创建分布式锁
func NewDistributedLockFromCacheUtil(cacheUtil *CacheUtil, key string, value string) *DistributedLock {
	return &DistributedLock{
		client: cacheUtil.Redis,
		key:    key,
		value:  value,
	}
}

// Acquire 获取分布式锁
// 参数：
// - ctx: 上下文，用于控制操作超时
// - expiration: 锁的过期时间
// 返回：
// - bool: true表示获取锁成功，false表示获取锁失败
// - error: 操作错误信息
// 说明：
// 使用Redis的SETNX命令实现，当键不存在时设置值并返回true
func (dl *DistributedLock) Acquire(ctx context.Context, expiration time.Duration) (bool, error) {
	// 使用SETNX命令尝试获取锁
	// SETNX命令在键不存在时设置值并返回1，否则返回0
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("acquire distributed lock error: %w", err)
	}
	return result, nil
}

// Release 释放分布式锁
// 参数：
// - ctx: 上下文，用于控制操作超时
// 返回：
// - error: 操作错误信息
// 说明：
// 使用Lua脚本保证原子性操作，只有锁的持有者才能释放锁
func (dl *DistributedLock) Release(ctx context.Context) error {
	// 使用 Lua 脚本保证原子性操作
	// 脚本逻辑：先检查锁的值是否与当前值匹配，匹配则删除锁，否则返回0
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
// 参数：
// - ctx: 上下文，用于控制操作超时
// - expiration: 锁的过期时间
// - timeout: 尝试获取锁的超时时间
// 返回：
// - bool: true表示获取锁成功，false表示获取锁失败
// - error: 操作错误信息
// 说明：
// 在指定的超时时间内不断尝试获取锁，直到成功或超时
func (dl *DistributedLock) TryAcquireWithTimeout(ctx context.Context, expiration time.Duration, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout) // 计算截止时间
	for time.Now().Before(deadline) { // 循环尝试获取锁
		success, err := dl.Acquire(ctx, expiration)
		if err != nil {
			return false, err
		}
		if success {
			return true, nil // 获取锁成功
		}
		time.Sleep(100 * time.Millisecond) // 短暂休眠后重试
	}
	return false, nil // 超时，获取锁失败
}
