package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SKUController SKU控制器
type SKUController struct {
	BaseController
	skuService      *services.SKUService
	merchantService *services.MerchantService
}

// NewSKUController 创建SKU控制器
func NewSKUController(db *gorm.DB) *SKUController {
	return &SKUController{
		skuService:      services.NewSKUService(db),
		merchantService: services.NewMerchantService(db),
	}
}

// CreateSKU 创建单个SKU
func (c *SKUController) CreateSKU(ctx *gin.Context) {
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
		SKUCode          string                  `json:"sku_code" binding:"required"`
		Price            float64                 `json:"price" binding:"required"`
		OriginalPrice    float64                 `json:"original_price"`
		Stock            int                     `json:"stock"`
		Status           string                  `json:"status"`
		SpecCombinations []models.ProductSKUSpec `json:"spec_combinations"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	sku := &models.ProductSKU{
		ProductID:     productID,
		SKUCode:       req.SKUCode,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Stock:         req.Stock,
		Status:        req.Status,
	}

	if err := c.skuService.CreateSKU(sku, req.SpecCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, sku)
}

// BatchCreateSKU 批量创建SKU
func (c *SKUController) BatchCreateSKU(ctx *gin.Context) {
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
		SKUs []struct {
			models.ProductSKU
			SpecCombinations []models.ProductSKUSpec `json:"spec_combinations"`
		} `json:"skus" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	// 转换数据格式
	skus := make([]models.ProductSKU, len(req.SKUs))
	specCombinations := make([][]models.ProductSKUSpec, len(req.SKUs))
	for i, sku := range req.SKUs {
		skus[i] = sku.ProductSKU
		specCombinations[i] = sku.SpecCombinations
	}

	if err := c.skuService.BatchCreateSKU(productID, skus, specCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// UpdateSKU 更新SKU
func (c *SKUController) UpdateSKU(ctx *gin.Context) {
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
		SKUCode          string                  `json:"sku_code"`
		Price            float64                 `json:"price"`
		OriginalPrice    float64                 `json:"original_price"`
		Stock            int                     `json:"stock"`
		Status           string                  `json:"status"`
		SpecCombinations []models.ProductSKUSpec `json:"spec_combinations"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	updates := make(map[string]interface{})
	if req.SKUCode != "" {
		updates["sku_code"] = req.SKUCode
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	updates["original_price"] = req.OriginalPrice
	updates["stock"] = req.Stock
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := c.skuService.UpdateSKU(skuID, updates, req.SpecCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// DeleteSKU 删除SKU
func (c *SKUController) DeleteSKU(ctx *gin.Context) {
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

	if err := c.skuService.DeleteSKU(skuID, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// GetSKUsByProductID 获取商品的SKU列表
func (c *SKUController) GetSKUsByProductID(ctx *gin.Context) {
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

	skus, err := c.skuService.GetSKUsByProductID(productID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, skus)
}

// GenerateSKUsFromSpecs 根据规格组合自动生成SKU
func (c *SKUController) GenerateSKUsFromSpecs(ctx *gin.Context) {
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

	skus, err := c.skuService.GenerateSKUsFromSpecs(productID, req.BasePrice, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, skus)
}
