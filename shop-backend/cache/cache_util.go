package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"shop-backend/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CacheUtil 缓存工具类
// 提供各种缓存操作方法，包括：
// 1. 布隆过滤器操作
// 2. 商品缓存操作
// 3. 订单缓存操作
// 4. 空值缓存处理
// 5. 缓存键生成
// 6. 缓存过期管理
type CacheUtil struct {
	db    *gorm.DB      // 数据库连接，用于初始化布隆过滤器时查询数据
	Redis *redis.Client // Redis客户端，用于执行缓存操作
}

// NewCacheUtil 创建缓存工具实例
// 参数：
// - db: 数据库连接
// - redis: Redis客户端
// 返回：
// - *CacheUtil: 缓存工具实例
func NewCacheUtil(db *gorm.DB, redis *redis.Client) *CacheUtil {
	return &CacheUtil{
		db:    db,
		Redis: redis,
	}
}

// 缓存相关常量定义
const (
	// 布隆过滤器键名
	ProductBloomFilterKey = "product:bloom:filter" // 商品布隆过滤器键名
	OrderBloomFilterKey   = "order:bloom:filter"   // 订单布隆过滤器键名

	// 空值缓存过期时间（5分钟）
	// 用于缓存不存在的数据，防止缓存穿透
	NullValueExpiration = 5 * time.Minute

	// 商品详情缓存基础过期时间（1小时）
	ProductDetailExpiration = 1 * time.Hour
	// 商品详情缓存随机偏移量（0-300秒）
	// 用于防止缓存雪崩
	ProductDetailExpirationOffset = 300 * time.Second

	// 商品列表缓存基础过期时间（30分钟）
	ProductListExpiration = 30 * time.Minute
	// 商品列表缓存随机偏移量（0-300秒）
	ProductListExpirationOffset = 300 * time.Second

	// 订单详情缓存基础过期时间（30分钟）
	OrderDetailExpiration = 30 * time.Minute
	// 订单详情缓存随机偏移量（0-300秒）
	OrderDetailExpirationOffset = 300 * time.Second

	// 分布式锁过期时间（30秒）
	LockExpiration = 30 * time.Second
	// 分布式锁等待超时时间（5秒）
	LockTimeout = 5 * time.Second
)

// ProductCacheData 商品缓存数据结构
// 包含逻辑过期时间，用于实现缓存的逻辑过期机制
type ProductCacheData struct {
	Data      interface{} `json:"data"`       // 缓存的数据
	ExpiresAt int64       `json:"expires_at"` // 逻辑过期时间戳（Unix秒）
}

// InitBloomFilters 初始化布隆过滤器并预热数据
// 参数：
// - ctx: 上下文，用于控制操作超时
// 返回：
// - error: 操作错误信息
// 说明：
// 1. 初始化商品和订单布隆过滤器
// 2. 预分配布隆过滤器空间
// 3. 预热现有数据到布隆过滤器
func (cu *CacheUtil) InitBloomFilters(ctx context.Context) error {
	// 初始化商品布隆过滤器
	productBF := NewBloomFilter(cu.Redis, ProductBloomFilterKey)
	// 预分配大小，假设最多100万商品，误判率0.01
	err := productBF.Reserve(ctx, 1000000, 0.01)
	if err != nil {
		return fmt.Errorf("reserve product bloom filter error: %w", err)
	}

	// 初始化订单布隆过滤器
	orderBF := NewBloomFilter(cu.Redis, OrderBloomFilterKey)
	// 预分配大小，假设最多100万订单，误判率0.01
	err = orderBF.Reserve(ctx, 1000000, 0.01)
	if err != nil {
		return fmt.Errorf("reserve order bloom filter error: %w", err)
	}

	// 预热商品数据到布隆过滤器
	err = cu.warmupProductData(ctx, productBF)
	if err != nil {
		return fmt.Errorf("warmup product data error: %w", err)
	}

	// 预热订单数据到布隆过滤器
	err = cu.warmupOrderData(ctx, orderBF)
	if err != nil {
		return fmt.Errorf("warmup order data error: %w", err)
	}

	return nil
}

