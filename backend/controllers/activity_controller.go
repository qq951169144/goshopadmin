package controllers

import (
	"goshopadmin/constants"
	"goshopadmin/errors"
	"goshopadmin/models"
	"goshopadmin/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
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

// ActivityProductResponse 活动商品响应结构体，用于返回活动关联的商品信息
type ActivityProductResponse struct {
	ProductID   int             `json:"product_id"`   // 商品ID
	ProductName string          `json:"product_name"` // 商品名称
	SkuID       int             `json:"sku_id"`       // SKU ID
	SkuCode     string          `json:"sku_code"`     // SKU编码
	Price       decimal.Decimal `json:"price"`        // SKU价格
	Stock       int             `json:"stock"`        // SKU库存
}

// RedeemCodeRulesResponse 兑换码规则响应结构体，用于返回兑换码活动的配置规则
type RedeemCodeRulesResponse struct {
	Type         string `json:"type"`          // 兑换码类型：number(数字)、letter(字母)、mixed(混合)
	Length       int    `json:"length"`        // 兑换码长度
	ExcludeChars string `json:"exclude_chars"` // 排除的字符
}

// ActivityDetailResponse 活动详情响应结构体，用于返回单个活动的完整信息
type ActivityDetailResponse struct {
	ID              int                       `json:"id"`                          // 活动ID
	Name            string                    `json:"name"`                        // 活动名称
	Type            string                    `json:"type"`                        // 活动类型：seckill(秒杀)、redeem_code(兑换码)
	StartTime       string                    `json:"start_time"`                  // 开始时间，格式：2006-01-02 15:04:05
	EndTime         string                    `json:"end_time"`                    // 结束时间，格式：2006-01-02 15:04:05
	Status          string                    `json:"status"`                      // 活动状态：active(激活)、inactive(禁用)
	Products        []ActivityProductResponse `json:"products"`                    // 关联商品列表
	RedeemCodeRules *RedeemCodeRulesResponse  `json:"redeem_code_rules,omitempty"` // 兑换码规则（仅兑换码活动有值）
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

// CreateActivityProductRequest 创建活动商品请求结构体，定义活动关联商品的请求参数
type CreateActivityProductRequest struct {
	ProductID   int    `json:"product_id" binding:"required"` // 商品ID，必填
	SkuID       int    `json:"sku_id" binding:"required"`     // SKU ID，必填
	ProductType string `json:"product_type"`                  // 商品类型：seckill(秒杀)、redeem(兑换码)，默认为seckill
}

// CreateActivityRedeemSettingRequest 创建活动兑换码设置请求结构体，定义兑换码活动的配置参数
type CreateActivityRedeemSettingRequest struct {
	CodeType      string `json:"code_type"`      // 兑换码类型：number(数字)、letter(字母)、mixed(混合)
	CodeLength    int    `json:"code_length"`    // 兑换码长度，范围6-20
	ExcludeChars  string `json:"exclude_chars"`  // 排除的字符，如：01IOl（避免混淆的字符）
	TotalQuantity int    `json:"total_quantity"` // 兑换码总生成数量
	LimitPerUser  int    `json:"limit_per_user"` // 每个用户限制兑换数量
	NeedVerify    int    `json:"need_verify"`    // 是否需要验证：0(不需要)、1(需要)
}

// CreateActivityRequest 创建活动请求结构体，包含创建活动所需的全部参数
type CreateActivityRequest struct {
	Name          string                              `json:"name" binding:"required"`       // 活动名称，必填
	Type          string                              `json:"type" binding:"required"`       // 活动类型：seckill(秒杀)、redeem_code(兑换码)，必填
	StartTime     string                              `json:"start_time" binding:"required"` // 开始时间，格式：2006-01-02 15:04:05，必填
	EndTime       string                              `json:"end_time" binding:"required"`   // 结束时间，格式：2006-01-02 15:04:05，必填
	Status        string                              `json:"status"`                        // 活动状态：active(激活)、inactive(禁用)，默认active
	Products      []CreateActivityProductRequest      `json:"products"`                      // 关联商品列表
	RedeemSetting *CreateActivityRedeemSettingRequest `json:"redeem_setting"`                // 兑换码设置（仅兑换码活动需要）
}

// CreateActivity 创建活动
// @Summary 创建活动
// @Description 创建新活动，包含活动基本信息、关联商品和兑换码设置（仅兑换码活动）
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param activity body CreateActivityRequest true "活动信息"
// @Success 200 {object} map[string]interface{} "返回成功消息"
// @Router /api/activities [post]
func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	// 从上下文获取商户ID，用于验证用户权限和关联数据
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

	// 解析请求参数，包括活动名称、类型、时间、状态等信息
	var req CreateActivityRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 解析开始时间字符串为time.Time类型，格式：2006-01-02 15:04:05
	startTime, err := time.Parse(time.DateTime, req.StartTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 解析结束时间字符串为time.Time类型，格式：2006-01-02 15:04:05
	endTime, err := time.Parse(time.DateTime, req.EndTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 构建活动模型，包含商户ID、名称、类型、时间范围等基本信息
	activity := &models.Activity{
		MerchantID: merchantID,
		Name:       req.Name,
		Type:       req.Type,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     req.Status,
		CreatedBy:  createdBy.(int),
	}

	// 如果未指定状态，默认为激活状态
	if activity.Status == "" {
		activity.Status = constants.StatusActive
	}

	// 构建活动商品关联列表，秒杀活动类型默认为seckill，兑换码活动类型为redeem
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
			ActivityID:  activity.ID,
			ProductID:   p.ProductID,
			SkuID:       p.SkuID,
			MerchantID:  merchantID,
			ProductType: productType,
			Status:      constants.StatusActive,
		})
	}

	// 如果是兑换码活动且提供了兑换码设置，则构建兑换码配置
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

	// 调用服务层创建活动，包含活动信息、商品关联和兑换码设置
	err = c.activityService.CreateActivity(activity, activityProducts, redeemSetting)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动创建成功"})
}

