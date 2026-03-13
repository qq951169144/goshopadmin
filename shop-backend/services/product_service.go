package services

import (
	"gorm.io/gorm"
	"shop-backend/models"
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
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
	Stock       int     `json:"stock"`
	Image       string  `json:"image"`
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
		productList = append(productList, ProductInfo{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			SKU:         p.SKU,
			Stock:       p.Stock,
			Image:       p.Image,
		})
	}

	return &GetProductsResponse{
		Products: productList,
		Total:    total,
	}, nil
}

// GetProductDetail 获取商品详情
func (s *ProductService) GetProductDetail(id uint) (*ProductInfo, error) {
	var product models.Product
	result := s.db.First(&product, id)
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &ProductInfo{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		SKU:         product.SKU,
		Stock:       product.Stock,
		Image:       product.Image,
	}, nil
}
