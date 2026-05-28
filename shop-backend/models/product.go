package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Product 商品模型
type Product struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`             // 商品简介
	Detail      string    `json:"detail" gorm:"type:text"` // 商品详情富文本
	Price       decimal.Decimal   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	IsActivity  int       `json:"is_activity" gorm:"type:tinyint;default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 设置表名
func (Product) TableName() string {
	return "products"
}
