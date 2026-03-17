package models

import (
	"time"
)

// Cart 购物车模型
type Cart struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	UserID    int        `json:"user_id" gorm:"index"`
	SessionID string     `json:"session_id" gorm:"index"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Items     []CartItem `json:"items" gorm:"foreignKey:CartID"`
}

// CartItem 购物车项模型
type CartItem struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	CartID    int       `json:"cart_id" gorm:"index"`
	ProductID int       `json:"product_id" gorm:"index"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`
	Price     float64   `json:"price" gorm:"not null"`
	SKU       string    `json:"sku"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Cart) TableName() string {
	return "carts"
}

// TableName 设置表名
func (CartItem) TableName() string {
	return "cart_items"
}
