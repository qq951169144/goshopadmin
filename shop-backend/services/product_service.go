package services

import (
	"context"
	"fmt"
	"time"

	"shop-backend/cache"
	"shop-backend/models"

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
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Price           float64  `json:"price"`
	SKU             string   `json:"sku"`
	Stock           int      `json:"stock"`
	Image           string   `json:"image"`
	Images          []string `json:"images"`
	DefaultSkuPrice float64  `json:"default_sku_price"`
	Sales           int      `json:"sales"`
}

// ProductDetailInfo 商品详情信息结构
type ProductDetailInfo struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Detail       string    `json:"detail"`
	Price        float64   `json:"price"`
	Image        string    `json:"image"`
	Images       []string  `json:"images"`
	SKUs         []SKUInfo `json:"skus"`
	Sales        int       `json:"sales"`
	ReviewsCount int       `json:"reviews_count"`
}

// SKUInfo SKU信息结构
type SKUInfo struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Price      float64           `json:"price"`
	Stock      int               `json:"stock"`
	Attributes map[string]string `json:"attributes"`
}

// GetProductsRequest 获取商品列表请求
type GetProductsRequest struct {
	Page    int
	Limit   int
	Keyword string
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
	cacheKey := s.cacheUtil.GetProductListCacheKey(req.Page, req.Limit, req.Keyword)

	// 1. 尝试从多级缓存获取数据
	var cachedResponse GetProductsResponse
	err := s.multiLevelCache.Get(ctx, cacheKey, &cachedResponse)
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
			var defaultSku models.ProductSKU
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

	return result.(*GetProductsResponse), nil
}

// GetProductDetail 获取商品详情
func (s *ProductService) GetProductDetail(id int) (*ProductDetailInfo, error) {
	ctx := context.Background()

	// 生成缓存键
	nullKey := fmt.Sprintf("product:null:%d", id)

	// 1. 检查空值缓存
	nullExists, err := s.cacheUtil.GetNullValue(ctx, nullKey)
	if err == nil && nullExists {
		return nil, nil
	}

	// 2. 检查布隆过滤器
	exists, err := s.cacheUtil.CheckProductExists(ctx, id)
	if err == nil && !exists {
		// 布隆过滤器判断不存在，设置空值缓存
		s.cacheUtil.SetNullValue(ctx, nullKey)
		return nil, nil
	}

	// 3. 尝试从缓存获取
	cacheData, err := s.cacheUtil.GetProductCache(ctx, id)
	if err == nil && cacheData != nil {
		// 检查缓存是否过期
		if !s.cacheUtil.IsCacheExpired(cacheData) {
			// 缓存未过期，直接返回
			if data, ok := cacheData.Data.(map[string]interface{}); ok {
				return s.convertMapToProductDetail(data), nil
			}
		} else {
			// 缓存已过期，返回旧数据并后台异步更新
			go s.refreshProductCache(ctx, id)
			if data, ok := cacheData.Data.(map[string]interface{}); ok {
				return s.convertMapToProductDetail(data), nil
			}
		}
	}

	// 4. 缓存不存在或无法解析，使用分布式锁防止缓存击穿
	lock := cache.NewDistributedLockFromCacheUtil(s.cacheUtil, s.cacheUtil.GetProductCacheKey(id)+"_lock", fmt.Sprintf("lock:%d", id))
	acquired, err := lock.TryAcquireWithTimeout(ctx, cache.LockExpiration, cache.LockTimeout)
	if err != nil {
		// 锁获取失败，尝试直接查询数据库
		return s.getProductDetailFromDB(ctx, id, nullKey)
	}

	if !acquired {
		// 锁被其他线程获取，尝试再次从缓存获取
		cacheData, err = s.cacheUtil.GetProductCache(ctx, id)
		if err == nil && cacheData != nil {
			if data, ok := cacheData.Data.(map[string]interface{}); ok {
				return s.convertMapToProductDetail(data), nil
			}
		}
		// 还是没有缓存，直接查询数据库
		return s.getProductDetailFromDB(ctx, id, nullKey)
	}

	// 5. 成功获取锁，查询数据库并重建缓存
	defer lock.Release(ctx)

	// 再次检查缓存，防止其他线程已经更新
	cacheData, err = s.cacheUtil.GetProductCache(ctx, id)
	if err == nil && cacheData != nil && !s.cacheUtil.IsCacheExpired(cacheData) {
		if data, ok := cacheData.Data.(map[string]interface{}); ok {
			return s.convertMapToProductDetail(data), nil
		}
	}

	// 6. 查询数据库
	productDetail, err := s.getProductDetailFromDB(ctx, id, nullKey)
	if err != nil || productDetail == nil {
		return productDetail, err
	}

	// 7. 重建缓存
	s.cacheUtil.SetProductCache(ctx, id, productDetail)

	return productDetail, nil
}

