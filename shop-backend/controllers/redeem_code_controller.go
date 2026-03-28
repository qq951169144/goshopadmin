package controllers

import (
	"shop-backend/errors"
	"shop-backend/services"
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

// VerifyRedeemCodeRequest 验证兑换码请求参数
type VerifyRedeemCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// VerifyRedeemCode 验证兑换码
func (c *RedeemCodeController) VerifyRedeemCode(ctx *gin.Context) {
	var req VerifyRedeemCodeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	codeInfo, err := c.redeemCodeService.VerifyRedeemCode(req.Code)
	if err != nil {
		c.ResponseError(ctx, errors.CodeRedeemCodeInvalid, err)
		return
	}

	c.ResponseSuccess(ctx, codeInfo)
}

// RedeemCodeRequest 兑换码兑换请求参数
type RedeemCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// RedeemCode 兑换码兑换
func (c *RedeemCodeController) RedeemCode(ctx *gin.Context) {
	customerID, _ := ctx.Get("customer_id")

	var req RedeemCodeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	codeInfo, err := c.redeemCodeService.RedeemCode(req.Code, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeRedeemCodeInvalid, err)
		return
	}

	c.ResponseSuccess(ctx, codeInfo)
}

// GetRedeemCodeLogs 获取用户兑换码使用记录
func (c *RedeemCodeController) GetRedeemCodeLogs(ctx *gin.Context) {
	customerID, _ := ctx.Get("customer_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	logs, total, err := c.redeemCodeService.GetRedeemCodeLogs(customerID.(int), page, pageSize)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"items":     logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRedeemCodeByActivity 获取活动的兑换码信息
func (c *RedeemCodeController) GetRedeemCodeByActivity(ctx *gin.Context) {
	activityID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	setting, err := c.redeemCodeService.GetRedeemCodeByActivity(activityID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, setting)
}
