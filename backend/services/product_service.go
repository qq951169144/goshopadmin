package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goshopadmin/cache"
	"goshopadmin/constants"
	"goshopadmin/models"
	"goshopadmin/utils"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ProductService 商品服务
type ProductService struct {
	DB        *gorm.DB
	Redis     *redis.Client
	CacheUtil *cache.CacheUtil
}

// NewProductService 创建商品服务实例
func NewProductService(db *gorm.DB, redis *redis.Client) *ProductService {
	cacheUtil := cache.NewCacheUtil(db, redis)
	return &ProductService{DB: db, Redis: redis, CacheUtil: cacheUtil}
}

// DeleteProductCache 删除商品缓存
func (s *ProductService) DeleteProductCache(ctx context.Context, productID int) error {
	// 删除商品详情缓存
	s.CacheUtil.DeleteProductCache(ctx, productID)

	// 删除商品空值缓存
	nullKey := fmt.Sprintf("product:null:%d", productID)
	s.CacheUtil.SetNullValue(ctx, nullKey)

	// 删除商品列表缓存
	s.CacheUtil.DeleteProductListCache(ctx)

	return nil
}

// AddProductToBloomFilter 添加商品到布隆过滤器
func (s *ProductService) AddProductToBloomFilter(ctx context.Context, productID int) error {
	// 使用缓存工具添加商品到布隆过滤器
	return s.CacheUtil.AddProductToBloomFilter(ctx, productID)
}

// GetProductsByMerchantID 获取商户的商品列表（带缓存）
func (s *ProductService) GetProductsByMerchantID(merchantID int, name string) ([]models.Product, error) {
	ctx := context.Background()

	// 生成缓存键（基于商户ID和名称关键字）
	cacheKey := fmt.Sprintf("product:list:merchant:%d:name:%s", merchantID, name)

	// 查询缓存
	val, err := s.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.Product
		if json.Unmarshal([]byte(val), &products) == nil {
			return products, nil
		}
	}

	// 缓存未命中，查询数据库
	var products []models.Product
	query := s.DB.Where("merchant_id = ?", merchantID)

	// 如果有名称关键字，添加模糊查询条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	result := query.Preload("Category").Preload("Images").Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}

	// 将查询结果写入缓存
	if len(products) > 0 {
		jsonData, _ := json.Marshal(products)
		s.Redis.Set(ctx, cacheKey, jsonData, 30*time.Minute)
	}

	return products, nil
}

// GetProductByID 根据ID获取商品详情（带缓存）
func (s *ProductService) GetProductByID(id int, merchantID int) (models.Product, error) {
	ctx := context.Background()

	// 1. 检查空值缓存
	nullKey := fmt.Sprintf("product:null:%d", id)
	nullExists, err := s.CacheUtil.GetNullValue(ctx, nullKey)
	utils.Info("空值检查nullKey = %s, nullExists = %s error", nullKey, nullExists, err)
	if err == nil && nullExists {
		return models.Product{}, err
	}

	// 2. 检查布隆过滤器
	exists, err := s.CacheUtil.CheckProductExists(ctx, id)
	utils.Info("空值检查exists = %s, error = %s", exists, err)
	if err == nil && !exists {
		// 缓存空值
		s.CacheUtil.SetNullValue(ctx, nullKey)
		return models.Product{}, err
	}

	// 3. 查询缓存
	cachedData, err := s.CacheUtil.GetProductCache(ctx, id)
	if err == nil && cachedData != nil && !s.CacheUtil.IsCacheExpired(cachedData) {
		if productData, ok := cachedData.Data.(map[string]interface{}); ok {
			// 检查商品是否属于该商户
			if merchantIDFloat, ok := productData["MerchantID"].(float64); ok {
				if int(merchantIDFloat) == merchantID {
					// 转换回Product对象
					var product models.Product
					jsonData, _ := json.Marshal(productData)
					json.Unmarshal(jsonData, &product)
					return product, nil
				}
			}
		}
	}

	// 4. 缓存未命中，查询数据库
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).Preload("Category").Preload("Images").Preload("Skus").First(&product)
	if result.Error != nil {
		// 缓存空值
		s.CacheUtil.SetNullValue(ctx, nullKey)
		return models.Product{}, result.Error
	}

	// 5. 将查询结果写入缓存
	s.CacheUtil.SetProductCache(ctx, id, product)

	return product, nil
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(product *models.Product, merchantID int) error {
	product.MerchantID = merchantID
	// 创建商品（不再自动创建默认SKU）
	result := s.DB.Create(product)
	if result.Error != nil {
		return result.Error
	}

	// 商品创建成功后，添加到布隆过滤器并清理相关缓存
	ctx := context.Background()
	s.AddProductToBloomFilter(ctx, int(product.ID))
	s.DeleteProductCache(ctx, int(product.ID))

	return nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(productID int, merchantID int, updateData map[string]interface{}) error {
	// 检查商品是否属于该商户
	var existingProduct models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", productID, merchantID).First(&existingProduct)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 使用map更新指定字段
	result = s.DB.Model(&existingProduct).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}

	// 商品更新成功后，清理相关缓存
	ctx := context.Background()
	s.DeleteProductCache(ctx, productID)

	return nil
}

