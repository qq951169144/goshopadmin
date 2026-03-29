package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// BloomFilter 布隆过滤器结构体
type BloomFilter struct {
	client *redis.Client
	key    string
}

// NewBloomFilter 创建布隆过滤器实例
func NewBloomFilter(client *redis.Client, key string) *BloomFilter {
	return &BloomFilter{
		client: client,
		key:    key,
	}
}

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(ctx context.Context, item string) error {
	_, err := bf.client.Do(ctx, "BF.ADD", bf.key, item).Result()
	if err != nil {
		return fmt.Errorf("add item to bloom filter error: %w", err)
	}
	return nil
}

// Exists 检查元素是否存在于布隆过滤器中
func (bf *BloomFilter) Exists(ctx context.Context, item string) (bool, error) {
	result, err := bf.client.Do(ctx, "BF.EXISTS", bf.key, item).Result()
	if err != nil {
		return false, fmt.Errorf("check item in bloom filter error: %w", err)
	}
	return result.(int64) == 1, nil
}

// ExistsFilter 检查布隆过滤器本身是否存在（在Redis中）
func (bf *BloomFilter) ExistsFilter(ctx context.Context) (bool, error) {
	result, err := bf.client.Exists(ctx, bf.key).Result()
	if err != nil {
		return false, fmt.Errorf("check bloom filter exists error: %w", err)
	}
	return result == 1, nil
}

// Reserve 初始化布隆过滤器
func (bf *BloomFilter) Reserve(ctx context.Context, size int64, errorRate float64) error {
	// 先检查布隆过滤器是否已存在
	exists, err := bf.ExistsFilter(ctx)
	if err != nil {
		return err
	}
	// 如果已存在，则跳过创建
	if exists {
		return nil
	}

	_, err = bf.client.Do(ctx, "BF.RESERVE", bf.key, errorRate, size).Result()
	if err != nil {
		return fmt.Errorf("reserve bloom filter error: %w", err)
	}
	return nil
}

// AddMulti 批量添加元素到布隆过滤器
func (bf *BloomFilter) AddMulti(ctx context.Context, items []string) error {
	args := []interface{}{"BF.MADD", bf.key}
	for _, item := range items {
		args = append(args, item)
	}
	_, err := bf.client.Do(ctx, args...).Result()
	if err != nil {
		return fmt.Errorf("add multi items to bloom filter error: %w", err)
	}
	return nil
}

// ExistsMulti 批量检查元素是否存在于布隆过滤器中
func (bf *BloomFilter) ExistsMulti(ctx context.Context, items []string) (map[string]bool, error) {
	args := []interface{}{"BF.MEXISTS", bf.key}
	for _, item := range items {
		args = append(args, item)
	}
	result, err := bf.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("check multi items in bloom filter error: %w", err)
	}

	existsMap := make(map[string]bool)
	results := result.([]interface{})
	for i, item := range items {
		existsMap[item] = results[i].(int64) == 1
	}
	return existsMap, nil
}
