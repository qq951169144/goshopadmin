package controllers

import (
	"net/http"
	"strconv"

	"goshopadmin/models"
	"goshopadmin/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductController 商品控制器
type ProductController struct {
	productService *services.ProductService
}

// NewProductController 创建商品控制器实例
func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{
		productService: services.NewProductService(db),
	}
}

// GetProducts 获取商品列表
// @Summary 获取商品列表
// @Description 获取当前商户的商品列表
// @Tags 商品管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/products [get]
func (c *ProductController) GetProducts(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取商品列表
	products, err := c.productService.GetProductsByMerchantID(merchantID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商品列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取商品列表成功", "data": products})
}

// GetProduct 获取商品详情
// @Summary 获取商品详情
// @Description 根据商品ID获取商品详情
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [get]
func (c *ProductController) GetProduct(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取商品ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	// 获取商品详情
	product, err := c.productService.GetProductByID(id, merchantID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取商品详情成功", "data": product})
}

// CreateProduct 创建商品
// @Summary 创建商品
// @Description 创建新商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param product body models.Product true "商品信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/products [post]
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 绑定请求体
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 创建商品
	if err := c.productService.CreateProduct(&product, merchantID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建商品失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建商品成功", "data": product})
}

// UpdateProduct 更新商品
// @Summary 更新商品
// @Description 更新商品信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Param product body models.Product true "商品信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [put]
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取商品ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	// 绑定请求体
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 设置商品ID
	product.ID = id

	// 更新商品
	if err := c.productService.UpdateProduct(&product, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新商品成功", "data": product})
}

// DeleteProduct 删除商品
// @Summary 删除商品
// @Description 删除商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [delete]
func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取商品ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的商品ID"})
		return
	}

	// 删除商品
	if err := c.productService.DeleteProduct(id, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除商品成功"})
}

// GetCategories 获取商品分类列表
// @Summary 获取商品分类列表
// @Description 获取当前商户的商品分类列表
// @Tags 商品管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories [get]
func (c *ProductController) GetCategories(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取分类列表
	categories, err := c.productService.GetCategoriesByMerchantID(merchantID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取分类列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取分类列表成功", "data": categories})
}

// GetCategory 获取商品分类详情
// @Summary 获取商品分类详情
// @Description 根据分类ID获取分类详情
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories/{id} [get]
func (c *ProductController) GetCategory(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取分类ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的分类ID"})
		return
	}

	// 获取分类详情
	category, err := c.productService.GetCategoryByID(id, merchantID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "分类不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取分类详情成功", "data": category})
}

// CreateCategory 创建商品分类
// @Summary 创建商品分类
// @Description 创建新商品分类
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param category body models.ProductCategory true "分类信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories [post]
func (c *ProductController) CreateCategory(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 绑定请求体
	var category models.ProductCategory
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 创建分类
	if err := c.productService.CreateCategory(&category, merchantID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建分类失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建分类成功", "data": category})
}

// UpdateCategory 更新商品分类
// @Summary 更新商品分类
// @Description 更新商品分类信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Param category body models.ProductCategory true "分类信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories/{id} [put]
func (c *ProductController) UpdateCategory(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取分类ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的分类ID"})
		return
	}

	// 绑定请求体
	var category models.ProductCategory
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 设置分类ID
	category.ID = id

	// 更新分类
	if err := c.productService.UpdateCategory(&category, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新分类成功", "data": category})
}

// DeleteCategory 删除商品分类
// @Summary 删除商品分类
// @Description 删除商品分类
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories/{id} [delete]
func (c *ProductController) DeleteCategory(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取分类ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的分类ID"})
		return
	}

	// 删除分类
	if err := c.productService.DeleteCategory(id, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除分类成功"})
}

// AddProductImage 添加商品图片
// @Summary 添加商品图片
// @Description 为商品添加图片
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param image body models.ProductImage true "图片信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-images [post]
func (c *ProductController) AddProductImage(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 绑定请求体
	var image models.ProductImage
	if err := ctx.ShouldBindJSON(&image); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 添加图片
	if err := c.productService.AddProductImage(&image, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "添加图片成功", "data": image})
}

// DeleteProductImage 删除商品图片
// @Summary 删除商品图片
// @Description 删除商品图片
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "图片ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-images/{id} [delete]
func (c *ProductController) DeleteProductImage(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取图片ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的图片ID"})
		return
	}

	// 删除图片
	if err := c.productService.DeleteProductImage(id, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除图片成功"})
}

// AddProductSKU 添加商品SKU
// @Summary 添加商品SKU
// @Description 为商品添加SKU
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param sku body models.ProductSKU true "SKU信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-skus [post]
func (c *ProductController) AddProductSKU(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 绑定请求体
	var sku models.ProductSKU
	if err := ctx.ShouldBindJSON(&sku); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 添加SKU
	if err := c.productService.AddProductSKU(&sku, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "添加SKU成功", "data": sku})
}

// UpdateProductSKU 更新商品SKU
// @Summary 更新商品SKU
// @Description 更新商品SKU信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "SKU ID"
// @Param sku body models.ProductSKU true "SKU信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-skus/{id} [put]
func (c *ProductController) UpdateProductSKU(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取SKU ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的SKU ID"})
		return
	}

	// 绑定请求体
	var sku models.ProductSKU
	if err := ctx.ShouldBindJSON(&sku); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
		return
	}

	// 设置SKU ID
	sku.ID = id

	// 更新SKU
	if err := c.productService.UpdateProductSKU(&sku, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新SKU成功", "data": sku})
}

// DeleteProductSKU 删除商品SKU
// @Summary 删除商品SKU
// @Description 删除商品SKU
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "SKU ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-skus/{id} [delete]
func (c *ProductController) DeleteProductSKU(ctx *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 获取SKU ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的SKU ID"})
		return
	}

	// 删除SKU
	if err := c.productService.DeleteProductSKU(id, merchantID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除SKU成功"})
}
