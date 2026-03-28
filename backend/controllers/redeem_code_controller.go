package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RedeemCodeController 兑换码控制器
type RedeemCodeController struct {
	BaseController
	redeemCodeService *services.RedeemCodeService
	DB                *gorm.DB
}

// NewRedeemCodeController 创建兑换码控制器实例
func NewRedeemCodeController(db *gorm.DB) *RedeemCodeController {
	return &RedeemCodeController{
		redeemCodeService: services.NewRedeemCodeService(db),
		DB:                db,
	}
}

// GenerateRedeemCodesRequest 生成兑换码请求
type GenerateRedeemCodesRequest struct {
	ActivityID   int    `json:"activity_id" binding:"required"`              // 活动ID
	Quantity     int    `json:"quantity" binding:"required,min=1,max=10000"` // 生成数量，默认值：无
	CodeType     string `json:"code_type"`                                   // 码类型，默认值：alphanumeric
	CodeLength   int    `json:"code_length"`                                 // 码长度，默认值：8
	ExcludeChars string `json:"exclude_chars"`                               // 排除字符，默认值：IO10
}

// GenerateRedeemCodes 生成兑换码
// @Summary 生成兑换码
// @Description 为活动生成兑换码
// @Tags 兑换码管理
// @Accept json
// @Produce json
// @Param request body GenerateRedeemCodesRequest true "生成兑换码请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/redeem-codes/generate [post]
func (c *RedeemCodeController) GenerateRedeemCodes(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	createdBy, _ := ctx.Get("user_id")

	var req GenerateRedeemCodesRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 设置默认值
	if req.CodeType == "" {
		req.CodeType = "alphanumeric"
	}
	if req.CodeLength == 0 {
		req.CodeLength = 8
	}
	if req.ExcludeChars == "" {
		req.ExcludeChars = "IO10"
	}

	codes, err := c.redeemCodeService.GenerateRedeemCodes(
		req.ActivityID,
		merchantID,
		createdBy.(int),
		req.Quantity,
		req.CodeType,
		req.CodeLength,
		req.ExcludeChars,
	)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"generated_count": len(codes),
		"codes":           codes,
	})
}

// GetRedeemCodes 获取兑换码列表
// @Summary 获取兑换码列表
// @Description 获取活动的兑换码列表
// @Tags 兑换码管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Param status query string false "兑换码状态"
// @Param page query int false "页码"
// @Param size query int false "每页大小"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id}/redeem-codes [get]
func (c *RedeemCodeController) GetRedeemCodes(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	activityID, _ := strconv.Atoi(ctx.Param("id"))
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("size", "20"))

	codes, total, stats, err := c.redeemCodeService.GetRedeemCodes(activityID, merchantID, status, page, pageSize)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	// 构建响应数据
	var codeList []map[string]interface{}
	for _, code := range codes {
		codeList = append(codeList, map[string]interface{}{
			"id":               code.ID,
			"code":             code.Code,
			"status":           code.Status,
			"valid_start_time": code.ValidStartTime,
			"valid_end_time":   code.ValidEndTime,
			"created_at":       code.CreatedAt,
		})
	}

	c.ResponseSuccess(ctx, gin.H{
		"list":  codeList,
		"total": total,
		"stats": stats,
	})
}

// ExportRedeemCodes 导出兑换码
// @Summary 导出兑换码
// @Description 导出活动的兑换码
// @Tags 兑换码管理
// @Accept json
// @Produce json
// @Param id path int true "活动ID"
// @Param status query string false "兑换码状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id}/redeem-codes/export [get]
func (c *RedeemCodeController) ExportRedeemCodes(ctx *gin.Context) {
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

	// 验证活动是否属于该商户
	activityService := services.NewActivityService(c.DB)
	_, err = activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	// TODO: 后续再写导出逻辑
	c.ResponseSuccess(ctx, gin.H{"message": "导出功能开发中"})
}

