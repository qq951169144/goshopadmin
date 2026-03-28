package controllers

import (
	"goshopadmin/constants"
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityController 活动控制器
type ActivityController struct {
	BaseController
	activityService   *services.ActivityService
	redeemCodeService *services.RedeemCodeService
	skuService        *services.SKUService
	DB                *gorm.DB
}

func NewActivityController(db *gorm.DB) *ActivityController {
	return &ActivityController{
		activityService:   services.NewActivityService(db),
		redeemCodeService: services.NewRedeemCodeService(db),
		skuService:        services.NewSKUService(db),
		DB:                db,
	}
}

// CreateActivityProductRequest 创建活动商品请求结构体
type CreateActivityProductRequest struct {
	ProductID     int     `json:"product_id" binding:"required"`
	SKUID         int     `json:"sku_id" binding:"required"`
	OriginalPrice float64 `json:"original_price" binding:"required"`
	ActivityPrice float64 `json:"activity_price" binding:"required"`
	Stock         int     `json:"stock" binding:"required"`
	ProductType   string  `json:"product_type"`
}

// CreateActivityRedeemSettingRequest 创建活动兑换码设置请求结构体
type CreateActivityRedeemSettingRequest struct {
	CodeType      string `json:"code_type"`
	CodeLength    int    `json:"code_length"`
	ExcludeChars  string `json:"exclude_chars"`
	TotalQuantity int    `json:"total_quantity"`
	LimitPerUser  int    `json:"limit_per_user"`
	NeedVerify    int    `json:"need_verify"`
}

// CreateActivityRequest 创建活动请求结构体
type CreateActivityRequest struct {
	Name          string                              `json:"name" binding:"required"`
	Type          string                              `json:"type" binding:"required"`
	StartTime     string                              `json:"start_time" binding:"required"`
	EndTime       string                              `json:"end_time" binding:"required"`
	Status        string                              `json:"status"`
	Products      []CreateActivityProductRequest      `json:"products"`
	RedeemSetting *CreateActivityRedeemSettingRequest `json:"redeem_setting"`
}

// CreateActivity 创建活动
// @Summary 创建活动
// @Description 创建新活动
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param activity body CreateActivityRequest true "活动信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities [post]
func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	createdBy, _ := ctx.Get("user_id")
	var req CreateActivityRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	startTime, err := time.Parse(time.DateTime, req.StartTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	endTime, err := time.Parse(time.DateTime, req.EndTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	activity := &models.Activity{
		MerchantID: merchantID,
		Name:       req.Name,
		Type:       req.Type,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     req.Status,
		CreatedBy:  createdBy.(int),
	}

	if activity.Status == "" {
		activity.Status = constants.StatusActive
	}

	var activityProducts []models.ActivityProduct
	for _, p := range req.Products {
		productType := p.ProductType
		if productType == "" {
			productType = "seckill"
		}

		activityProducts = append(activityProducts, models.ActivityProduct{
			ActivityID:    activity.ID,
			ProductID:     p.ProductID,
			SKUID:         p.SKUID,
			MerchantID:    merchantID,
			OriginalPrice: p.OriginalPrice,
			ActivityPrice: p.ActivityPrice,
			Stock:         p.Stock,
			ProductType:   productType,
			Status:        constants.StatusActive,
		})
	}

	var redeemSetting *models.ActivityRedeemSetting
	if activity.Type == constants.ActivityTypeRedeemCode && req.RedeemSetting != nil {
		redeemSetting = &models.ActivityRedeemSetting{
			ActivityID:    activity.ID,
			MerchantID:    merchantID,
			CodeType:      req.RedeemSetting.CodeType,
			CodeLength:    req.RedeemSetting.CodeLength,
			ExcludeChars:  req.RedeemSetting.ExcludeChars,
			TotalQuantity: req.RedeemSetting.TotalQuantity,
			LimitPerUser:  req.RedeemSetting.LimitPerUser,
			NeedVerify:    req.RedeemSetting.NeedVerify,
		}
	}

	err = c.activityService.CreateActivity(activity, activityProducts, redeemSetting)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动创建成功"})
}

// UpdateActivityRequest 更新活动请求结构体
type UpdateActivityRequest struct {
	Name      string `json:"name" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

// UpdateActivity 更新活动
// @Summary 更新活动
// @Description 更新活动信息
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Param activity body UpdateActivityRequest true "活动信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id} [put]
func (c *ActivityController) UpdateActivity(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req UpdateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	activity, err := c.activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	startTime, err := time.Parse(time.DateTime, req.StartTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}
	activity.StartTime = startTime

	endTime, err := time.Parse(time.DateTime, req.EndTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}
	activity.EndTime = endTime

	activity.Name = req.Name
	activity.Status = req.Status

	err = c.activityService.UpdateActivity(activity, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动更新成功"})
}

// DeleteActivity 删除活动
// @Summary 删除活动
// @Description 删除活动
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id} [delete]
func (c *ActivityController) DeleteActivity(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	err = c.activityService.DeleteActivity(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动删除成功"})
}

// GetActivity 获取活动详情
// @Summary 获取活动详情
// @Description 根据活动ID获取活动详情
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id} [get]
func (c *ActivityController) GetActivity(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	activity, err := c.activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, activity)
}

// GetActivities 获取活动列表
// @Summary 获取活动列表
// @Description 获取当前商户的活动列表
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页大小"
// @Param type query string false "活动类型"
// @Param status query string false "活动状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities [get]
func (c *ActivityController) GetActivities(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	activityType := ctx.Query("type")
	status := ctx.Query("status")

	activities, total, err := c.activityService.GetActivities(merchantID, activityType, status, page, pageSize)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	var activityList []map[string]interface{}
	for _, activity := range activities {
		activityList = append(activityList, map[string]interface{}{
			"id":            activity.ID,
			"name":          activity.Name,
			"type":          activity.Type,
			"start_time":    activity.StartTime,
			"end_time":      activity.EndTime,
			"status":        activity.Status,
			"product_count": len(activity.Products),
			"created_at":    activity.CreatedAt,
		})
	}

	c.ResponseSuccess(ctx, gin.H{
		"list":  activityList,
		"total": total,
	})
}

// UpdateActivityStatusRequest 更新活动状态请求结构体
type UpdateActivityStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateActivityStatus 更新活动状态
// @Summary 更新活动状态
// @Description 更新活动状态
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Param status body UpdateActivityStatusRequest true "活动状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id}/status [put]
func (c *ActivityController) UpdateActivityStatus(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req UpdateActivityStatusRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	err = c.activityService.UpdateActivityStatus(activityID, req.Status, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动状态更新成功"})
}
