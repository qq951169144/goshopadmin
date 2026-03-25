package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"goshopadmin/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CacheUtil 缓存工具类
type CacheUtil struct {
	db    *gorm.DB
	Redis *redis.Client
}

// NewCacheUtil 创建缓存工具实例
func NewCacheUtil(db *gorm.DB, redis *redis.Client) *CacheUtil {
	return &CacheUtil{
		db:    db,
		Redis: redis,
	}
}

// 布隆过滤器键名
const (
	ProductBloomFilterKey = "product:bloom:filter"
	// 空值缓存过期时间（5分钟）
	NullValueExpiration = 5 * time.Minute
	// 商品详情缓存基础过期时间（1小时）
	ProductDetailExpiration = 1 * time.Hour
	// 商品详情缓存随机偏移量（0-300秒）
	ProductDetailExpirationOffset = 300 * time.Second
	// 商品列表缓存基础过期时间（30分钟）
	ProductListExpiration = 30 * time.Minute
	// 商品列表缓存随机偏移量（0-300秒）
	ProductListExpirationOffset = 300 * time.Second
	// 分布式锁过期时间（30秒）
	LockExpiration = 30 * time.Second
	// 分布式锁等待超时时间（5秒）
	LockTimeout = 5 * time.Second
)

// ProductCacheData 商品缓存数据结构
type ProductCacheData struct {
	Data      interface{} `json:"data"`
	ExpiresAt int64       `json:"expires_at"` // 逻辑过期时间
}

// InitBloomFilters 初始化布隆过滤器并预热数据
func (cu *CacheUtil) InitBloomFilters(ctx context.Context) error {
	// 初始化商品布隆过滤器
	productBF := NewBloomFilter(cu.Redis, ProductBloomFilterKey)
	// 预分配大小，假设最多100万商品
	err := productBF.Reserve(ctx, 1000000, 0.01)
	if err != nil {
		return fmt.Errorf("reserve product bloom filter error: %w", err)
	}

	// 预热商品数据
	err = cu.warmupProductData(ctx, productBF)
	if err != nil {
		return fmt.Errorf("warmup product data error: %w", err)
	}

	return nil
}

// warmupProductData 预热商品数据到布隆过滤器
func (cu *CacheUtil) warmupProductData(ctx context.Context, bf *BloomFilter) error {
	// 分批查询商品ID
	var products []models.Product
	batchSize := 1000
	offset := 0

	for {
		result := cu.db.Model(&models.Product{}).Select("id").Offset(offset).Limit(batchSize).Find(&products)
		if result.Error != nil {
			return result.Error
		}

		if len(products) == 0 {
			break
		}

		// 批量添加到布隆过滤器
		var productIDs []string
		for _, p := range products {
			productIDs = append(productIDs, strconv.Itoa(int(p.ID)))
		}

		err := bf.AddMulti(ctx, productIDs)
		if err != nil {
			return err
		}

		offset += batchSize
		products = products[:0]
	}

	return nil
}

// CheckProductExists 检查商品是否存在（使用布隆过滤器）
func (cu *CacheUtil) CheckProductExists(ctx context.Context, productID int) (bool, error) {
	bf := NewBloomFilter(cu.Redis, ProductBloomFilterKey)
	return bf.Exists(ctx, strconv.Itoa(productID))
}

// SetNullValue 缓存空值
func (cu *CacheUtil) SetNullValue(ctx context.Context, key string) error {
	return cu.Redis.Set(ctx, key, "null", NullValueExpiration).Err()
}

// GetNullValue 检查是否存在空值缓存
func (cu *CacheUtil) GetNullValue(ctx context.Context, key string) (bool, error) {
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "null", nil
}

// AddProductToBloomFilter 添加商品ID到布隆过滤器
func (cu *CacheUtil) AddProductToBloomFilter(ctx context.Context, productID int) error {
	bf := NewBloomFilter(cu.Redis, ProductBloomFilterKey)
	return bf.Add(ctx, strconv.Itoa(productID))
}

// GetProductCacheKey 生成商品缓存键
func (cu *CacheUtil) GetProductCacheKey(productID int) string {
	return fmt.Sprintf("product:detail:%d", productID)
}

// GetProductLockKey 生成商品分布式锁键
func (cu *CacheUtil) GetProductLockKey(productID int) string {
	return fmt.Sprintf("product:lock:%d", productID)
}

// GetProductCache 获取商品详情缓存
func (cu *CacheUtil) GetProductCache(ctx context.Context, productID int) (*ProductCacheData, error) {
	key := cu.GetProductCacheKey(productID)
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cacheData ProductCacheData
	err = json.Unmarshal([]byte(val), &cacheData)
	if err != nil {
		return nil, err
	}

	return &cacheData, nil
}

// SetProductCache 设置商品详情缓存
func (cu *CacheUtil) SetProductCache(ctx context.Context, productID int, data interface{}) error {
	key := cu.GetProductCacheKey(productID)

	// 计算逻辑过期时间
	now := time.Now()
	expiresAt := now.Add(ProductDetailExpiration).Unix()

	cacheData := ProductCacheData{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	// 设置物理过期时间（比逻辑过期时间长，用于后台更新）
	physicalExpiration := ProductDetailExpiration + time.Duration(rand.Int63n(int64(ProductDetailExpirationOffset)))
	return cu.Redis.Set(ctx, key, jsonData, physicalExpiration).Err()
}

// DeleteProductCache 删除商品详情缓存
func (cu *CacheUtil) DeleteProductCache(ctx context.Context, productID int) error {
	key := cu.GetProductCacheKey(productID)
	return cu.Redis.Del(ctx, key).Err()
}

// GetProductListCacheKey 生成商品列表缓存键
func (cu *CacheUtil) GetProductListCacheKey(page, limit int, keyword string) string {
	return fmt.Sprintf("product:list:%d:%d:%s", page, limit, keyword)
}

// GetProductListCache 获取商品列表缓存
func (cu *CacheUtil) GetProductListCache(ctx context.Context, page, limit int, keyword string) (interface{}, error) {
	key := cu.GetProductListCacheKey(page, limit, keyword)
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// SetProductListCache 设置商品列表缓存
func (cu *CacheUtil) SetProductListCache(ctx context.Context, page, limit int, keyword string, data interface{}) error {
	key := cu.GetProductListCacheKey(page, limit, keyword)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 设置物理过期时间，添加随机偏移量防止缓存雪崩
	physicalExpiration := ProductListExpiration + time.Duration(rand.Int63n(int64(ProductListExpirationOffset)))
	return cu.Redis.Set(ctx, key, jsonData, physicalExpiration).Err()
}

// DeleteProductListCache 删除商品列表缓存
func (cu *CacheUtil) DeleteProductListCache(ctx context.Context) error {
	// 删除所有商品列表缓存
	keys, err := cu.Redis.Keys(ctx, "product:list:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return cu.Redis.Del(ctx, keys...).Err()
	}

	return nil
}

// IsCacheExpired 检查缓存是否过期
func (cu *CacheUtil) IsCacheExpired(cacheData *ProductCacheData) bool {
	if cacheData == nil {
		return true
	}
	return time.Now().Unix() > cacheData.ExpiresAt
}
