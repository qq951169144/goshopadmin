package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MerchantController 商户控制器
type MerchantController struct {
	BaseController
	merchantService *services.MerchantService
}

// NewMerchantController 创建商户控制器实例
func NewMerchantController(db *gorm.DB) *MerchantController {
	return &MerchantController{
		merchantService: services.NewMerchantService(db),
	}
}

// GetMerchants 获取商户列表
func (c *MerchantController) GetMerchants(ctx *gin.Context) {
	merchants, err := c.merchantService.GetMerchants()
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, merchants)
}

// GetMerchant 获取单个商户信息
func (c *MerchantController) GetMerchant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	merchant, err := c.merchantService.GetMerchantByID(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, merchant)
}

// CreateMerchantRequest 创建商户请求结构
type CreateMerchantRequest struct {
	Name            string `json:"name" binding:"required"`
	ContactName     string `json:"contact_name" binding:"required"`
	ContactPhone    string `json:"contact_phone" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Address         string `json:"address" binding:"required"`
	BusinessLicense string `json:"business_license" binding:"required"`
	TaxNumber       string `json:"tax_number" binding:"required"`
}

// CreateMerchant 创建商户
func (c *MerchantController) CreateMerchant(ctx *gin.Context) {
	var req CreateMerchantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取当前用户ID
	userID, ok := c.GetUserID(ctx)
	if !ok {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	merchant, err := c.merchantService.CreateMerchant(
		req.Name,
		req.ContactName,
		req.ContactPhone,
		req.Email,
		req.Address,
		req.BusinessLicense,
		req.TaxNumber,
		userID,
	)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, merchant)
}

// UpdateMerchantRequest 更新商户请求结构
type UpdateMerchantRequest struct {
	Name            string `json:"name"`
	ContactName     string `json:"contact_name"`
	ContactPhone    string `json:"contact_phone"`
	Email           string `json:"email"`
	Address         string `json:"address"`
	BusinessLicense string `json:"business_license"`
	TaxNumber       string `json:"tax_number"`
	Status          string `json:"status"`
}

// UpdateMerchant 更新商户
func (c *MerchantController) UpdateMerchant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req UpdateMerchantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取当前用户ID
	userID, ok := c.GetUserID(ctx)
	if !ok {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	merchant, err := c.merchantService.UpdateMerchant(
		id,
		req.Name,
		req.ContactName,
		req.ContactPhone,
		req.Email,
		req.Address,
		req.BusinessLicense,
		req.TaxNumber,
		req.Status,
		userID,
	)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, merchant)
}

// AuditMerchantRequest 审核商户请求结构
type AuditMerchantRequest struct {
	AuditStatus string `json:"audit_status" binding:"required"`
	AuditNote   string `json:"audit_note"`
}

// AuditMerchant 审核商户
func (c *MerchantController) AuditMerchant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req AuditMerchantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取当前用户ID
	userID, ok := c.GetUserID(ctx)
	if !ok {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	err = c.merchantService.AuditMerchant(id, req.AuditStatus, req.AuditNote, userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// GetMerchantUsers 获取商户用户列表
func (c *MerchantController) GetMerchantUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	merchantID, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	merchantUsers, err := c.merchantService.GetMerchantUsers(merchantID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, merchantUsers)
}

// AddMerchantUserRequest 添加商户用户请求结构
type AddMerchantUserRequest struct {
	UserID int    `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// AddMerchantUser 添加商户用户
func (c *MerchantController) AddMerchantUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	merchantID, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req AddMerchantUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	err = c.merchantService.AddMerchantUser(merchantID, req.UserID, req.Role)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// RemoveMerchantUser 移除商户用户
func (c *MerchantController) RemoveMerchantUser(ctx *gin.Context) {
	merchantIDStr := ctx.Param("id")
	userIDStr := ctx.Param("user_id")

	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	err = c.merchantService.RemoveMerchantUser(merchantID, userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// DeleteMerchant 禁用商户
func (c *MerchantController) DeleteMerchant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	err = c.merchantService.DeleteMerchant(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}
