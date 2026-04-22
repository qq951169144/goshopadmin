package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SkuController struct {
	BaseController
	skuService      *services.SkuService
	merchantService *services.MerchantService
}

func NewSkuController(db *gorm.DB) *SkuController {
	return &SkuController{
		skuService:      services.NewSkuService(db),
		merchantService: services.NewMerchantService(db),
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ========== 请求结构体定义==========

type SkuSpecComboReq struct {
	SpecID      int `json:"spec_id"`
	SpecValueID int `json:"spec_value_id"`
}

type SkuCreateReq struct {
	SkuCode          string            `json:"sku_code" binding:"required"`
	Price            float64           `json:"price" binding:"required"`
	OriginalPrice    float64           `json:"original_price"`
	Stock            int               `json:"stock"`
	Status           string            `json:"status"`
	SpecCombinations []SkuSpecComboReq `json:"spec_combinations"`
}

type SkuUpdateReq struct {
	SkuCode          string            `json:"sku_code"`
	Price            float64           `json:"price"`
	OriginalPrice    float64           `json:"original_price"`
	Stock            int               `json:"stock"`
	Status           string            `json:"status"`
	SpecCombinations []SkuSpecComboReq `json:"spec_combinations"`
}

type BatchSkuCreateReqItem struct {
	SkuCode          string            `json:"sku_code"`
	Price            float64           `json:"price"`
	OriginalPrice    float64           `json:"original_price"`
	Stock            int               `json:"stock"`
	Status           string            `json:"status"`
	IsActivity       bool              `json:"is_activity"`
	ActivityID       int               `json:"activity_id"`
	SpecCombinations []SkuSpecComboReq `json:"spec_combinations"`
}

type BatchSkuCreateReq struct {
	Skus []BatchSkuCreateReqItem `json:"skus" binding:"required"`
}

type GenerateSkusReq struct {
	BasePrice float64 `json:"base_price"`
}

func (c *SkuController) CreateSku(ctx *gin.Context) {
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

	var req SkuCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	specCombinations := make([]models.ProductSkuSpec, len(req.SpecCombinations))
	for i, combo := range req.SpecCombinations {
		specCombinations[i] = models.ProductSkuSpec{
			SpecID:      combo.SpecID,
			SpecValueID: combo.SpecValueID,
		}
	}

	sku := &models.ProductSku{
		ProductID:     productID,
		MerchantID:    merchantID,
		SkuCode:       req.SkuCode,
		Price:         decimal.NewFromFloat(req.Price),
		OriginalPrice: decimal.NewFromFloat(req.OriginalPrice),
		Stock:         req.Stock,
		Status:        req.Status,
	}

	if err := c.skuService.CreateSku(sku, specCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, sku)
}

func (c *SkuController) BatchCreateSku(ctx *gin.Context) {
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

	var req BatchSkuCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	skus := make([]models.ProductSku, len(req.Skus))
	specCombinations := make([][]models.ProductSkuSpec, len(req.Skus))
	for i, item := range req.Skus {
		skus[i] = models.ProductSku{
			ProductID:     productID,
			MerchantID:    merchantID,
			SkuCode:       item.SkuCode,
			Price:         decimal.NewFromFloat(item.Price),
			OriginalPrice: decimal.NewFromFloat(item.OriginalPrice),
			Stock:         item.Stock,
			Status:        item.Status,
			IsActivity:    boolToInt(item.IsActivity),
			ActivityID:    item.ActivityID,
		}
		if item.Status == "" {
			skus[i].Status = "active"
		}

		specCombinations[i] = make([]models.ProductSkuSpec, len(item.SpecCombinations))
		for j, combo := range item.SpecCombinations {
			specCombinations[i][j] = models.ProductSkuSpec{
				SpecID:      combo.SpecID,
				SpecValueID: combo.SpecValueID,
			}
		}
	}

	if err := c.skuService.BatchCreateSku(productID, skus, specCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

func (c *SkuController) UpdateSku(ctx *gin.Context) {
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

	var req SkuUpdateReq
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

	specCombinations := make([]models.ProductSkuSpec, len(req.SpecCombinations))
	for i, combo := range req.SpecCombinations {
		specCombinations[i] = models.ProductSkuSpec{
			SpecID:      combo.SpecID,
			SpecValueID: combo.SpecValueID,
		}
	}

	if err := c.skuService.UpdateSku(skuID, updates, specCombinations, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

func (c *SkuController) DeleteSku(ctx *gin.Context) {
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

func (c *SkuController) GetSkusByProductID(ctx *gin.Context) {
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

func (c *SkuController) GenerateSkusFromSpecs(ctx *gin.Context) {
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

	var req GenerateSkusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	preview, err := c.skuService.GenerateSkusFromSpecs(productID, req.BasePrice, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, preview)
}
