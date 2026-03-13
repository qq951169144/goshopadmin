package services

import (
	"shop-backend/models"

	"gorm.io/gorm"
)

// ProductService 商品服务
type ProductService struct {
	db *gorm.DB
}

// NewProductService 创建商品服务实例
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// ProductInfo 商品信息结构
type ProductInfo struct {
	ID              uint     `json:"id"`
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
	ID           uint      `json:"id"`
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
	ID         uint              `json:"id"`
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

		// 查询商品图片列表 - 优先获取is_main=1的图片，否则按sort_order排序取第一张
		var productImages []models.ProductImage
		s.db.Where("product_id = ?", p.ID).Order("is_main DESC, sort_order ASC").Find(&productImages)

		// 构建图片URL列表 - is_main优先，然后按sort_order
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

		// 如果没有设置is_main，使用第一张图片作为主图
		if mainImage == "" && len(images) > 0 {
			mainImage = images[0]
		}

		// 如果product_images没有图片，使用products表的image字段
		if mainImage == "" && p.Image != "" {
			mainImage = p.Image
			images = append([]string{p.Image}, images...)
		}

		productList = append(productList, ProductInfo{
			ID:              p.ID,
			Name:            p.Name,
			Description:     p.Description,
			Price:           p.Price,
			SKU:             p.SKU,
			Stock:           p.Stock,
			Image:           mainImage,
			Images:          images,
			DefaultSkuPrice: defaultSkuPrice,
			Sales:           0, // 可以从订单统计获取
		})
	}

	return &GetProductsResponse{
		Products: productList,
		Total:    total,
	}, nil
}

// GetProductDetail 获取商品详情
func (s *ProductService) GetProductDetail(id int) (*ProductDetailInfo, error) {
	var product models.Product
	result := s.db.First(&product, id)
	if result.RowsAffected == 0 {
		return nil, nil
	}

	// 查询SKU列表
	var skus []models.ProductSKU
	s.db.Where("product_id = ?", id).Find(&skus)

	var skuList []SKUInfo
	for _, sku := range skus {
		skuList = append(skuList, SKUInfo{
			ID:    sku.ID,
			Name:  sku.Name,
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

	// 查询商品图片列表 - is_main DESC(主图优先), sort_order ASC(排序)
	var productImages []models.ProductImage
	s.db.Where("product_id = ?", id).Order("is_main DESC, sort_order ASC").Find(&productImages)

	// 构建图片URL列表 - is_main放第一位
	var images []string
	for _, img := range productImages {
		if img.ImageURL != "" {
			images = append(images, img.ImageURL)
		}
	}

	// 如果product_images没有图片，使用products表的image字段
	if len(images) == 0 && product.Image != "" {
		images = append(images, product.Image)
	}

	return &ProductDetailInfo{
		ID:           product.ID,
		Name:         product.Name,
		Description:  product.Description,
		Detail:       product.Detail, // 富文本详情
		Price:        product.Price,
		Image:        product.Image,
		Images:       images,
		SKUs:         skuList,
		Sales:        0,
		ReviewsCount: 0,
	}, nil
}
