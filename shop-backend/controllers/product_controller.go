package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/services"
)

// ProductController 商品控制器
type ProductController struct {
	BaseController
	productService *services.ProductService
}

// NewProductController 创建商品控制器实例
func NewProductController(productService *services.ProductService) *ProductController {
	return &ProductController{
		productService: productService,
	}
}

// GetProducts 获取商品列表
func (c *ProductController) GetProducts(ctx *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	keyword := ctx.Query("keyword")

	// 从服务层获取商品列表
	resp, err := c.productService.GetProducts(services.GetProductsRequest{
		Page:    page,
		Limit:   limit,
		Keyword: keyword,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"products": resp.Products,
		"total":    resp.Total,
	})
}

// GetProductDetail 获取商品详情
func (c *ProductController) GetProductDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// 从服务层获取商品详情
	product, err := c.productService.GetProductDetail(uint(id))
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		c.ResponseError(ctx, http.StatusNotFound, "Product not found")
		return
	}

	c.ResponseSuccess(ctx, product)
}
