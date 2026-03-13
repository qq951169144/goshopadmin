package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/services"
)

// CustomerController 客户控制器
type CustomerController struct {
	BaseController
	customerService *services.CustomerService
}

// NewCustomerController 创建客户控制器实例
func NewCustomerController(customerService *services.CustomerService) *CustomerController {
	return &CustomerController{
		customerService: customerService,
	}
}

// UpdateProfileRequest 更新个人信息请求结构
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetProfile 获取个人信息
func (c *CustomerController) GetProfile(ctx *gin.Context) {
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 从服务层获取客户信息
	customer, err := c.customerService.GetProfile(customerID.(uint))
	if err != nil {
		c.ResponseError(ctx, http.StatusNotFound, err.Error())
		return
	}

	c.ResponseSuccess(ctx, customer)
}

// UpdateProfile 更新个人信息
func (c *CustomerController) UpdateProfile(ctx *gin.Context) {
	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 更新客户信息
	customer, err := c.customerService.UpdateProfile(customerID.(uint), services.UpdateProfileRequest{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
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
	// 从上下文中获取用户ID
	customerID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// 从服务层获取订单列表
	orders, total, err := c.customerService.GetOrders(customerID.(uint), page, limit)
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"orders": orders,
		"total":  total,
	})
}
