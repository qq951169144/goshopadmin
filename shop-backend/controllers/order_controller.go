package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shop-backend/services"
)

// OrderController 订单控制器
type OrderController struct {
	BaseController
	orderService *services.OrderService
}

// NewOrderController 创建订单控制器实例
func NewOrderController(orderService *services.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

// CreateOrderRequest 创建订单请求结构
type CreateOrderRequest struct {
	Items []services.OrderItemInfo `json:"items" binding:"required"`
}

// CreateOrder 创建订单
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 创建订单
	order, err := c.orderService.CreateOrder(services.CreateOrderRequest{
		UserID: userID.(uint),
		Items:  req.Items,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, order)
}

// GetOrderDetail 获取订单详情
func (c *OrderController) GetOrderDetail(ctx *gin.Context) {
	orderID := ctx.Param("id")

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 从服务层获取订单详情
	order, err := c.orderService.GetOrderDetail(orderID, userID.(uint))
	if err != nil {
		c.ResponseError(ctx, http.StatusNotFound, err.Error())
		return
	}

	c.ResponseSuccess(ctx, order)
}
