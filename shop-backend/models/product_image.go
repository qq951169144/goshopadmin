package models

import (
	"time"
)

// ProductImage 商品图片模型
type ProductImage struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	ProductID int       `json:"product_id" gorm:"not null;index"`
	ImageURL  string    `json:"image_url" gorm:"size:500;not null"`
	IsMain    bool      `json:"is_main" gorm:"default:false"`      // 是否主图
	Sort      int       `json:"sort" gorm:"column:sort;default:0"` // 排序
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 设置表名
func (ProductImage) TableName() string {
	return "product_images"
}