// warmupProductData 预热商品数据到布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - bf: 布隆过滤器实例
// 返回：
// - error: 操作错误信息
// 说明：
// 分批查询商品ID并添加到布隆过滤器
func (cu *CacheUtil) warmupProductData(ctx context.Context, bf *BloomFilter) error {
	// 分批查询商品ID
	var products []models.Product
	batchSize := 1000 // 每批查询1000条
	offset := 0

	for {
		// 查询商品ID
		result := cu.db.Model(&models.Product{}).Select("id").Offset(offset).Limit(batchSize).Find(&products)
		if result.Error != nil {
			return result.Error
		}

		if len(products) == 0 {
			break // 没有更多数据
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

		// 移动到下一批
		offset += batchSize
		products = products[:0] // 清空切片，准备下一批
	}

	return nil
}

// warmupOrderData 预热订单数据到布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - bf: 布隆过滤器实例
// 返回：
// - error: 操作错误信息
// 说明：
// 分批查询订单号并添加到布隆过滤器
func (cu *CacheUtil) warmupOrderData(ctx context.Context, bf *BloomFilter) error {
	// 分批查询订单号
	var orders []models.Order
	batchSize := 1000 // 每批查询1000条
	offset := 0

	for {
		// 查询订单号
		result := cu.db.Model(&models.Order{}).Select("order_no").Offset(offset).Limit(batchSize).Find(&orders)
		if result.Error != nil {
			return result.Error
		}

		if len(orders) == 0 {
			break // 没有更多数据
		}

		// 批量添加到布隆过滤器
		var orderNos []string
		for _, o := range orders {
			orderNos = append(orderNos, o.OrderNo)
		}

		err := bf.AddMulti(ctx, orderNos)
		if err != nil {
			return err
		}

		// 移动到下一批
		offset += batchSize
		orders = orders[:0] // 清空切片，准备下一批
	}

	return nil
}

// CheckProductExists 检查商品是否存在（使用布隆过滤器）
// 参数：
// - ctx: 上下文，用于控制操作超时
// - productID: 商品ID
// 返回：
// - bool: true表示商品可能存在，false表示商品肯定不存在
// - error: 操作错误信息
// 说明：
// 使用布隆过滤器快速判断商品是否存在，减少数据库查询
func (cu *CacheUtil) CheckProductExists(ctx context.Context, productID int) (bool, error) {
	bf := NewBloomFilter(cu.Redis, ProductBloomFilterKey)
	return bf.Exists(ctx, strconv.Itoa(productID))
}

// CheckOrderExists 检查订单是否存在（使用布隆过滤器）
// 参数：
// - ctx: 上下文，用于控制操作超时
// - orderNo: 订单号
// 返回：
// - bool: true表示订单可能存在，false表示订单肯定不存在
// - error: 操作错误信息
// 说明：
// 使用布隆过滤器快速判断订单是否存在，减少数据库查询
func (cu *CacheUtil) CheckOrderExists(ctx context.Context, orderNo string) (bool, error) {
	bf := NewBloomFilter(cu.Redis, OrderBloomFilterKey)
	return bf.Exists(ctx, orderNo)
}

// SetNullValue 缓存空值
// 参数：
// - ctx: 上下文，用于控制操作超时
// - key: 缓存键
// 返回：
// - error: 操作错误信息
// 说明：
// 用于缓存不存在的数据，防止缓存穿透
func (cu *CacheUtil) SetNullValue(ctx context.Context, key string) error {
	return cu.Redis.Set(ctx, key, "null", NullValueExpiration).Err()
}

// GetNullValue 检查是否存在空值缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - key: 缓存键
// 返回：
// - bool: true表示存在空值缓存，false表示不存在
// - error: 操作错误信息
// 说明：
// 用于检查是否存在空值缓存，避免重复查询不存在的数据
func (cu *CacheUtil) GetNullValue(ctx context.Context, key string) (bool, error) {
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // 缓存不存在
	}
	if err != nil {
		return false, err // 其他错误
	}
	return val == "null", nil // 检查是否为空值缓存
}

// AddProductToBloomFilter 添加商品ID到布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - productID: 商品ID
// 返回：
// - error: 操作错误信息
// 说明：
// 当新增商品时，将商品ID添加到布隆过滤器
func (cu *CacheUtil) AddProductToBloomFilter(ctx context.Context, productID int) error {
	bf := NewBloomFilter(cu.Redis, ProductBloomFilterKey)
	return bf.Add(ctx, strconv.Itoa(productID))
}

// AddOrderToBloomFilter 添加订单号到布隆过滤器
// 参数：
// - ctx: 上下文，用于控制操作超时
// - orderNo: 订单号
// 返回：
// - error: 操作错误信息
// 说明：
// 当创建订单时，将订单号添加到布隆过滤器
func (cu *CacheUtil) AddOrderToBloomFilter(ctx context.Context, orderNo string) error {
	bf := NewBloomFilter(cu.Redis, OrderBloomFilterKey)
	return bf.Add(ctx, orderNo)
}

