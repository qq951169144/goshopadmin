package models

import (
	"time"
)

// ProductSpecificationValue 规格值表
type ProductSpecificationValue struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	SpecID    int       `json:"spec_id" gorm:"not null;index"`
	Value     string    `json:"value" gorm:"size:50;not null"`
	Sort      int       `json:"sort" gorm:"default:0"`
	Status    string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 设置表名
func (ProductSpecificationValue) TableName() string {
	return "product_specification_values"
}