// UpdateActivityRequest 更新活动请求结构体，包含更新活动所需的全部参数
type UpdateActivityRequest struct {
	Name          string                              `json:"name" binding:"required"`       // 活动名称，必填
	StartTime     string                              `json:"start_time" binding:"required"` // 开始时间，格式：2006-01-02 15:04:05，必填
	EndTime       string                              `json:"end_time" binding:"required"`   // 结束时间，格式：2006-01-02 15:04:05，必填
	Status        string                              `json:"status" binding:"required"`     // 活动状态：active(激活)、inactive(禁用)，必填
	Products      []CreateActivityProductRequest      `json:"products"`                      // 关联商品列表
	RedeemSetting *CreateActivityRedeemSettingRequest `json:"redeem_setting"`                // 兑换码设置（仅兑换码活动需要）
}

// UpdateActivity 更新活动
// @Summary 更新活动
// @Description 更新指定ID的活动信息，包括活动基本信息、关联商品和兑换码设置（仅兑换码活动）
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Param activity body UpdateActivityRequest true "活动信息"
// @Success 200 {object} map[string]interface{} "返回成功消息"
// @Router /api/activities/{id} [put]
func (c *ActivityController) UpdateActivity(ctx *gin.Context) {
	// 从上下文获取商户ID，用于验证用户权限和数据归属
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	// 从URL路径参数获取活动ID
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 解析请求参数，包含更新的活动信息
	var req UpdateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 根据活动ID和商户ID查询现有活动，确保活动存在且属于当前商户
	activity, err := c.activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	// 解析并更新开始时间，格式：2006-01-02 15:04:05
	startTime, err := time.Parse(time.DateTime, req.StartTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}
	activity.StartTime = startTime

	// 解析并更新结束时间，格式：2006-01-02 15:04:05
	endTime, err := time.Parse(time.DateTime, req.EndTime)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}
	activity.EndTime = endTime

	// 更新活动名称和状态
	activity.Name = req.Name
	activity.Status = req.Status

	// 重新构建活动商品关联列表，替换原有的商品关联关系
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
			ActivityID:  activity.ID,
			ProductID:   p.ProductID,
			SkuID:       p.SkuID,
			MerchantID:  merchantID,
			ProductType: productType,
			Status:      constants.StatusActive,
		})
	}

	// 如果是兑换码活动且提供了兑换码设置，则更新兑换码配置
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
// @Description 删除指定ID的活动，同时删除关联的商品和兑换码设置
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Success 200 {object} map[string]interface{} "返回成功消息"
// @Router /api/activities/{id} [delete]
func (c *ActivityController) DeleteActivity(ctx *gin.Context) {
	// 从上下文获取商户ID，用于验证用户权限和数据归属
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	// 从URL路径参数获取要删除的活动ID
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 调用服务层删除活动，包含活动本身及其关联的商品和兑换码设置
	err = c.activityService.DeleteActivity(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动删除成功"})
}

// GetActivity 获取活动详情
// @Summary 获取活动详情
// @Description 根据活动ID获取单个活动的完整信息，包括基本信息、关联商品和兑换码规则
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Success 200 {object} ActivityDetailResponse "返回活动详情"
// @Router /api/activities/{id} [get]
func (c *ActivityController) GetActivity(ctx *gin.Context) {
	// 从上下文获取商户ID，用于验证用户权限和数据归属
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	// 从URL路径参数获取活动ID
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 根据活动ID和商户ID查询活动详情，包含关联的商品和兑换码设置
	activity, err := c.activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	// 构建活动详情响应，将时间格式化为 2006-01-02 15:04:05 字符串格式
	response := ActivityDetailResponse{
		ID:        activity.ID,
		Name:      activity.Name,
		Type:      activity.Type,
		StartTime: activity.StartTime.Format(time.DateTime),
		EndTime:   activity.EndTime.Format(time.DateTime),
		Status:    activity.Status,
		Products:  make([]ActivityProductResponse, 0),
	}

	// 处理关联商品，填充商品名称和SKU编码等信息
	for _, product := range activity.Products {
		productResponse := ActivityProductResponse{
			ProductID:   product.ProductID,
			ProductName: "",
			SkuID:       product.SkuID,
			SkuCode:     "",
		}

		// 获取商品名称
		if product.Product.ID != 0 {
			productResponse.ProductName = product.Product.Name
		}

		// 获取SKU编码、价格和库存
		if product.Sku.ID != 0 {
			productResponse.SkuCode = product.Sku.SkuCode
			productResponse.Price = product.Sku.Price
			productResponse.Stock = product.Sku.Stock
		}

		response.Products = append(response.Products, productResponse)
	}

	// 如果是兑换码活动且有兑换码设置，则填充兑换码规则响应
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
// @Description 获取当前商户的活动列表，支持分页和按类型、状态筛选
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param size query int false "每页大小，默认10"
// @Param type query string false "活动类型筛选：seckill(秒杀)、redeem_code(兑换码)"
// @Param status query string false "活动状态筛选：active(激活)、inactive(禁用)"
// @Success 200 {object} map[string]interface{} "返回活动列表和总数"
// @Router /api/activities [get]
func (c *ActivityController) GetActivities(ctx *gin.Context) {
	// 从上下文获取商户ID，用于筛选当前商户的活动数据
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	// 解析分页参数，默认页码1，每页10条
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	// 解析筛选参数：活动类型和状态
	activityType := ctx.Query("type")
	status := ctx.Query("status")

	// 调用服务层获取活动列表，包含分页和筛选条件
	activities, total, err := c.activityService.GetActivities(merchantID, activityType, status, page, pageSize)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	// 构建活动列表响应，将时间格式化为 2006-01-02 15:04:05 字符串格式
	var activityList []map[string]interface{}
	for _, activity := range activities {
		activityList = append(activityList, map[string]interface{}{
			"id":            activity.ID,
			"name":          activity.Name,
			"type":          activity.Type,
			"start_time":    activity.StartTime.Format(time.DateTime),
			"end_time":      activity.EndTime.Format(time.DateTime),
			"status":        activity.Status,
			"product_count": len(activity.Products),
			"created_at":    activity.CreatedAt.Format(time.DateTime),
		})
	}

	// 返回活动列表和总数，支持前端分页组件
	c.ResponseSuccess(ctx, gin.H{
		"list":  activityList,
		"total": total,
	})
}

// UpdateActivityStatusRequest 更新活动状态请求结构体，用于快速更新活动的启用/禁用状态
type UpdateActivityStatusRequest struct {
	Status string `json:"status" binding:"required"` // 活动状态：active(激活)、inactive(禁用)，必填
}

// UpdateActivityStatus 更新活动状态
// @Summary 更新活动状态
// @Description 更新指定ID活动的状态，用于启用或禁用活动
// @Tags 活动管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Param status body UpdateActivityStatusRequest true "活动状态"
// @Success 200 {object} map[string]interface{} "返回成功消息"
// @Router /api/activities/{id}/status [put]
func (c *ActivityController) UpdateActivityStatus(ctx *gin.Context) {
	// 从上下文获取商户ID，用于验证用户权限和数据归属
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.merchantService)
	if err != nil {
		if err.Error() == errors.GetErrorMessage(errors.CodeUnauthorized) {
			c.ResponseError(ctx, errors.CodeUnauthorized, err)
		} else {
			c.ResponseError(ctx, errors.CodeForbidden, err)
		}
		return
	}

	// 从URL路径参数获取活动ID
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 解析请求参数，包含新的活动状态
	var req UpdateActivityStatusRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 调用服务层更新活动状态
	err = c.activityService.UpdateActivityStatus(activityID, req.Status, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "活动状态更新成功"})
}