// GetProductCacheKey 生成商品缓存键
// 参数：
// - productID: 商品ID
// 返回：
// - string: 缓存键
// 说明：
// 生成标准化的商品缓存键
func (cu *CacheUtil) GetProductCacheKey(productID int) string {
	return fmt.Sprintf("product:detail:%d", productID)
}

// GetProductLockKey 生成商品分布式锁键
// 参数：
// - productID: 商品ID
// 返回：
// - string: 分布式锁键
// 说明：
// 生成标准化的商品分布式锁键，用于防止缓存击穿
func (cu *CacheUtil) GetProductLockKey(productID int) string {
	return fmt.Sprintf("product:lock:%d", productID)
}

// GetProductCache 获取商品详情缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - productID: 商品ID
// 返回：
// - *ProductCacheData: 缓存数据
// - error: 操作错误信息
// 说明：
// 从Redis获取商品详情缓存
func (cu *CacheUtil) GetProductCache(ctx context.Context, productID int) (*ProductCacheData, error) {
	key := cu.GetProductCacheKey(productID)
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存不存在
	}
	if err != nil {
		return nil, err // 其他错误
	}

	// 解析缓存数据
	var cacheData ProductCacheData
	err = json.Unmarshal([]byte(val), &cacheData)
	if err != nil {
		return nil, err // 解析错误
	}

	return &cacheData, nil
}

// SetProductCache 设置商品详情缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - productID: 商品ID
// - data: 缓存数据
// 返回：
// - error: 操作错误信息
// 说明：
// 设置商品详情缓存，包含逻辑过期时间
func (cu *CacheUtil) SetProductCache(ctx context.Context, productID int, data interface{}) error {
	key := cu.GetProductCacheKey(productID)

	// 计算逻辑过期时间
	now := time.Now()
	expiresAt := now.Add(ProductDetailExpiration).Unix()

	// 构建缓存数据
	cacheData := ProductCacheData{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	// 序列化数据
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	// 设置物理过期时间（比逻辑过期时间长，用于后台更新）
	// 添加随机偏移量防止缓存雪崩
	physicalExpiration := ProductDetailExpiration + time.Duration(rand.Int63n(int64(ProductDetailExpirationOffset)))
	return cu.Redis.Set(ctx, key, jsonData, physicalExpiration).Err()
}

// DeleteProductCache 删除商品详情缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - productID: 商品ID
// 返回：
// - error: 操作错误信息
// 说明：
// 当商品数据更新时，删除对应缓存
func (cu *CacheUtil) DeleteProductCache(ctx context.Context, productID int) error {
	key := cu.GetProductCacheKey(productID)
	return cu.Redis.Del(ctx, key).Err()
}

// GetProductListCacheKey 生成商品列表缓存键
// 参数：
// - page: 页码
// - limit: 每页条数
// - keyword: 搜索关键词
// 返回：
// - string: 缓存键
// 说明：
// 生成标准化的商品列表缓存键，包含分页和搜索条件
func (cu *CacheUtil) GetProductListCacheKey(page, limit int, keyword string) string {
	return fmt.Sprintf("product:list:%d:%d:%s", page, limit, keyword)
}

// GetProductListCache 获取商品列表缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - page: 页码
// - limit: 每页条数
// - keyword: 搜索关键词
// 返回：
// - interface{}: 缓存数据
// - error: 操作错误信息
// 说明：
// 从Redis获取商品列表缓存
func (cu *CacheUtil) GetProductListCache(ctx context.Context, page, limit int, keyword string) (interface{}, error) {
	key := cu.GetProductListCacheKey(page, limit, keyword)
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存不存在
	}
	if err != nil {
		return nil, err // 其他错误
	}

	// 解析缓存数据
	var data interface{}
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, err // 解析错误
	}

	return data, nil
}

// SetProductListCache 设置商品列表缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - page: 页码
// - limit: 每页条数
// - keyword: 搜索关键词
// - data: 缓存数据
// 返回：
// - error: 操作错误信息
// 说明：
// 设置商品列表缓存，添加随机偏移量防止缓存雪崩
func (cu *CacheUtil) SetProductListCache(ctx context.Context, page, limit int, keyword string, data interface{}) error {
	key := cu.GetProductListCacheKey(page, limit, keyword)

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 设置物理过期时间，添加随机偏移量防止缓存雪崩
	physicalExpiration := ProductListExpiration + time.Duration(rand.Int63n(int64(ProductListExpirationOffset)))
	return cu.Redis.Set(ctx, key, jsonData, physicalExpiration).Err()
}

