package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/errors"
	"shop-backend/services"
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
