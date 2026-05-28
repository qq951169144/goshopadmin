package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Cart 购物车模型
type Cart struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	CustomerID int        `json:"customer_id" gorm:"index"`
	SessionID  string     `json:"session_id" gorm:"index"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Items      []CartItem `json:"items" gorm:"foreignKey:CartID"`
}

// CartItem 购物车项模型
type CartItem struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	CartID    int       `json:"cart_id" gorm:"index"`
	ProductID int       `json:"product_id" gorm:"index"`
	SkuID     int       `json:"sku_id" gorm:"index"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`
	Price     decimal.Decimal   `json:"price" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// 关联关系（GORM预加载使用，非外键约束）
	Product Product    `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Sku     ProductSku `json:"sku,omitempty" gorm:"foreignKey:SkuID"`
}

// TableName 设置表名
func (Cart) TableName() string {
	return "carts"
}

// TableName 设置表名
func (CartItem) TableName() string {
	return "cart_items"
}