// DeleteProductListCache 删除商品列表缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// 返回：
// - error: 操作错误信息
// 说明：
// 删除所有商品列表缓存，当商品数据更新时使用
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
// 参数：
// - cacheData: 缓存数据
// 返回：
// - bool: true表示缓存已过期，false表示缓存未过期
// 说明：
// 检查缓存的逻辑过期时间
func (cu *CacheUtil) IsCacheExpired(cacheData *ProductCacheData) bool {
	if cacheData == nil {
		return true // 缓存不存在，视为过期
	}
	return time.Now().Unix() > cacheData.ExpiresAt // 比较当前时间与过期时间
}

// GetOrderCacheKey 生成订单缓存键（包含用户ID，确保权限隔离）
// 参数：
// - orderNo: 订单号
// - customerID: 客户ID
// 返回：
// - string: 缓存键
// 说明：
// 生成标准化的订单缓存键，包含客户ID以确保权限隔离
func (cu *CacheUtil) GetOrderCacheKey(orderNo string, customerID int) string {
	return fmt.Sprintf("order:detail:%s:%d", orderNo, customerID)
}

// GetOrderLockKey 生成订单分布式锁键
// 参数：
// - orderNo: 订单号
// 返回：
// - string: 分布式锁键
// 说明：
// 生成标准化的订单分布式锁键，用于防止缓存击穿
func (cu *CacheUtil) GetOrderLockKey(orderNo string) string {
	return fmt.Sprintf("order:lock:%s", orderNo)
}

// GetOrderCache 获取订单详情缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - orderNo: 订单号
// - customerID: 客户ID
// 返回：
// - interface{}: 缓存数据
// - error: 操作错误信息
// 说明：
// 从Redis获取订单详情缓存
func (cu *CacheUtil) GetOrderCache(ctx context.Context, orderNo string, customerID int) (interface{}, error) {
	key := cu.GetOrderCacheKey(orderNo, customerID)
	val, err := cu.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存不存在
	}
	if err != nil {
		return nil, err // 其他错误
	}

	// 解析缓存数据
	var data interface{}
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, err // 解析错误
	}

	return data, nil
}

// SetOrderCache 设置订单详情缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - orderNo: 订单号
// - customerID: 客户ID
// - data: 缓存数据
// 返回：
// - error: 操作错误信息
// 说明：
// 设置订单详情缓存，添加随机偏移量防止缓存雪崩
func (cu *CacheUtil) SetOrderCache(ctx context.Context, orderNo string, customerID int, data interface{}) error {
	key := cu.GetOrderCacheKey(orderNo, customerID)

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 设置物理过期时间，添加随机偏移量防止缓存雪崩
	physicalExpiration := OrderDetailExpiration + time.Duration(rand.Int63n(int64(OrderDetailExpirationOffset)))
	return cu.Redis.Set(ctx, key, jsonData, physicalExpiration).Err()
}

// DeleteOrderCache 删除订单详情缓存
// 参数：
// - ctx: 上下文，用于控制操作超时
// - orderNo: 订单号
// - customerID: 客户ID
// 返回：
// - error: 操作错误信息
// 说明：
// 当订单数据更新时，删除对应缓存
func (cu *CacheUtil) DeleteOrderCache(ctx context.Context, orderNo string, customerID int) error {
	key := cu.GetOrderCacheKey(orderNo, customerID)
	return cu.Redis.Del(ctx, key).Err()
}

// DeleteOrderCacheByOrderNo 根据订单号删除所有相关缓存（适用于管理员操作）
// 参数：
// - ctx: 上下文，用于控制操作超时
// - orderNo: 订单号
// 返回：
// - error: 操作错误信息
// 说明：
// 删除该订单号的所有缓存，适用于管理员操作
func (cu *CacheUtil) DeleteOrderCacheByOrderNo(ctx context.Context, orderNo string) error {
	// 匹配所有包含该订单号的缓存键
	keys, err := cu.Redis.Keys(ctx, fmt.Sprintf("order:detail:%s:*", orderNo)).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return cu.Redis.Del(ctx, keys...).Err()
	}

	return nil
}
