package controllers

import (
	stderrors "errors"
	"fmt"
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"goshopadmin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductController 商品控制器
type ProductController struct {
	BaseController
	productService *services.ProductService
}

// getMerchantIDFromContext 从上下文获取商户ID（私有方法）
func (c *ProductController) getMerchantIDFromContext(ctx *gin.Context) (int, error) {
	// 从上下文获取用户ID
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return 0, stderrors.New("未授权")
	}

	// 获取商户ID
	merchantID, err := c.productService.GetMerchantIDByUserID(userID)
	if err != nil {
		return 0, err
	}

	return merchantID, nil
}

// CreateProductRequest 创建商品请求

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Detail      string  `json:"detail"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  int     `json:"category_id" binding:"required"`
	Status      string  `json:"status"`
}

// UpdateProductRequest 更新商品请求

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Detail      string  `json:"detail"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  int     `json:"category_id"`
	Status      string  `json:"status"`
}

// CreateCategoryRequest 创建分类请求

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID int    `json:"parent_id"`
	Level    int    `json:"level"`
	Sort     int    `json:"sort"`
	Status   string `json:"status"`
}

// UpdateCategoryRequest 更新分类请求

type UpdateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID int    `json:"parent_id"`
	Level    int    `json:"level"`
	Sort     int    `json:"sort"`
	Status   string `json:"status"`
}

// AddProductImageRequest 添加商品图片请求

type AddProductImageRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	ImageURL  string `json:"image_url" binding:"required"`
	IsMain    bool   `json:"is_main"`
	Sort      int    `json:"sort"`
}

// UpdateProductImageRequest 更新商品图片请求

type UpdateProductImageRequest struct {
	ProductID int    `json:"product_id" binding:"required"`
	ImageURL  string `json:"image_url"`
	IsMain    bool   `json:"is_main"`
	Sort      int    `json:"sort"`
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取商品列表
	products, err := c.productService.GetProductsByMerchantID(merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, products)
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取商品ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取商品详情
	product, err := c.productService.GetProductByID(id, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeProductNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, product)
}

// CreateProduct 创建商品
// @Summary 创建商品
// @Description 创建新商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "商品信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/products [post]
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 绑定请求体
	var req CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 预处理数据
	product := models.Product{
		MerchantID:  merchantID,
		Name:        req.Name,
		Description: req.Description,
		Detail:      req.Detail,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		Status:      req.Status,
	}

	// 创建商品
	if err := c.productService.CreateProduct(&product, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, product)
}

// UpdateProduct 更新商品
// @Summary 更新商品
// @Description 更新商品信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Param product body UpdateProductRequest true "商品信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [put]
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取商品ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 绑定请求体
	var req UpdateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 构建更新字段map，只包含请求中传递的字段
	updateData := make(map[string]interface{})
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Detail != "" {
		updateData["detail"] = req.Detail
	}
	if req.CategoryID > 0 {
		updateData["category_id"] = req.CategoryID
	}
	if req.Status != "" {
		updateData["status"] = req.Status
	}

	// 更新商品
	if err := c.productService.UpdateProduct(id, merchantID, updateData); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"id": id})
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取商品ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 删除商品
	if err := c.productService.DeleteProduct(id, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取分类列表
	categories, err := c.productService.GetCategoriesByMerchantID(merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, categories)
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取分类ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取分类详情
	category, err := c.productService.GetCategoryByID(id, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, category)
}

// CreateCategory 创建商品分类
// @Summary 创建商品分类
// @Description 创建新商品分类
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param category body CreateCategoryRequest true "分类信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories [post]
func (c *ProductController) CreateCategory(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 绑定请求体
	var req CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 预处理数据
	category := models.ProductCategory{
		MerchantID: merchantID,
		Name:       req.Name,
		ParentID:   req.ParentID,
		Level:      req.Level,
		Sort:       req.Sort,
		Status:     req.Status,
	}

	// 创建分类
	if err := c.productService.CreateCategory(&category, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, category)
}

// UpdateCategory 更新商品分类
// @Summary 更新商品分类
// @Description 更新商品分类信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Param category body UpdateCategoryRequest true "分类信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-categories/{id} [put]
func (c *ProductController) UpdateCategory(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取分类ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 绑定请求体
	var req UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 预处理数据
	category := models.ProductCategory{
		ID:         id,
		MerchantID: merchantID,
		Name:       req.Name,
		ParentID:   req.ParentID,
		Level:      req.Level,
		Sort:       req.Sort,
		Status:     req.Status,
	}

	// 更新分类
	if err := c.productService.UpdateCategory(&category, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, category)
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取分类ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 删除分类
	if err := c.productService.DeleteCategory(id, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// AddProductImage 添加商品图片
// @Summary 添加商品图片
// @Description 为商品添加图片
// @Tags 商品管理
// @Accept multipart/form-data
// @Produce json
// @Param product_id formData int true "商品ID"
// @Param image formData file true "图片文件"
// @Param is_main formData bool false "是否主图"
// @Param sort formData int false "排序"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-images [post]
func (c *ProductController) AddProductImage(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取商品ID
	productID, err := strconv.Atoi(ctx.PostForm("product_id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, fmt.Errorf("商品ID格式错误: %w", err))
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("image")
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, fmt.Errorf("获取上传文件失败: %w", err))
		return
	}

	// 上传图片到存储
	imageURL, err := utils.UploadImage(file, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取其他参数
	isMain := ctx.PostForm("is_main") == "true"
	sort, _ := strconv.Atoi(ctx.PostForm("sort"))

	// 预处理数据
	image := models.ProductImage{
		ProductID: productID,
		ImageURL:  imageURL,
		IsMain:    isMain,
		Sort:      sort,
	}

	// 添加图片
	if err := c.productService.AddProductImage(&image, merchantID); err != nil {
		// 如果数据库操作失败，删除已上传的图片
		utils.DeleteImage(imageURL)
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, image)
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
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取图片ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 删除图片
	if err := c.productService.DeleteProductImage(id, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// UpdateProductImage 更新商品图片
// @Summary 更新商品图片
// @Description 更新商品图片信息（排序、主图设置）
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "图片ID"
// @Param image body UpdateProductImageRequest true "图片信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/product-images/{id} [put]
func (c *ProductController) UpdateProductImage(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.getMerchantIDFromContext(ctx)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	// 获取图片ID
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 绑定请求体
	var req UpdateProductImageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 预处理数据
	image := models.ProductImage{
		ID:        id,
		ProductID: req.ProductID,
		ImageURL:  req.ImageURL,
		IsMain:    req.IsMain,
		Sort:      req.Sort,
	}

	// 更新图片
	if err := c.productService.UpdateProductImage(&image, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, image)
}
