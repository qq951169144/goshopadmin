package controllers

import (
	"goshopadmin/services"
	"goshopadmin/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MerchantController 商户控制器
type MerchantController struct {
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商户列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取商户列表成功",
		"data":    merchants,
	})
}

// GetMerchant 获取单个商户信息
func (c *MerchantController) GetMerchant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	merchant, err := c.merchantService.GetMerchantByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商户信息失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取商户信息成功",
		"data":    merchant,
	})
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
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
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
		userID.(int),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建商户失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建商户成功",
		"data":    merchant,
	})
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
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req UpdateMerchantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
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
		userID.(int),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新商户失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新商户成功",
		"data":    merchant,
	})
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
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req AuditMerchantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	err = c.merchantService.AuditMerchant(id, req.AuditStatus, req.AuditNote, userID.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "审核商户失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "审核商户成功",
	})
}

// GetMerchantUsers 获取商户用户列表
func (c *MerchantController) GetMerchantUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	merchantID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	merchantUsers, err := c.merchantService.GetMerchantUsers(merchantID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取商户用户列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取商户用户列表成功",
		"data":    merchantUsers,
	})
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
		utils.Info("解析商户ID失败: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req AddMerchantUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Info("解析请求体失败: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	err = c.merchantService.AddMerchantUser(merchantID, req.UserID, req.Role)
	if err != nil {
		utils.Info("添加商户用户失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "添加商户用户失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加商户用户成功",
	})
}

// RemoveMerchantUser 移除商户用户
func (c *MerchantController) RemoveMerchantUser(ctx *gin.Context) {
	merchantIDStr := ctx.Param("id")
	userIDStr := ctx.Param("user_id")

	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	err = c.merchantService.RemoveMerchantUser(merchantID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "移除商户用户失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "移除商户用户成功",
	})
}

// DeleteMerchant 禁用商户
func (c *MerchantController) DeleteMerchant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	err = c.merchantService.DeleteMerchant(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "禁用商户失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "禁用商户成功",
	})
}
