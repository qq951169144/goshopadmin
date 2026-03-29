package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SkuController SKU控制器
type SkuController struct {
	BaseController
	skuService      *services.SkuService
	merchantService *services.MerchantService
}

// NewSkuController 创建SKU控制器
func NewSkuController(db *gorm.DB) *SkuController {
	return &SkuController{
		skuService:      services.NewSkuService(db),
		merchantService: services.NewMerchantService(db),
	}
}

// CreateSku 创建单个SKU
func (c *SkuController) CreateSku(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		SkuCode          string                  `json:"sku_code" binding:"required"`
		Price            float64                 `json:"price" binding:"required"`
		OriginalPrice    float64                 `json:"original_price"`
		Stock            int                     `json:"stock"`
		Status           string                  `json:"status"`
		SpecCombinations []models.ProductSkuSpec `json:"spec_combinations"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	sku := &models.ProductSku{
		ProductID:     productID,
		SkuCode:       req.SkuCode,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Stock:         req.Stock,
		Status:        req.Status,
	}

	if err := c.skuService.CreateSku(sku, req.SpecCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, sku)
}

// BatchCreateSku 批量创建SKU
func (c *SkuController) BatchCreateSku(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		Skus []struct {
			models.ProductSku
			SpecCombinations []models.ProductSkuSpec `json:"spec_combinations"`
		} `json:"skus" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	// 转换数据格式
	skus := make([]models.ProductSku, len(req.Skus))
	specCombinations := make([][]models.ProductSkuSpec, len(req.Skus))
	for i, sku := range req.Skus {
		skus[i] = sku.ProductSku
		specCombinations[i] = sku.SpecCombinations
	}

	if err := c.skuService.BatchCreateSku(productID, skus, specCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// UpdateSku 更新SKU
func (c *SkuController) UpdateSku(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	skuID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		SkuCode          string                  `json:"sku_code"`
		Price            float64                 `json:"price"`
		OriginalPrice    float64                 `json:"original_price"`
		Stock            int                     `json:"stock"`
		Status           string                  `json:"status"`
		SpecCombinations []models.ProductSkuSpec `json:"spec_combinations"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	updates := make(map[string]interface{})
	if req.SkuCode != "" {
		updates["sku_code"] = req.SkuCode
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	updates["original_price"] = req.OriginalPrice
	updates["stock"] = req.Stock
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := c.skuService.UpdateSku(skuID, updates, req.SpecCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// DeleteSku 删除SKU
func (c *SkuController) DeleteSku(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	skuID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if err := c.skuService.DeleteSku(skuID, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// GetSkusByProductID 获取商品的SKU列表
func (c *SkuController) GetSkusByProductID(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	skus, err := c.skuService.GetSkusByProductID(productID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, skus)
}

// GenerateSkusFromSpecs 根据规格组合自动生成SKU
func (c *SkuController) GenerateSkusFromSpecs(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		BasePrice float64 `json:"base_price"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	skus, err := c.skuService.GenerateSkusFromSpecs(productID, req.BasePrice, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, skus)
}
