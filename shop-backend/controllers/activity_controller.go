package controllers

import (
	"shop-backend/errors"
	"shop-backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityController 活动控制器
type ActivityController struct {
	BaseController
	activityService *services.ActivityService
	DB              *gorm.DB
}

// NewActivityController 创建活动控制器实例
func NewActivityController(db *gorm.DB) *ActivityController {
	return &ActivityController{
		activityService: services.NewActivityService(db),
		DB:              db,
	}
}

// GetActiveActivities 获取当前有效的活动列表
func (c *ActivityController) GetActiveActivities(ctx *gin.Context) {
	activities, err := c.activityService.GetActiveActivities()
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, activities)
}

// GetActivity 获取活动详情
func (c *ActivityController) GetActivity(ctx *gin.Context) {
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	activity, err := c.activityService.GetActivityByID(activityID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, activity)
}

// GetActivityProducts 获取活动商品列表
func (c *ActivityController) GetActivityProducts(ctx *gin.Context) {
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	products, err := c.activityService.GetActivityProducts(activityID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, products)
}

// GetActivityProductSkus 获取活动商品的SKU列表
func (c *ActivityController) GetActivityProductSkus(ctx *gin.Context) {
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	skus, err := c.activityService.GetActivityProductSkus(activityID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, skus)
}
