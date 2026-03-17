package models

import (
	"time"
)

// ProductSpecification 商品规格表
type ProductSpecification struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	ProductID int       `json:"product_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"size:50;not null"`
	Sort      int       `json:"sort" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Values []ProductSpecificationValue `json:"values" gorm:"foreignKey:SpecID"`
}

// TableName 设置表名
func (ProductSpecification) TableName() string {
	return "product_specifications"
}
