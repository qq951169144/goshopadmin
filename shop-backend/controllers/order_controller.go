package controllers

import (
	"fmt"
	"net/http"
	"time"

	"shop-backend/config"
	"shop-backend/models"

	"github.com/gin-gonic/gin"
)

// 订单创建请求结构
type CreateOrderRequest struct {
	Items []CartItem `json:"items" binding:"required"`
}

// 订单结构
type Order struct {
	OrderID    string    `json:"order_id"`
	Amount     float64   `json:"amount"`
	PaymentURL string    `json:"payment_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// 创建订单
func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 计算订单金额
	var amount float64
	for _, item := range req.Items {
		amount += item.Price * float64(item.Quantity)
	}

	// 生成订单ID
	// 获取当天订单数量
	var count int64
	today := time.Now().Format("20060102")
	config.DB.Model(&models.Order{}).Where("order_id LIKE ?", "ORD"+today+"%").Count(&count)
	orderID := fmt.Sprintf("ORD%s%04d", today, count+1)

	// 生成支付URL
	paymentURL := fmt.Sprintf("/api/payment/fake-pay?order_id=%s", orderID)

	// 开始事务
	tx := config.DB.Begin()

	// 创建订单
	order := models.Order{
		OrderNo:       orderID,
		CustomerID:    int(userID.(uint)),
		MerchantID:    1, // 默认商户ID
		TotalAmount:   amount,
		Status:        "pending",
		AddressID:     1, // 默认地址ID
		PaymentMethod: "fake",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		ResponseError(c, http.StatusInternalServerError, "Failed to create order")
		return
	}

	// 创建订单项
	for _, item := range req.Items {
		orderItem := models.OrderItem{
			OrderID:       int(order.ID),
			ProductID:     int(item.ProductID),
			SkuID:         0,    // 默认SKU ID
			ProductName:   "",   // 产品名称
			SkuAttributes: "{}", // SKU属性
			Price:         item.Price,
			Quantity:      item.Quantity,
			TotalAmount:   item.Price * float64(item.Quantity),
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			ResponseError(c, http.StatusInternalServerError, "Failed to create order item")
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	ResponseSuccess(c, Order{
		OrderID:    orderID,
		Amount:     amount,
		PaymentURL: paymentURL,
		Status:     "pending",
		CreatedAt:  order.CreatedAt,
	})
}

// 获取订单详情
func GetOrderDetail(c *gin.Context) {
	orderID := c.Param("id")

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 从数据库获取订单详情
	var order models.Order
	result := config.DB.Where("order_no = ? AND customer_id = ?", orderID, userID).Preload("Items").First(&order)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "Order not found")
		return
	}

	// 转换为前端需要的格式
	var items []CartItem
	for _, item := range order.Items {
		items = append(items, CartItem{
			ProductID: uint(item.ProductID),
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       "", // 订单项中没有SKU字段
		})
	}

	ResponseSuccess(c, gin.H{
		"order_id":   order.OrderNo,
		"amount":     order.TotalAmount,
		"status":     order.Status,
		"items":      items,
		"created_at": order.CreatedAt,
		"paid_at":    order.UpdatedAt,
	})
}
