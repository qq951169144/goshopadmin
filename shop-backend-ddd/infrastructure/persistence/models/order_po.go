package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderPO 订单持久化对象（对应数据库表）
type OrderPO struct {
	ID          int             `gorm:"primaryKey;autoIncrement"`
	OrderNo     string          `gorm:"column:order_no;size:64;uniqueIndex;not null"`
	CustomerID  int             `gorm:"column:customer_id;not null;index"`
	MerchantID  int             `gorm:"column:merchant_id;not null"`
	TotalAmount decimal.Decimal `gorm:"column:total_amount;type:decimal(10,2);not null"`
	Status      string          `gorm:"column:status;type:enum('pending','paid','shipped','completed','cancelled');default:'pending'"`
	AddressID   int             `gorm:"column:address_id;not null"`
	Remark      string          `gorm:"column:remark;size:500"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime"`
	CancelledAt *time.Time      `gorm:"column:cancelled_at"`
}

func (OrderPO) TableName() string { return "orders" }

// OrderItemPO 订单项持久化对象
type OrderItemPO struct {
	ID          int             `gorm:"primaryKey;autoIncrement"`
	OrderID     int             `gorm:"column:order_id;not null;index"`
	ProductID   int             `gorm:"column:product_id;not null"`
	SkuID       int             `gorm:"column:sku_id;not null"`
	ProductName string          `gorm:"column:product_name;size:100;not null"`
	SkuAttrs    string          `gorm:"column:sku_attributes;type:json"`
	Price       decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null"`
	Quantity    int             `gorm:"column:quantity;not null"`
	TotalAmount decimal.Decimal `gorm:"column:total_amount;type:decimal(10,2);not null"`
}

func (OrderItemPO) TableName() string { return "order_items" }
