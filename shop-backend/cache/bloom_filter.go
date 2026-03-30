package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// BloomFilter 布隆过滤器结构体
// 布隆过滤器是一种空间效率很高的概率型数据结构，用于判断一个元素是否在一个集合中
// 特点：
// 1. 可能会误判（假阳性），但不会漏判（假阴性）
// 2. 空间效率高，适合处理大规模数据
// 3. 查询速度快，时间复杂度为O(k)，k为哈希函数数量
type BloomFilter struct {
	client *redis.Client // Redis客户端，用于执行布隆过滤器命令
	key    string        // 布隆过滤器在Redis中的键名
}

// NewBloomFilter 创建布隆过滤器实例
// 参数：
// - client: Redis客户端实例
// - key: 布隆过滤器在Redis中的键名
// 返回：
// - *BloomFilter: 布隆过滤器实例
func NewBloomFilter(client *redis.Client, key string) *BloomFilter {
	return &BloomFilter{
		client: client,
		key:    key,
	}
}

// Add 添加元素到布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - item: 要添加的元素
// 返回：
// - error: 操作错误信息
func (bf *BloomFilter) Add(ctx context.Context, item string) error {
	// 使用Redis的BF.ADD命令添加元素
	_, err := bf.client.Do(ctx, "BF.ADD", bf.key, item).Result()
	if err != nil {
		return fmt.Errorf("add item to bloom filter error: %w", err)
	}
	return nil
}

// Exists 检查元素是否存在于布隆过滤器中
// 参数：
// - ctx: 上下文，用于控制操作超时
// - item: 要检查的元素
// 返回：
// - bool: true表示元素可能存在，false表示元素肯定不存在
// - error: 操作错误信息
func (bf *BloomFilter) Exists(ctx context.Context, item string) (bool, error) {
	// 使用Redis的BF.EXISTS命令检查元素
	result, err := bf.client.Do(ctx, "BF.EXISTS", bf.key, item).Result()
	if err != nil {
		return false, fmt.Errorf("check item in bloom filter error: %w", err)
	}
	// Redis返回1表示存在，0表示不存在
	return result.(int64) == 1, nil
}

// ExistsFilter 检查布隆过滤器本身是否存在（在Redis中）
// 参数：
// - ctx: 上下文，用于控制操作超时
// 返回：
// - bool: true表示布隆过滤器存在，false表示不存在
// - error: 操作错误信息
func (bf *BloomFilter) ExistsFilter(ctx context.Context) (bool, error) {
	result, err := bf.client.Exists(ctx, bf.key).Result()
	if err != nil {
		return false, fmt.Errorf("check bloom filter exists error: %w", err)
	}
	return result == 1, nil
}

// Reserve 初始化布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - size: 布隆过滤器的预计元素数量
// - errorRate: 期望的误判率（0-1之间）
// 返回：
// - error: 操作错误信息
// 说明：
// 此方法需要在添加元素前调用，用于预分配布隆过滤器的空间
// 误判率越低，需要的空间越大
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

	// 使用Redis的BF.RESERVE命令初始化布隆过滤器
	_, err = bf.client.Do(ctx, "BF.RESERVE", bf.key, errorRate, size).Result()
	if err != nil {
		return fmt.Errorf("reserve bloom filter error: %w", err)
	}
	return nil
}

// AddMulti 批量添加元素到布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - items: 要添加的元素列表
// 返回：
// - error: 操作错误信息
// 说明：
// 批量操作比单个操作更高效，适合初始化时批量添加数据
func (bf *BloomFilter) AddMulti(ctx context.Context, items []string) error {
	// 构建BF.MADD命令的参数
	args := []interface{}{"BF.MADD", bf.key}
	for _, item := range items {
		args = append(args, item)
	}
	// 执行批量添加操作
	_, err := bf.client.Do(ctx, args...).Result()
	if err != nil {
		return fmt.Errorf("add multi items to bloom filter error: %w", err)
	}
	return nil
}

// ExistsMulti 批量检查元素是否存在于布隆过滤器中
// 参数：
// - ctx: 上下文，用于控制操作超时
// - items: 要检查的元素列表
// 返回：
// - map[string]bool: 元素到存在状态的映射
// - error: 操作错误信息
// 说明：
// 批量操作比单个操作更高效，适合同时检查多个元素
func (bf *BloomFilter) ExistsMulti(ctx context.Context, items []string) (map[string]bool, error) {
	// 构建BF.MEXISTS命令的参数
	args := []interface{}{"BF.MEXISTS", bf.key}
	for _, item := range items {
		args = append(args, item)
	}
	// 执行批量检查操作
	result, err := bf.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("check multi items in bloom filter error: %w", err)
	}

	// 处理返回结果，构建映射
	existsMap := make(map[string]bool)
	results := result.([]interface{})
	for i, item := range items {
		// Redis返回1表示存在，0表示不存在
		existsMap[item] = results[i].(int64) == 1
	}
	return existsMap, nil
}
