package models

import (
	"time"
)

// ProductSKU 商品SKU模型
type ProductSKU struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	ProductID uint    `json:"product_id" gorm:"not null;index"`
	Name      string  `json:"name" gorm:"size:100;not null"`
	Price     float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock     int     `json:"stock" gorm:"not null;default:0"`
	Image     string  `json:"image" gorm:"size:255"`
	Attributes string `json:"attributes" gorm:"type:json"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 设置表名
func (ProductSKU) TableName() string {
	return "product_skus"
}
