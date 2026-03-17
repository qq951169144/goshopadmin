package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"

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

// CreateSpecification 创建商品规格
func (c *SpecificationController) CreateSpecification(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	spec := &models.ProductSpecification{
		ProductID: productID,
		Name:      req.Name,
		Sort:      req.Sort,
	}

	if err := c.specService.CreateSpecification(spec, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, spec)
}

// UpdateSpecification 更新规格
func (c *SpecificationController) UpdateSpecification(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	specID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if err := c.specService.UpdateSpecification(specID, merchantID, req.Name, req.Sort); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// DeleteSpecification 删除规格
func (c *SpecificationController) DeleteSpecification(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	specID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if err := c.specService.DeleteSpecification(specID, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// CreateSpecificationValue 创建规格值
func (c *SpecificationController) CreateSpecificationValue(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	specID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		Value string `json:"value" binding:"required"`
		Image string `json:"image"`
		Sort  int    `json:"sort"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	value := &models.ProductSpecificationValue{
		SpecID: specID,
		Value:  req.Value,
		Sort:   req.Sort,
		Status: "active",
	}

	if err := c.specService.CreateSpecificationValue(value, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, value)
}

// UpdateSpecificationValue 更新规格值
func (c *SpecificationController) UpdateSpecificationValue(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	valueID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	var req struct {
		Value  string `json:"value" binding:"required"`
		Image  string `json:"image"`
		Sort   int    `json:"sort"`
		Status string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	if err := c.specService.UpdateSpecificationValue(valueID, merchantID, req.Value, req.Image, req.Sort, req.Status); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// DeleteSpecificationValue 删除规格值
func (c *SpecificationController) DeleteSpecificationValue(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	valueID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	if err := c.specService.DeleteSpecificationValue(valueID, merchantID); err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// GetSpecificationsByProductID 获取商品的规格列表
func (c *SpecificationController) GetSpecificationsByProductID(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	// 获取商户ID
	merchantID, err := c.getMerchantIDByUserID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeForbidden, err)
		return
	}

	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	specs, err := c.specService.GetSpecificationsByProductID(productID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, specs)
}

// getMerchantIDByUserID 根据用户ID获取商户ID
func (c *SpecificationController) getMerchantIDByUserID(userID int) (int, error) {
	// 这里需要调用merchantService来获取商户ID
	// 暂时返回一个模拟值，实际应该从service获取
	return 1, nil
}
