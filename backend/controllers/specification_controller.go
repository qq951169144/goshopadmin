package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SpecificationController struct {
	BaseController
	specService     *services.SpecificationService
	merchantService *services.MerchantService
}

func NewSpecificationController(db *gorm.DB) *SpecificationController {
	return &SpecificationController{
		specService:     services.NewSpecificationService(db),
		merchantService: services.NewMerchantService(db),
	}
}

func (c *SpecificationController) CreateSpecification(ctx *gin.Context) {
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

func (c *SpecificationController) UpdateSpecification(ctx *gin.Context) {
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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

func (c *SpecificationController) DeleteSpecification(ctx *gin.Context) {
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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

func (c *SpecificationController) CreateSpecificationValue(ctx *gin.Context) {
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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

func (c *SpecificationController) UpdateSpecificationValue(ctx *gin.Context) {
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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

func (c *SpecificationController) DeleteSpecificationValue(ctx *gin.Context) {
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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

func (c *SpecificationController) GetSpecificationsByProductID(ctx *gin.Context) {
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

	specs, err := c.specService.GetSpecificationsByProductID(productID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, specs)
}
