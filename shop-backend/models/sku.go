package models

import (
	"time"
)

// ProductSKU 商品SKU模型
type ProductSKU struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	ProductID     int       `json:"product_id" gorm:"not null;index"`
	SKUCode       string    `json:"sku_code" gorm:"size:50;not null;uniqueIndex"`
	Price         float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	OriginalPrice float64   `json:"original_price" gorm:"type:decimal(10,2);default:0"`
	Stock         int       `json:"stock" gorm:"not null;default:0"`
	Status        string    `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 设置表名
func (ProductSKU) TableName() string {
	return "product_skus"
}
