package controllers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/errors"
	"shop-backend/services"
	"shop-backend/utils"
	"gorm.io/gorm"
)

// CustomerController 客户控制器
type CustomerController struct {
	BaseController
	customerService *services.CustomerService
}

// NewCustomerController 创建客户控制器实例
func NewCustomerController(db *gorm.DB) *CustomerController {
	return &CustomerController{
		customerService: services.NewCustomerService(db),
	}
}

// UpdateProfileRequest 更新个人信息请求结构
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetProfile 获取个人信息
func (c *CustomerController) GetProfile(ctx *gin.Context) {
	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 从服务层获取客户信息
	customer, err := c.customerService.GetProfile(customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeUserNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"username": customer.Username,
		"email":    customer.Email,
		"phone":    customer.Phone,
		"nickname": customer.Nickname,
		"avatar":   customer.Avatar,
	})
}

// UpdateProfile 更新个人信息
func (c *CustomerController) UpdateProfile(ctx *gin.Context) {
	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 更新客户信息
	customer, err := c.customerService.UpdateProfile(customerID.(int), services.UpdateProfileRequest{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message":  "Profile updated",
		"username": customer.Username,
		"email":    customer.Email,
		"phone":    customer.Phone,
		"nickname": customer.Nickname,
		"avatar":   customer.Avatar,
	})
}

// GetOrders 获取订单列表
func (c *CustomerController) GetOrders(ctx *gin.Context) {
	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	status := ctx.DefaultQuery("status", "") // 状态筛选参数

	// 从服务层获取订单列表
	orders, total, err := c.customerService.GetOrders(customerID.(int), page, limit, status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"orders": orders,
		"total":  total,
	})
}

// UploadAvatar 上传客户头像
func (c *CustomerController) UploadAvatar(ctx *gin.Context) {
	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("avatar")
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, fmt.Errorf("获取上传文件失败: %w", err))
		return
	}

	// 上传头像图片到存储
	avatarURL, err := utils.UploadAvatar(file, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 获取旧头像URL
	oldAvatarURL, _ := c.customerService.GetAvatarURL(customerID.(int))

	// 更新数据库中的头像URL
	if err := c.customerService.UpdateAvatar(customerID.(int), avatarURL); err != nil {
		// 如果数据库操作失败，删除已上传的图片
		utils.DeleteImage(avatarURL)
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	// 删除旧头像文件（如果存在）
	if oldAvatarURL != "" {
		utils.DeleteImage(oldAvatarURL)
	}

	c.ResponseSuccess(ctx, gin.H{
		"avatar": avatarURL,
	})
}
