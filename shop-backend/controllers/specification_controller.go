package controllers

import (
	"strconv"

	"shop-backend/errors"
	"shop-backend/services"

	"github.com/gin-gonic/gin"
)

// SpecificationController 规格控制器
type SpecificationController struct {
	BaseController
	specService *services.SpecificationService
}

// NewSpecificationController 创建规格控制器
func NewSpecificationController(specService *services.SpecificationService) *SpecificationController {
	return &SpecificationController{
		specService: specService,
	}
}

// GetProductDetail 获取商品详情（含规格信息）
func (c *SpecificationController) GetProductDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 从服务层获取带规格信息的商品详情
	product, err := c.specService.GetProductDetailWithSpecs(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeProductNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, product)
}

// GetProductSKUs 获取商品的SKU列表
func (c *SpecificationController) GetProductSKUs(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	skus, err := c.specService.GetSKUsByProductID(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"skus": skus,
	})
}

// GetSKUBySpecCombination 根据规格组合查询SKU
func (c *SpecificationController) GetSKUBySpecCombination(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	specs := ctx.Query("specs")
	if specs == "" {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	sku, err := c.specService.GetSKUBySpecCombination(id, specs)
	if err != nil {
		c.ResponseError(ctx, errors.CodeProductNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, sku)
}
