package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"shop-backend/services"
)

// PaymentController 支付控制器
type PaymentController struct {
	BaseController
	orderService *services.OrderService
}

// NewPaymentController 创建支付控制器实例
func NewPaymentController(orderService *services.OrderService) *PaymentController {
	return &PaymentController{
		orderService: orderService,
	}
}

// FakePay 伪支付页面
func (c *PaymentController) FakePay(ctx *gin.Context) {
	orderID := ctx.Query("order_id")

	// 查找订单
	order, err := c.orderService.GetOrderByID(orderID)
	if err != nil {
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(http.StatusNotFound, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>订单不存在</title>
				<style>
					body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
					h1 { color: #f44336; }
					p { font-size: 18px; margin: 20px 0; }
					button { padding: 10px 20px; font-size: 16px; background-color: #333; color: white; border: none; border-radius: 4px; cursor: pointer; }
					button:hover { background-color: #555; }
				</style>
			</head>
			<body>
				<h1>订单不存在</h1>
				<p>请检查订单号是否正确</p>
				<button onclick="window.history.back()">返回</button>
			</body>
			</html>
		`)
		return
	}

	// 生成支付成功页面
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>支付成功</title>
			<style>
				body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
				h1 { color: #4CAF50; }
				p { font-size: 18px; margin: 20px 0; }
				button { padding: 10px 20px; font-size: 16px; background-color: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer; }
				button:hover { background-color: #45a049; }
			</style>
		</head>
		<body>
			<h1>支付成功</h1>
			<p>订单号: %s</p>
			<p>支付金额: ¥%.2f</p>
			<p>支付时间: %s</p>
			<button onclick="window.location.href='/order/%s'">查看订单</button>
		</body>
		</html>
	`, orderID, order.TotalAmount, time.Now().Format("2006-01-02 15:04:05"), orderID)

	// 模拟支付回调
	go func() {
		// 生成交易ID
		transactionID := fmt.Sprintf("TRX%s", time.Now().Format("20060102150405"))

		// 更新订单状态
		c.orderService.UpdateOrderStatus(orderID, "paid", transactionID)
	}()
}

// PaymentCallbackRequest 支付回调请求结构
type PaymentCallbackRequest struct {
	OrderID       string  `json:"order_id" binding:"required"`
	TransactionID string  `json:"transaction_id" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

// PaymentCallback 支付回调
func (c *PaymentController) PaymentCallback(ctx *gin.Context) {
	var req PaymentCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	// 查找订单
	order, err := c.orderService.GetOrderByID(req.OrderID)
	if err != nil {
		c.ResponseError(ctx, http.StatusNotFound, "Order not found")
		return
	}

	// 验证金额
	if order.TotalAmount != req.Amount {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid amount")
		return
	}

	// 更新订单状态
	err = c.orderService.UpdateOrderStatus(req.OrderID, req.Status, req.TransactionID)
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message":        "Payment callback received",
		"order_id":       req.OrderID,
		"transaction_id": req.TransactionID,
		"status":         req.Status,
	})
}
