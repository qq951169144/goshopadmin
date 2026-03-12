package controllers

import (
	"fmt"
	"net/http"
	"time"

	"shop-backend/config"
	"shop-backend/models"

	"github.com/gin-gonic/gin"
)

// 支付回调请求结构
type PaymentCallbackRequest struct {
	OrderID       string  `json:"order_id" binding:"required"`
	TransactionID string  `json:"transaction_id" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

// 伪支付页面
func FakePay(c *gin.Context) {
	orderID := c.Query("order_id")

	// 查找订单
	var order models.Order
	result := config.DB.Where("order_id = ?", orderID).First(&order)
	if result.RowsAffected == 0 {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusNotFound, `
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
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `
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

		// 调用支付回调接口
		// 实际项目中应该使用HTTP客户端调用
		fmt.Printf("模拟支付回调: order_id=%s, transaction_id=%s\n", orderID, transactionID)

		// 直接更新订单状态
		order.Status = "paid"
		order.TransactionID = transactionID
		config.DB.Save(&order)
	}()
}

// 支付回调
func PaymentCallback(c *gin.Context) {
	var req PaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 查找订单
	var order models.Order
	result := config.DB.Where("order_id = ?", req.OrderID).First(&order)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "Order not found")
		return
	}

	// 验证金额
	if order.TotalAmount != req.Amount {
		ResponseError(c, http.StatusBadRequest, "Invalid amount")
		return
	}

	// 更新订单状态
	order.Status = req.Status
	order.TransactionID = req.TransactionID
	if err := config.DB.Save(&order).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	ResponseSuccess(c, gin.H{
		"message":        "Payment callback received",
		"order_id":       req.OrderID,
		"transaction_id": req.TransactionID,
		"status":         req.Status,
	})
}
