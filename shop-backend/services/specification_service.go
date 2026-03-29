package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"shop-backend/cache"
	"shop-backend/constants"
	"shop-backend/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SpecificationService 规格服务
type SpecificationService struct {
	db        *gorm.DB
	cacheUtil *cache.CacheUtil
}

// NewSpecificationService 创建规格服务实例
func NewSpecificationService(db *gorm.DB, cacheUtil *cache.CacheUtil) *SpecificationService {
	return &SpecificationService{
		db:        db,
		cacheUtil: cacheUtil,
	}
}

// SpecificationInfo 规格信息结构
type SpecificationInfo struct {
	ID     int                      `json:"id"`
	Name   string                   `json:"name"`
	Values []SpecificationValueInfo `json:"values"`
}

// SpecificationValueInfo 规格值信息结构
type SpecificationValueInfo struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Image string `json:"image"`
}

// SkuInfoWithSpecs 带规格信息的SKU
type SkuInfoWithSpecs struct {
	ID              int            `json:"id"`
	SkuCode         string         `json:"sku_code"`
	Price           float64        `json:"price"`
	OriginalPrice   float64        `json:"original_price"`
	Stock           int            `json:"stock"`
	Status          string         `json:"status"`
	SpecCombination map[string]int `json:"spec_combination"` // spec_id -> spec_value_id
}

// ProductDetailWithSpecs 带规格信息的商品详情
type ProductDetailWithSpecs struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Detail         string              `json:"detail"`
	Price          float64             `json:"price"`
	Image          string              `json:"image"`
	Images         []string            `json:"images"`
	Specifications []SpecificationInfo `json:"specifications"`
	Skus           []SkuInfoWithSpecs  `json:"sku_list"`
	PriceRange     PriceRange          `json:"price_range"`
	Sales          int                 `json:"sales"`
	ReviewsCount   int                 `json:"reviews_count"`
}

// PriceRange 价格范围
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// GetProductDetailWithSpecs 获取带规格信息的商品详情（带缓存策略）
// 缓存策略：布隆过滤器 + 缓存 + 分布式锁
func (s *SpecificationService) GetProductDetailWithSpecs(productID int) (*ProductDetailWithSpecs, error) {
	ctx := context.Background()

	// 1. 检查空值缓存（防止缓存穿透）
	nullKey := fmt.Sprintf("product:null:%d", productID)
	nullExists, err := s.cacheUtil.GetNullValue(ctx, nullKey)
	if err == nil && nullExists {
		return nil, errors.New("商品不存在")
	}

	// 2. 检查布隆过滤器（快速判断商品是否存在）
	exists, err := s.cacheUtil.CheckProductExists(ctx, productID)
	if err == nil && !exists {
		// 布隆过滤器判断不存在，设置空值缓存
		s.cacheUtil.SetNullValue(ctx, nullKey)
		return nil, errors.New("商品不存在")
	}

	// 3. 尝试从缓存获取数据
	cacheKey := s.cacheUtil.GetProductCacheKey(productID)
	cachedData, err := s.cacheUtil.GetProductCache(ctx, productID)
	if err == nil && cachedData != nil {
		// 缓存命中，检查是否逻辑过期
		if !s.cacheUtil.IsCacheExpired(cachedData) {
			// 缓存未过期，直接返回
			if data, ok := cachedData.Data.(*ProductDetailWithSpecs); ok {
				return data, nil
			}
			// 类型不匹配，尝试解析JSON
			jsonData, _ := json.Marshal(cachedData.Data)
			var result ProductDetailWithSpecs
			if err := json.Unmarshal(jsonData, &result); err == nil {
				return &result, nil
			}
		}
		// 缓存已过期，继续执行后续逻辑（返回旧数据并后台更新）
		if data, ok := cachedData.Data.(*ProductDetailWithSpecs); ok {
			// 后台异步更新缓存
			go s.rebuildProductCache(productID)
			return data, nil
		}
	}

	// 4. 获取分布式锁（防止缓存击穿）
	lockKey := s.cacheUtil.GetProductLockKey(productID)
	lockValue := uuid.New().String()
	distributedLock := cache.NewDistributedLockFromCacheUtil(s.cacheUtil, lockKey, lockValue)

	lockAcquired, err := distributedLock.Acquire(ctx, cache.LockExpiration)
	if err != nil {
		// 获取锁失败，直接查询数据库
		return s.queryProductDetailFromDB(productID)
	}

	if !lockAcquired {
		// 未获取到锁，说明其他进程正在重建缓存，短暂等待后重试读取缓存
		time.Sleep(100 * time.Millisecond)
		cachedData, err := s.cacheUtil.GetProductCache(ctx, productID)
		if err == nil && cachedData != nil {
			if data, ok := cachedData.Data.(*ProductDetailWithSpecs); ok {
				return data, nil
			}
		}
		// 缓存仍未命中，直接查询数据库
		return s.queryProductDetailFromDB(productID)
	}

	// 获取到锁，查询数据库并重建缓存
	defer distributedLock.Release(ctx)

	// 5. 查询数据库
	result, err := s.queryProductDetailFromDB(productID)
	if err != nil {
		// 商品不存在，设置空值缓存
		if err.Error() == "商品不存在" {
			s.cacheUtil.SetNullValue(ctx, nullKey)
		}
		return nil, err
	}

	// 6. 将查询结果写入缓存，添加随机过期时间防止缓存雪崩
	physicalExpiration := cache.ProductDetailExpiration + time.Duration(rand.Int63n(int64(cache.ProductDetailExpirationOffset)))
	s.cacheUtil.SetProductCache(ctx, productID, result)
	// 更新实际过期时间
	s.cacheUtil.Redis.Expire(ctx, cacheKey, physicalExpiration)

	return result, nil
}

