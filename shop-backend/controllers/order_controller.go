package controllers

import (
	"shop-backend/cache"
	"shop-backend/errors"
	"shop-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrderController 订单控制器
type OrderController struct {
	BaseController
	orderService *services.OrderService
}

// NewOrderController 创建订单控制器实例
func NewOrderController(db *gorm.DB, cacheUtil *cache.CacheUtil) *OrderController {
	return &OrderController{
		orderService: services.NewOrderService(db, cacheUtil),
	}
}

// CreateOrderItemRequest 创建订单项请求结构
type CreateOrderItemRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	SkuID     int `json:"sku_id"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// CreateOrderRequest 创建订单请求结构
type CreateOrderRequest struct {
	AddressID int                      `json:"address_id" binding:"required"`
	Items     []CreateOrderItemRequest `json:"items" binding:"required"`
	Remark    string                   `json:"remark"`
}

// CreateOrder 创建订单
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req CreateOrderRequest
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

	// 转换请求参数
	var orderItems []services.OrderItemRequest
	for _, item := range req.Items {
		orderItems = append(orderItems, services.OrderItemRequest{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		})
	}

	// 创建订单
	order, err := c.orderService.CreateOrder(services.CreateOrderRequest{
		CustomerID: customerID.(int),
		AddressID:  req.AddressID,
		Items:      orderItems,
		Remark:     req.Remark,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, order)
}

// GetOrderDetail 获取订单详情
func (c *OrderController) GetOrderDetail(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 从服务层获取订单详情
	order, err := c.orderService.GetOrderDetail(orderNo, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeOrderNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, order)
}

// CancelOrder 取消订单
func (c *OrderController) CancelOrder(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 调用服务层取消订单
	err := c.orderService.CancelOrder(orderNo, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "订单已取消",
	})
}

// ConfirmReceipt 确认收货
func (c *OrderController) ConfirmReceipt(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 调用服务层确认收货
	err := c.orderService.ConfirmReceipt(orderNo, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "确认收货成功",
	})
}