// DeleteProduct 禁用商品
func (s *ProductService) DeleteProduct(id int, merchantID int) error {
	// 检查商品是否属于该商户
	var existingProduct models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&existingProduct)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 更新商品状态为禁用
	result = s.DB.Model(&existingProduct).Update("status", constants.StatusInactive)
	if result.Error != nil {
		return result.Error
	}

	// 商品禁用成功后，清理相关缓存
	ctx := context.Background()
	s.DeleteProductCache(ctx, id)

	return nil
}

// GetCategoriesByMerchantID 获取商户的商品分类列表
func (s *ProductService) GetCategoriesByMerchantID(merchantID int) ([]models.ProductCategory, error) {
	var categories []models.ProductCategory
	result := s.DB.Where("merchant_id = ?", merchantID).Order("sort ASC").Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

// GetCategoryByID 根据ID获取商品分类详情
func (s *ProductService) GetCategoryByID(id int, merchantID int) (models.ProductCategory, error) {
	var category models.ProductCategory
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&category)
	if result.Error != nil {
		return models.ProductCategory{}, result.Error
	}
	return category, nil
}

// CreateCategory 创建商品分类
func (s *ProductService) CreateCategory(category *models.ProductCategory, merchantID int) error {
	category.MerchantID = merchantID
	result := s.DB.Create(category)
	return result.Error
}

// UpdateCategory 更新商品分类
func (s *ProductService) UpdateCategory(category *models.ProductCategory, merchantID int) error {
	// 检查分类是否属于该商户
	var existingCategory models.ProductCategory
	result := s.DB.Where("id = ? AND merchant_id = ?", category.ID, merchantID).First(&existingCategory)
	if result.Error != nil {
		return errors.New("分类不存在或不属于该商户")
	}

	// 使用Updates方法更新，只更新非零值字段，避免覆盖时间戳
	result = s.DB.Model(&existingCategory).Updates(category)
	return result.Error
}

// DeleteCategory 禁用商品分类
func (s *ProductService) DeleteCategory(id int, merchantID int) error {
	// 检查分类是否属于该商户
	var existingCategory models.ProductCategory
	result := s.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&existingCategory)
	if result.Error != nil {
		return errors.New("分类不存在或不属于该商户")
	}

	// 检查分类下是否有商品
	var productCount int64
	s.DB.Model(&models.Product{}).Where("category_id = ?", id).Count(&productCount)
	if productCount > 0 {
		return errors.New("分类下存在商品，无法禁用")
	}

	// 更新分类状态为禁用
	result = s.DB.Model(&existingCategory).Update("status", constants.StatusInactive)
	return result.Error
}

// GetProductImageByID 根据ID获取商品图片
func (s *ProductService) GetProductImageByID(id int, merchantID int) (*models.ProductImage, error) {
	var image models.ProductImage
	result := s.DB.First(&image, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 检查图片所属的商品是否属于该商户
	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", image.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return nil, errors.New("商品不存在或不属于该商户")
	}

	return &image, nil
}

// AddProductImage 添加商品图片
func (s *ProductService) AddProductImage(image *models.ProductImage, merchantID int) error {
	// 检查商品是否属于该商户
	var product models.Product
	result := s.DB.Where("id = ? AND merchant_id = ?", image.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 如果是主图，将其他图片设置为非主图
	if image.IsMain {
		s.DB.Model(&models.ProductImage{}).Where("product_id = ?", image.ProductID).Update("is_main", false)
	}

	// 添加图片
	result = s.DB.Create(image)
	if result.Error != nil {
		return result.Error
	}

	// 商品图片添加成功后，清理相关商品缓存
	ctx := context.Background()
	s.DeleteProductCache(ctx, image.ProductID)

	return nil
}

// DeleteProductImage 删除商品图片
func (s *ProductService) DeleteProductImage(id int, merchantID int) error {
	// 检查图片所属的商品是否属于该商户
	var image models.ProductImage
	result := s.DB.First(&image, id)
	if result.Error != nil {
		return errors.New("图片不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", image.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 保存商品ID，用于后续清理缓存
	productID := image.ProductID

	// 删除图片
	result = s.DB.Delete(&image)
	if result.Error != nil {
		return result.Error
	}

	// 商品图片删除成功后，清理相关商品缓存
	ctx := context.Background()
	s.DeleteProductCache(ctx, productID)

	return nil
}

// UpdateProductImage 更新商品图片
func (s *ProductService) UpdateProductImage(image *models.ProductImage, merchantID int) error {
	// 检查图片所属的商品是否属于该商户
	var existingImage models.ProductImage
	result := s.DB.First(&existingImage, image.ID)
	if result.Error != nil {
		return errors.New("图片不存在")
	}

	var product models.Product
	result = s.DB.Where("id = ? AND merchant_id = ?", existingImage.ProductID, merchantID).First(&product)
	if result.Error != nil {
		return errors.New("商品不存在或不属于该商户")
	}

	// 保存商品ID，用于后续清理缓存
	productID := existingImage.ProductID

	// 开启事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 如果设置为主图，将其他图片设置为非主图
	if image.IsMain {
		if err := tx.Model(&models.ProductImage{}).Where("product_id = ? AND id != ?", existingImage.ProductID, image.ID).Update("is_main", false).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新图片信息
	if err := tx.Model(&existingImage).Updates(image).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 商品图片更新成功后，清理相关商品缓存
	ctx := context.Background()
	s.DeleteProductCache(ctx, productID)

	return nil
}