// queryProductDetailFromDB 从数据库查询商品详情
func (s *SpecificationService) queryProductDetailFromDB(productID int) (*ProductDetailWithSpecs, error) {
	var product models.Product
	result := s.db.First(&product, productID)
	if result.RowsAffected == 0 {
		return nil, errors.New("商品不存在")
	}

	// 查询商品图片列表
	var productImages []models.ProductImage
	s.db.Where("product_id = ?", productID).Order("is_main DESC, sort ASC").Find(&productImages)

	var images []string
	var mainImage string
	for _, img := range productImages {
		if img.ImageURL != "" {
			images = append(images, img.ImageURL)
			if img.IsMain && mainImage == "" {
				mainImage = img.ImageURL
			}
		}
	}
	if mainImage == "" && len(images) > 0 {
		mainImage = images[0]
	}

	// 查询规格列表
	var specifications []models.ProductSpecification
	s.db.Where("product_id = ?", productID).Order("sort ASC").Find(&specifications)

	var specInfos []SpecificationInfo
	for _, spec := range specifications {
		var values []models.ProductSpecificationValue
		s.db.Where("spec_id = ? AND status = ?", spec.ID, constants.StatusActive).Order("sort ASC").Find(&values)

		var valueInfos []SpecificationValueInfo
		for _, value := range values {
			valueInfos = append(valueInfos, SpecificationValueInfo{
				ID:    value.ID,
				Value: value.Value,
			})
		}

		specInfos = append(specInfos, SpecificationInfo{
			ID:     spec.ID,
			Name:   spec.Name,
			Values: valueInfos,
		})
	}

	// 查询SKU列表
	var skus []models.ProductSku
	s.db.Where("product_id = ?", productID).Find(&skus)

	var skuInfos []SkuInfoWithSpecs
	var minPrice, maxPrice float64
	for i, sku := range skus {
		// 查询SKU的规格组合
		var skuSpecs []models.ProductSkuSpec
		s.db.Where("sku_id = ?", sku.ID).Find(&skuSpecs)

		specCombination := make(map[string]int)
		for _, skuSpec := range skuSpecs {
			specCombination[strconv.Itoa(skuSpec.SpecID)] = skuSpec.SpecValueID
		}

		skuInfos = append(skuInfos, SkuInfoWithSpecs{
			ID:              sku.ID,
			SkuCode:         sku.SkuCode,
			Price:           sku.Price,
			OriginalPrice:   sku.OriginalPrice,
			Stock:           sku.Stock,
			Status:          sku.Status,
			SpecCombination: specCombination,
		})

		// 计算价格范围
		if i == 0 || sku.Price < minPrice {
			minPrice = sku.Price
		}
		if i == 0 || sku.Price > maxPrice {
			maxPrice = sku.Price
		}
	}

	// 如果没有SKU，使用商品默认价格
	if len(skuInfos) == 0 {
		skuInfos = append(skuInfos, SkuInfoWithSpecs{
			ID:              0,
			SkuCode:         "default",
			Price:           product.Price,
			OriginalPrice:   0,
			Stock:           product.Stock,
			Status:          "active",
			SpecCombination: make(map[string]int),
		})
		minPrice = product.Price
		maxPrice = product.Price
	}

	return &ProductDetailWithSpecs{
		ID:             product.ID,
		Name:           product.Name,
		Description:    product.Description,
		Detail:         product.Detail,
		Price:          product.Price,
		Image:          mainImage,
		Images:         images,
		Specifications: specInfos,
		Skus:           skuInfos,
		PriceRange: PriceRange{
			Min: minPrice,
			Max: maxPrice,
		},
		Sales:        0,
		ReviewsCount: 0,
	}, nil
}