// refreshProductCache 后台异步更新商品缓存
func (s *ProductService) refreshProductCache(ctx context.Context, productID int) {
	// 生成缓存键
	nullKey := fmt.Sprintf("product:null:%d", productID)

	// 查询数据库
	productDetail, err := s.getProductDetailFromDB(ctx, productID, nullKey)
	if err != nil || productDetail == nil {
		return
	}

	// 更新缓存
	s.cacheUtil.SetProductCache(ctx, productID, productDetail)
}

// getProductDetailFromDB 从数据库获取商品详情
func (s *ProductService) getProductDetailFromDB(ctx context.Context, id int, nullKey string) (*ProductDetailInfo, error) {
	// 查询数据库
	var product models.Product
	result := s.db.First(&product, id)
	if result.RowsAffected == 0 {
		// 数据库也不存在，设置空值缓存
		s.cacheUtil.SetNullValue(ctx, nullKey)
		return nil, nil
	}

	// 查询SKU列表
	var skus []models.ProductSKU
	s.db.Where("product_id = ?", id).Find(&skus)

	var skuList []SKUInfo
	for _, sku := range skus {
		skuList = append(skuList, SKUInfo{
			ID:    sku.ID,
			Name:  sku.SKUCode,
			Price: sku.Price,
			Stock: sku.Stock,
		})
	}

	// 如果没有SKU，使用商品默认价格创建一个SKU
	if len(skuList) == 0 {
		skuList = append(skuList, SKUInfo{
			ID:    0,
			Name:  "默认规格",
			Price: product.Price,
			Stock: product.Stock,
		})
	}

	// 查询商品图片列表 - is_main DESC(主图优先), sort ASC(排序)
	var productImages []models.ProductImage
	s.db.Where("product_id = ?", id).Order("is_main DESC, sort ASC").Find(&productImages)

	// 构建图片URL列表 - is_main放第一位
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
	if mainImage == "" && len(images) > 0 {
		mainImage = images[0]
	}

	productDetail := &ProductDetailInfo{
		ID:           int(product.ID),
		Name:         product.Name,
		Description:  product.Description,
		Detail:       product.Detail, // 富文本详情
		Price:        product.Price,
		Image:        mainImage,
		Images:       images,
		SKUs:         skuList,
		Sales:        0,
		ReviewsCount: 0,
	}

	return productDetail, nil
}

// convertMapToProductDetail 将map转换为ProductDetailInfo
func (s *ProductService) convertMapToProductDetail(data map[string]interface{}) *ProductDetailInfo {
	productDetail := &ProductDetailInfo{}

	if id, ok := data["id"].(float64); ok {
		productDetail.ID = int(id)
	}
	if name, ok := data["name"].(string); ok {
		productDetail.Name = name
	}
	if description, ok := data["description"].(string); ok {
		productDetail.Description = description
	}
	if detail, ok := data["detail"].(string); ok {
		productDetail.Detail = detail
	}
	if price, ok := data["price"].(float64); ok {
		productDetail.Price = price
	}
	if image, ok := data["image"].(string); ok {
		productDetail.Image = image
	}
	if images, ok := data["images"].([]interface{}); ok {
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				productDetail.Images = append(productDetail.Images, imgStr)
			}
		}
	}
	if skus, ok := data["skus"].([]interface{}); ok {
		for _, sku := range skus {
			if skuMap, ok := sku.(map[string]interface{}); ok {
				skuInfo := SKUInfo{}
				if id, ok := skuMap["id"].(float64); ok {
					skuInfo.ID = int(id)
				}
				if name, ok := skuMap["name"].(string); ok {
					skuInfo.Name = name
				}
				if price, ok := skuMap["price"].(float64); ok {
					skuInfo.Price = price
				}
				if stock, ok := skuMap["stock"].(float64); ok {
					skuInfo.Stock = int(stock)
				}
				productDetail.SKUs = append(productDetail.SKUs, skuInfo)
			}
		}
	}
	if sales, ok := data["sales"].(float64); ok {
		productDetail.Sales = int(sales)
	}
	if reviewsCount, ok := data["reviews_count"].(float64); ok {
		productDetail.ReviewsCount = int(reviewsCount)
	}

	return productDetail
}
