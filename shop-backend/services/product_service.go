package services

import (
	"context"
	"time"

	"shop-backend/cache"
	"shop-backend/models"
	"shop-backend/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductService 商品服务
type ProductService struct {
	db              *gorm.DB
	cacheUtil       *cache.CacheUtil
	circuitBreaker  *cache.CircuitBreaker
	multiLevelCache *cache.MultiLevelCache
}

// NewProductService 创建商品服务实例
func NewProductService(db *gorm.DB, cacheUtil *cache.CacheUtil) *ProductService {
	// 初始化熔断器
	circuitBreaker := cache.NewCircuitBreaker(
		"product_service",
		0.5,            // 失败率阈值 50%
		30*time.Second, // 重置超时时间
		5,              // 半开状态下允许的最大请求数
	)

	// 初始化多级缓存
	multiLevelCache := cache.NewMultiLevelCache(
		30*time.Minute,  // 本地缓存默认过期时间
		1*time.Hour,     // 本地缓存清理间隔
		cacheUtil.Redis, // Redis客户端
	)

	return &ProductService{
		db:              db,
		cacheUtil:       cacheUtil,
		circuitBreaker:  circuitBreaker,
		multiLevelCache: multiLevelCache,
	}
}

// ProductInfo 商品信息结构
type ProductInfo struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Price           decimal.Decimal `json:"price"`
	SKU             string          `json:"sku"`
	Stock           int             `json:"stock"`
	Image           string          `json:"image"`
	Images          []string        `json:"images"`
	DefaultSkuPrice decimal.Decimal `json:"default_sku_price"`
	Sales           int             `json:"sales"`
}

// SKUInfo SKU信息结构
type SKUInfo struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Price      decimal.Decimal   `json:"price"`
	Stock      int               `json:"stock"`
	Attributes map[string]string `json:"attributes"`
}

// GetProductsRequest 获取商品列表请求
type GetProductsRequest struct {
	Page            int
	Limit           int
	Keyword         string
	ExcludeActivity bool
}

// GetProductsResponse 获取商品列表响应
type GetProductsResponse struct {
	Products []ProductInfo `json:"products"`
	Total    int64         `json:"total"`
}

// GetProducts 获取商品列表
func (s *ProductService) GetProducts(req GetProductsRequest) (*GetProductsResponse, error) {
	ctx := context.Background()

	// 生成缓存键
	cacheKey := s.cacheUtil.GetProductListCacheKey(req.Page, req.Limit, req.Keyword, req.ExcludeActivity)

	// 1. 尝试从多级缓存获取数据
	var cachedResponse GetProductsResponse
	err := s.multiLevelCache.Get(ctx, cacheKey, &cachedResponse)
	utils.Info("检查日志生成:cachedResponse = %s", cachedResponse)
	if err == nil {
		// 缓存命中，直接返回
		return &cachedResponse, nil
	}

	// 2. 缓存未命中，使用熔断器保护数据库查询
	result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		// 构建查询
		query := s.db.Model(&models.Product{})

		// 应用过滤条件
		if req.Keyword != "" {
			query = query.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
		}

		// 只查询激活状态的商品
		query = query.Where("status = ?", "active")

		// 排除活动商品
		if req.ExcludeActivity {
			query = query.Where("is_activity = ?", 0)
		}

		// 计算总数
		var total int64
		query.Count(&total)

		// 分页
		offset := (req.Page - 1) * req.Limit
		var products []models.Product
		query.Offset(offset).Limit(req.Limit).Find(&products)

		// 转换为前端需要的格式
		var productList []ProductInfo
		for _, p := range products {
			// 查询默认SKU价格（取第一个SKU的价格）
			var defaultSku models.ProductSku
			s.db.Where("product_id = ?", p.ID).Order("id ASC").First(&defaultSku)

			defaultSkuPrice := p.Price
			if defaultSku.ID > 0 {
				defaultSkuPrice = defaultSku.Price
			}

			// 查询商品图片列表 - 优先获取is_main=1的图片，否则按sort排序取第一张
			var productImages []models.ProductImage
			s.db.Where("product_id = ?", p.ID).Order("is_main DESC, sort ASC").Find(&productImages)

			// 构建图片URL列表 - is_main优先，然后按sort
			images := make([]string, 0)
			var mainImage string

			for _, img := range productImages {
				if img.ImageURL != "" {
					images = append(images, img.ImageURL)
					if img.IsMain && mainImage == "" {
						mainImage = img.ImageURL
					}
				}
			}

			// 如果没有设置is_main，使用第一张图片作为主图
			if mainImage == "" && len(images) > 0 {
				mainImage = images[0]
			}

			productList = append(productList, ProductInfo{
				ID:              int(p.ID),
				Name:            p.Name,
				Description:     p.Description,
				Price:           p.Price,
				Stock:           p.Stock,
				Image:           mainImage,
				Images:          images,
				DefaultSkuPrice: defaultSkuPrice,
				Sales:           0, // 可以从订单统计获取
			})
		}

		response := &GetProductsResponse{
			Products: productList,
			Total:    total,
		}

		// 3. 将查询结果写入缓存，添加随机过期时间防止缓存雪崩
		physicalExpiration := cache.ProductListExpiration + time.Duration(time.Now().UnixNano()%int64(cache.ProductListExpirationOffset))
		s.multiLevelCache.Set(ctx, cacheKey, response, physicalExpiration)

		return response, nil
	})

	if err != nil {
		// 熔断器触发或其他错误，返回降级数据
		return &GetProductsResponse{
			Products: []ProductInfo{},
			Total:    0,
		}, nil
	}
	utils.Info("检查日志生成:redisCacheResult = %s", result)
	return result.(*GetProductsResponse), nil
}

// GetProductByID 根据商品ID获取商品
func (s *ProductService) GetProductByID(productID int) (*models.Product, error) {
	var product models.Product
	result := s.db.Where("id = ?", productID).First(&product)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &product, nil
}

// ReduceStock 扣减商品库存
func (s *ProductService) ReduceStock(productID, quantity int) error {
	var product models.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return err
	}

	if product.Stock < quantity {
		return nil
	}

	product.Stock -= quantity
	return s.db.Save(&product).Error
}

// IncreaseStock 增加商品库存
func (s *ProductService) IncreaseStock(productID, quantity int) error {
	var product models.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return err
	}

	product.Stock += quantity
	return s.db.Save(&product).Error
}
