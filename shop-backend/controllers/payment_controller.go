package controllers

import (
	"fmt"
	"time"

	"shop-backend/cache"
	"shop-backend/constants"
	"shop-backend/errors"
	"shop-backend/pkg/mq"
	"shop-backend/services"
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaymentController 支付控制器
type PaymentController struct {
	BaseController
	orderService *services.OrderService
}

// NewPaymentController 创建支付控制器实例
func NewPaymentController(db *gorm.DB, cacheUtil *cache.CacheUtil) *PaymentController {
	return &PaymentController{
		orderService: services.NewOrderService(db, cacheUtil),
	}
}

// FakePay 伪支付接口
func (c *PaymentController) FakePay(ctx *gin.Context) {
	orderNo := ctx.Query("orderNo")
	if orderNo == "" {
		c.ResponseError(ctx, errors.CodeParamMissing, nil)
		return
	}

	// 查找订单
	order, err := c.orderService.GetOrderByOrderNo(orderNo)
	if err != nil {
		c.ResponseError(ctx, errors.CodeOrderNotFound, err)
		return
	}

	// 模拟支付回调
	go func() {
		// 生成交易ID
		transactionID := fmt.Sprintf("TRX%s", time.Now().Format("20060102150405"))

		// 更新订单状态和支付状态
		err := c.orderService.UpdateOrderStatus(orderNo, order.CustomerID, constants.OrderStatusPaid, constants.PaymentStatusSuccess, transactionID)
		if err != nil {
			utils.Error("更新订单状态失败: %v", err)
			return
		}

		// 发送状态变更消息
		conn, err := mq.NewConnection()
		if err != nil {
			utils.Error("创建MQ连接失败: %v", err)
			return
		}
		defer conn.Close()

		producer := mq.NewProducer(conn)
		msg := map[string]interface{}{
			"order_id":   order.ID,
			"status":     constants.OrderStatusPaid,
			"updated_at": time.Now(),
		}
		err = producer.Publish(constants.MQExchangeOrderStatus, constants.MQRoutingKeyOrderStatus, msg)
		if err != nil {
			utils.Error("发送订单状态变更消息失败: %v", err)
		}
	}()

	// 返回 JSON 响应
	c.ResponseSuccess(ctx, gin.H{
		"order_no": orderNo,
		"amount":   order.TotalAmount,
		"message":  "支付成功",
	})
}

// PaymentCallbackRequest 支付回调请求结构
type PaymentCallbackRequest struct {
	OrderNo       string  `json:"order_no" binding:"required"`
	TransactionID string  `json:"transaction_id" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

// PaymentCallback 支付回调
func (c *PaymentController) PaymentCallback(ctx *gin.Context) {
	var req PaymentCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 查找订单
	order, err := c.orderService.GetOrderByOrderNo(req.OrderNo)
	if err != nil {
		c.ResponseError(ctx, errors.CodeOrderNotFound, err)
		return
	}

	// 验证金额
	if order.TotalAmount != req.Amount {
		c.ResponseError(ctx, errors.CodeParamInvalid, nil)
		return
	}

	// 更新订单状态
	err = c.orderService.UpdateOrderStatus(req.OrderNo, order.CustomerID, req.Status, constants.PaymentStatusSuccess, req.TransactionID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	// 发送状态变更消息
	go func() {
		conn, err := mq.NewConnection()
		if err != nil {
			utils.Error("创建MQ连接失败: %v", err)
			return
		}
		defer conn.Close()

		producer := mq.NewProducer(conn)
		msg := map[string]interface{}{
			"order_id":   order.ID,
			"status":     req.Status,
			"updated_at": time.Now(),
		}
		err = producer.Publish(constants.MQExchangeOrderStatus, constants.MQRoutingKeyOrderStatus, msg)
		if err != nil {
			utils.Error("发送订单状态变更消息失败: %v", err)
		}
	}()

	c.ResponseSuccess(ctx, gin.H{
		"message":        "Payment callback received",
		"order_no":       req.OrderNo,
		"transaction_id": req.TransactionID,
		"status":         req.Status,
	})
}
