package models

import (
	"time"
)

// Order 订单模型
type Order struct {
	ID            int          `json:"id" gorm:"primaryKey"`
	OrderNo       string       `json:"order_no" gorm:"size:32;unique;not null"`
	CustomerID    int          `json:"customer_id" gorm:"not null;index"`
	MerchantID    int          `json:"merchant_id" gorm:"not null;index"`
	TotalAmount   float64      `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Status        string       `json:"status" gorm:"size:191;not null;default:pending;index"`
	AddressID     int          `json:"address_id" gorm:"not null;index"`
	CreatedAt     time.Time    `json:"created_at" gorm:"type:datetime(3)"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"type:datetime(3)"`
	PaymentMethod string       `json:"payment_method" gorm:"type:longtext"`
	TransactionID string       `json:"transaction_id" gorm:"type:longtext"`
	PaidAt        *time.Time   `json:"paid_at"`
	ShippedAt     *time.Time   `json:"shipped_at"`
	DeliveredAt   *time.Time   `json:"delivered_at"`
	CancelledAt   *time.Time   `json:"cancelled_at"`
	Items         []OrderItem  `json:"items" gorm:"foreignKey:OrderID"`
}

// OrderItem 订单项模型
type OrderItem struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	OrderID       int       `json:"order_id" gorm:"not null;index"`
	ProductID     int       `json:"product_id" gorm:"not null;index"`
	SkuID         int       `json:"sku_id"`
	ProductName   string    `json:"product_name" gorm:"size:100;not null"`
	SkuAttributes string    `json:"sku_attributes" gorm:"type:json"`
	Price         float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	Quantity      int       `json:"quantity" gorm:"not null"`
	TotalAmount   float64   `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"type:datetime(3)"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"type:datetime"`
}

// TableName 设置表名
func (Order) TableName() string {
	return "orders"
}

// TableName 设置表名
func (OrderItem) TableName() string {
	return "order_items"
}
