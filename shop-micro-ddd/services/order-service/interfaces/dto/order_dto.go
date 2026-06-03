package dto

import "github.com/shopspring/decimal"

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	CustomerID int            `json:"customer_id" binding:"required"`
	AddressID  int            `json:"address_id" binding:"required"`
	Items      []OrderItemInput `json:"items" binding:"required,min=1"`
	Remark     string         `json:"remark"`
}

// OrderItemInput 订单项输入
type OrderItemInput struct {
	ProductID int `json:"product_id" binding:"required"`
	SkuID     int `json:"sku_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	ID          int               `json:"id"`
	OrderNo     string            `json:"order_no"`
	CustomerID  int               `json:"customer_id"`
	MerchantID  int               `json:"merchant_id"`
	TotalAmount decimal.Decimal   `json:"total_amount"`
	Status      string            `json:"status"`
	AddressID   int               `json:"address_id"`
	Remark      string            `json:"remark"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   string            `json:"created_at"`
	CancelledAt *string           `json:"cancelled_at,omitempty"`
}

// OrderItemResponse 订单项响应
type OrderItemResponse struct {
	ProductID   int             `json:"product_id"`
	SkuID       int             `json:"sku_id"`
	ProductName string          `json:"product_name"`
	SkuAttrs    string          `json:"sku_attrs"`
	Price       decimal.Decimal `json:"price"`
	Quantity    int             `json:"quantity"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}
