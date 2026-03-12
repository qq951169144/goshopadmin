package models

import (
	"time"
)

// Product 商品模型
type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	SKU         string    `json:"sku" gorm:"unique"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Product) TableName() string {
	return "products"
}