// rebuildProductCache 后台重建商品缓存
func (s *SpecificationService) rebuildProductCache(productID int) {
	ctx := context.Background()

	// 获取分布式锁
	lockKey := s.cacheUtil.GetProductLockKey(productID)
	lockValue := uuid.New().String()
	distributedLock := cache.NewDistributedLockFromCacheUtil(s.cacheUtil, lockKey, lockValue)

	lockAcquired, err := distributedLock.Acquire(ctx, cache.LockExpiration)
	if err != nil || !lockAcquired {
		return // 获取锁失败，放弃重建
	}
	defer distributedLock.Release(ctx)

	// 查询数据库
	result, err := s.queryProductDetailFromDB(productID)
	if err != nil {
		return // 查询失败，放弃重建
	}

	// 更新缓存
	physicalExpiration := cache.ProductDetailExpiration + time.Duration(rand.Int63n(int64(cache.ProductDetailExpirationOffset)))
	s.cacheUtil.SetProductCache(ctx, productID, result)
	cacheKey := s.cacheUtil.GetProductCacheKey(productID)
	s.cacheUtil.Redis.Expire(ctx, cacheKey, physicalExpiration)
}

// GetSkusByProductID 获取商品的SKU列表
func (s *SpecificationService) GetSkusByProductID(productID int) ([]SkuInfoWithSpecs, error) {
	var skus []models.ProductSku
	s.db.Where("product_id = ?", productID).Find(&skus)

	var skuInfos []SkuInfoWithSpecs
	for _, sku := range skus {
		// 查询SKU的规格组合
		var skuSpecs []models.ProductSkuSpec
		s.db.Where("sku_id = ?", sku.ID).Find(&skuSpecs)

		specCombination := make(map[string]int)
		for _, skuSpec := range skuSpecs {
			specCombination[strconv.Itoa(skuSpec.SpecID)] = skuSpec.SpecValueID
		}

		skuInfos = append(skuInfos, SkuInfoWithSpecs{
			ID:              sku.ID,
			SkuCode:         sku.SkuCode,
			Price:           sku.Price,
			OriginalPrice:   sku.OriginalPrice,
			Stock:           sku.Stock,
			Status:          sku.Status,
			SpecCombination: specCombination,
		})
	}

	return skuInfos, nil
}

// GetSkuBySpecCombination 根据规格组合查询SKU
func (s *SpecificationService) GetSkuBySpecCombination(productID int, specQuery string) (*SkuInfoWithSpecs, error) {
	// 解析规格查询参数，格式: "1:1,2:4" 表示 spec_id=1 对应 spec_value_id=1, spec_id=2 对应 spec_value_id=4
	specPairs := strings.Split(specQuery, ",")
	if len(specPairs) == 0 {
		return nil, errors.New("规格参数格式错误")
	}

	// 构建规格条件
	specConditions := make(map[int]int)
	for _, pair := range specPairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		specID, _ := strconv.Atoi(parts[0])
		specValueID, _ := strconv.Atoi(parts[1])
		if specID > 0 && specValueID > 0 {
			specConditions[specID] = specValueID
		}
	}

	if len(specConditions) == 0 {
		return nil, errors.New("规格参数格式错误")
	}

	// 查询所有SKU
	var skus []models.ProductSku
	s.db.Where("product_id = ?", productID).Find(&skus)

	// 查找匹配的SKU
	for _, sku := range skus {
		var skuSpecs []models.ProductSkuSpec
		s.db.Where("sku_id = ?", sku.ID).Find(&skuSpecs)

		// 检查是否匹配所有规格条件
		match := true
		for specID, specValueID := range specConditions {
			found := false
			for _, skuSpec := range skuSpecs {
				if skuSpec.SpecID == specID && skuSpec.SpecValueID == specValueID {
					found = true
					break
				}
			}
			if !found {
				match = false
				break
			}
		}

		if match {
			specCombination := make(map[string]int)
			for _, skuSpec := range skuSpecs {
				specCombination[strconv.Itoa(skuSpec.SpecID)] = skuSpec.SpecValueID
			}

			return &SkuInfoWithSpecs{
				ID:              sku.ID,
				SkuCode:         sku.SkuCode,
				Price:           sku.Price,
				OriginalPrice:   sku.OriginalPrice,
				Stock:           sku.Stock,
				Status:          sku.Status,
				SpecCombination: specCombination,
			}, nil
		}
	}

	return nil, errors.New("未找到匹配的SKU")
}
