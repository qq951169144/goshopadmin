package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderPO 订单持久化对象
type OrderPO struct {
	ID          int            `gorm:"primaryKey;type:int;autoIncrement"`
	OrderNo     string         `gorm:"column:order_no;type:varchar(64);uniqueIndex;not null"`
	CustomerID  int            `gorm:"column:customer_id;type:int;not null"`
	MerchantID  int            `gorm:"column:merchant_id;type:int;not null"`
	TotalAmount decimal.Decimal `gorm:"column:total_amount;type:decimal(10,2);not null"`
	Status      string         `gorm:"column:status;type:enum('pending','paid','shipped','completed','cancelled');not null;default:pending"`
	AddressID   int            `gorm:"column:address_id;type:int;not null"`
	Remark      string         `gorm:"column:remark;type:varchar(500)"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	CancelledAt *time.Time     `gorm:"column:cancelled_at"`
	Items       []OrderItemPO  `gorm:"foreignKey:OrderID"`
}

// TableName 订单表名
func (OrderPO) TableName() string { return "orders" }

// OrderItemPO 订单项持久化对象
type OrderItemPO struct {
	ID          int             `gorm:"primaryKey;type:int;autoIncrement"`
	OrderID     int             `gorm:"column:order_id;type:int;not null;index"`
	ProductID   int             `gorm:"column:product_id;type:int;not null"`
	SkuID       int             `gorm:"column:sku_id;type:int;not null"`
	ProductName string          `gorm:"column:product_name;type:varchar(255);not null"`
	SkuAttrs    string          `gorm:"column:sku_attrs;type:varchar(500)"`
	Price       decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null"`
	Quantity    int             `gorm:"column:quantity;type:int;not null"`
	TotalAmount decimal.Decimal `gorm:"column:total_amount;type:decimal(10,2);not null"`
}

// TableName 订单项表名
func (OrderItemPO) TableName() string { return "order_items" }
