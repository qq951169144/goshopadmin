package controllers

import (
	"shop-backend/errors"
	"shop-backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityOrderController 活动订单控制器
type ActivityOrderController struct {
	BaseController
	activityOrderService *services.ActivityOrderService
	DB                   *gorm.DB
}

// NewActivityOrderController 创建活动订单控制器实例
func NewActivityOrderController(db *gorm.DB) *ActivityOrderController {
	return &ActivityOrderController{
		activityOrderService: services.NewActivityOrderService(db),
		DB:                   db,
	}
}

// CreateActivityOrderRequest 创建活动订单请求
type CreateActivityOrderRequest struct {
	ActivityID int                       `json:"activity_id" binding:"required"`
	Items      []CreateActivityOrderItem `json:"items" binding:"required,dive"`
}

// CreateActivityOrderItem 创建活动订单商品项
type CreateActivityOrderItem struct {
	ProductID int `json:"product_id" binding:"required"`
	SkuID     int `json:"sku_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// CreateActivityOrder 创建活动订单
func (c *ActivityOrderController) CreateActivityOrder(ctx *gin.Context) {
	customerID, _ := ctx.Get("customer_id")

	var req CreateActivityOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	items := make([]services.ActivityOrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i].ProductID = item.ProductID
		items[i].SkuID = item.SkuID
		items[i].Quantity = item.Quantity
	}

	order, err := c.activityOrderService.CreateActivityOrder(customerID.(int), req.ActivityID, items)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, order)
}

// GetActivityOrders 获取用户活动订单列表
func (c *ActivityOrderController) GetActivityOrders(ctx *gin.Context) {
	customerID, _ := ctx.Get("customer_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	orders, total, err := c.activityOrderService.GetActivityOrders(customerID.(int), page, pageSize)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"items":     orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetActivityOrder 获取活动订单详情
func (c *ActivityOrderController) GetActivityOrder(ctx *gin.Context) {
	customerID, _ := ctx.Get("customer_id")
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	order, err := c.activityOrderService.GetActivityOrderByID(orderID, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeOrderNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, order)
}

// CancelActivityOrder 取消活动订单
func (c *ActivityOrderController) CancelActivityOrder(ctx *gin.Context) {
	customerID, _ := ctx.Get("customer_id")
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamError, err)
		return
	}

	err = c.activityOrderService.CancelActivityOrder(orderID, customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"message": "订单取消成功"})
}