// ImportRedeemCodes 导入兑换码
// @Summary 导入兑换码
// @Description 为活动导入兑换码
// @Tags 兑换码管理
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "活动ID"
// @Param file formData file true "CSV文件"
// @Success 200 {object} map[string]interface{}
// @Router /api/activities/{id}/redeem-codes/import [post]
func (c *RedeemCodeController) ImportRedeemCodes(ctx *gin.Context) {
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

	// 验证活动是否属于该商户
	activityService := services.NewActivityService(c.DB)
	_, err = activityService.GetActivityByID(activityID, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeActivityNotFound, err)
		return
	}

	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}
	defer file.Close()

	// TODO: 后续再写导入逻辑
	c.ResponseSuccess(ctx, gin.H{"imported_count": 0, "failed_count": 0, "failed_codes": []string{}})
}

// VerifyRedeemCodeRequest 核销兑换码请求
type VerifyRedeemCodeRequest struct {
	Code   string `json:"code" binding:"required"` // 兑换码
	Remark string `json:"remark"`                  // 备注，默认值：空
}

// VerifyRedeemCode 核销兑换码
// @Summary 核销兑换码
// @Description 核销兑换码
// @Tags 兑换码管理
// @Accept json
// @Produce json
// @Param request body VerifyRedeemCodeRequest true "核销兑换码请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/redeem-codes/verify [post]
func (c *RedeemCodeController) VerifyRedeemCode(ctx *gin.Context) {
	verifyBy, _ := ctx.Get("user_id")

	var req VerifyRedeemCodeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	redeemCode, log, err := c.redeemCodeService.VerifyRedeemCode(req.Code, verifyBy.(int), req.Remark)
	if err != nil {
		c.ResponseError(ctx, errors.CodeRedeemCodeInvalid, err)
		return
	}

	// 构建响应数据
	response := map[string]interface{}{
		"code":        redeemCode.Code,
		"status":      redeemCode.Status,
		"verified_at": log.VerifyAt,
	}

	c.ResponseSuccess(ctx, response)
}

// GetRedeemCodeLogs 获取核销记录
// @Summary 获取核销记录
// @Description 获取兑换码核销记录
// @Tags 兑换码管理
// @Accept json
// @Produce json
// @Param activity_id query int false "活动ID"
// @Param page query int false "页码"
// @Param size query int false "每页大小"
// @Success 200 {object} map[string]interface{}
// @Router /api/redeem-code-logs [get]
func (c *RedeemCodeController) GetRedeemCodeLogs(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	activityID, _ := strconv.Atoi(ctx.Query("activity_id"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("size", "20"))

	logs, total, err := c.redeemCodeService.GetRedeemCodeLogs(activityID, merchantID, page, pageSize)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	// 构建响应数据
	var logList []map[string]interface{}
	for _, log := range logs {
		logList = append(logList, map[string]interface{}{
			"id":          log.ID,
			"code":        log.Code,
			"customer_id": log.CustomerID,
			"status":      log.Status,
			"verify_by":   log.VerifyBy,
			"verify_at":   log.VerifyAt,
			"remark":      log.Remark,
			"created_at":  log.CreatedAt,
		})
	}

	c.ResponseSuccess(ctx, gin.H{
		"list":  logList,
		"total": total,
	})
}

// UpdateRedeemCodeStatusRequest 更新兑换码状态请求
type UpdateRedeemCodeStatusRequest struct {
	Status string `json:"status" binding:"required"` // 状态，默认值：无
}

// UpdateRedeemCodeStatus 更新兑换码状态
// @Summary 更新兑换码状态
// @Description 更新兑换码状态
// @Tags 兑换码管理
// @Accept json
// @Produce json
// @Param id path int true "兑换码ID"
// @Param status body UpdateRedeemCodeStatusRequest true "状态更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/redeem-codes/{id}/status [put]
func (c *RedeemCodeController) UpdateRedeemCodeStatus(ctx *gin.Context) {
	// 获取商户ID
	merchantID, err := c.GetMerchantIDFromContext(ctx, c.DB)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUnauthorized, err)
		return
	}

	redeemCodeID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req UpdateRedeemCodeStatusRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	err = c.redeemCodeService.UpdateRedeemCodeStatus(redeemCodeID, req.Status, merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "兑换码状态更新成功"})
}
