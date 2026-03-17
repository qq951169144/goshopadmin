package models

import (
	"time"
)

// ProductSKUSpec SKU规格关联表
type ProductSKUSpec struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	SkuID       int       `json:"sku_id" gorm:"not null;index"`
	SpecID      int       `json:"spec_id" gorm:"not null;index"`
	SpecValueID int       `json:"spec_value_id" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 设置表名
func (ProductSKUSpec) TableName() string {
	return "product_sku_specs"
}
