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
	skuService        *services.SkuService
	merchantService   *services.MerchantService
	DB                *gorm.DB
}

// ActivityProductResponse 活动商品响应结构体
type ActivityProductResponse struct {
	ProductID     int     `json:"product_id"`
	ProductName   string  `json:"product_name"`
	SkuID         int     `json:"sku_id"`
	SkuCode       string  `json:"sku_code"`
	ActivityPrice float64 `json:"activity_price"`
	ActivityStock int     `json:"activity_stock"`
}

// RedeemCodeRulesResponse 兑换码规则响应结构体
type RedeemCodeRulesResponse struct {
	Type         string `json:"type"`
	Length       int    `json:"length"`
	ExcludeChars string `json:"exclude_chars"`
}

// ActivityDetailResponse 活动详情响应结构体
type ActivityDetailResponse struct {
	ID              int                       `json:"id"`
	Name            string                    `json:"name"`
	Type            string                    `json:"type"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         time.Time                 `json:"end_time"`
	Status          string                    `json:"status"`
	Products        []ActivityProductResponse `json:"products"`
	RedeemCodeRules *RedeemCodeRulesResponse  `json:"redeem_code_rules,omitempty"`
}

func NewActivityController(db *gorm.DB) *ActivityController {
	return &ActivityController{
		activityService:   services.NewActivityService(db),
		redeemCodeService: services.NewRedeemCodeService(db),
		skuService:        services.NewSkuService(db),
		merchantService:   services.NewMerchantService(db),
		DB:                db,
	}
}

// CreateActivityProductRequest 创建活动商品请求结构体
type CreateActivityProductRequest struct {
	ProductID     int     `json:"product_id" binding:"required"`
	SkuID         int     `json:"sku_id" binding:"required"`
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
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}
	// c.GetMerchantIDFromContext已经验证过了,可以直接取用
	createdBy, _ := ctx.Get("userID")

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
		// 修复 product_type 字段值，确保与数据库枚举类型匹配
		if activity.Type == constants.ActivityTypeRedeemCode {
			productType = "redeem"
		}

		activityProducts = append(activityProducts, models.ActivityProduct{
			ActivityID:    activity.ID,
			ProductID:     p.ProductID,
			SkuID:         p.SkuID,
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
	Name          string                              `json:"name" binding:"required"`
	StartTime     string                              `json:"start_time" binding:"required"`
	EndTime       string                              `json:"end_time" binding:"required"`
	Status        string                              `json:"status" binding:"required"`
	Products      []CreateActivityProductRequest      `json:"products"`
	RedeemSetting *CreateActivityRedeemSettingRequest `json:"redeem_setting"`
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
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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

	// 构建活动商品列表
	var activityProducts []models.ActivityProduct
	for _, p := range req.Products {
		productType := p.ProductType
		if productType == "" {
			productType = "seckill"
		}
		// 修复 product_type 字段值，确保与数据库枚举类型匹配
		if activity.Type == constants.ActivityTypeRedeemCode {
			productType = "redeem"
		}

		activityProducts = append(activityProducts, models.ActivityProduct{
			ActivityID:    activity.ID,
			ProductID:     p.ProductID,
			SkuID:         p.SkuID,
			MerchantID:    merchantID,
			OriginalPrice: p.OriginalPrice,
			ActivityPrice: p.ActivityPrice,
			Stock:         p.Stock,
			ProductType:   productType,
			Status:        constants.StatusActive,
		})
	}

	// 构建兑换码设置
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

	err = c.activityService.UpdateActivity(activity, activityProducts, redeemSetting, merchantID)
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
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// TODO:精简返回值
	activity, err := c.activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	// 构建响应数据
	response := ActivityDetailResponse{
		ID:        activity.ID,
		Name:      activity.Name,
		Type:      activity.Type,
		StartTime: activity.StartTime,
		EndTime:   activity.EndTime,
		Status:    activity.Status,
		Products:  make([]ActivityProductResponse, 0),
	}

	// 处理关联商品
	for _, product := range activity.Products {
		productResponse := ActivityProductResponse{
			ProductID:     product.ProductID,
			ProductName:   "",
			SkuID:         product.SkuID,
			SkuCode:       "",
			ActivityPrice: product.ActivityPrice,
			ActivityStock: product.Stock,
		}

		// 获取商品名称
		if product.Product.ID != 0 {
			productResponse.ProductName = product.Product.Name
		}

		// 获取SKU编码
		if product.Sku.ID != 0 {
			productResponse.SkuCode = product.Sku.SkuCode
		}

		response.Products = append(response.Products, productResponse)
	}

	// 处理兑换码规则
	if activity.Type == "redeem_code" && activity.RedeemSetting != nil {
		response.RedeemCodeRules = &RedeemCodeRulesResponse{
			Type:         activity.RedeemSetting.CodeType,
			Length:       activity.RedeemSetting.CodeLength,
			ExcludeChars: activity.RedeemSetting.ExcludeChars,
		}
	}

	c.ResponseSuccess(ctx, response)
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
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
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
