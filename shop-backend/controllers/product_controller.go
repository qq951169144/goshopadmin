package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/config"
	"shop-backend/models"
)

// 商品结构
type Product struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
	Stock       int     `json:"stock"`
	Image       string  `json:"image"`
}

// 获取商品列表
func GetProducts(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	keyword := c.Query("keyword")

	// 构建查询
	query := config.DB.Model(&models.Product{})

	// 应用过滤条件
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页
	offset := (page - 1) * limit
	var products []models.Product
	query.Offset(offset).Limit(limit).Find(&products)

	// 转换为前端需要的格式
	var productList []Product
	for _, p := range products {
		productList = append(productList, Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			SKU:         p.SKU,
			Stock:       p.Stock,
			Image:       p.Image,
		})
	}

	ResponseSuccess(c, gin.H{
		"products": productList,
		"total":    total,
	})
}

// 获取商品详情
func GetProductDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// 从数据库获取商品详情
	var product models.Product
	result := config.DB.First(&product, id)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "Product not found")
		return
	}

	// 转换为前端需要的格式
	ResponseSuccess(c, Product{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		SKU:         product.SKU,
		Stock:       product.Stock,
		Image:       product.Image,
	})
}
