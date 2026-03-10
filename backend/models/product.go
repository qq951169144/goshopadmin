package models

import (
	"time"
)

// Product 商品模型
type Product struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Detail      string    `json:"detail" gorm:"type:text"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock       int       `json:"stock" gorm:"not null"`
	CategoryID  int       `json:"category_id" gorm:"not null"`
	MerchantID  int       `json:"merchant_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Category ProductCategory `json:"category" gorm:"foreignKey:CategoryID"`
	Images   []ProductImage  `json:"images" gorm:"foreignKey:ProductID"`
	SKUs     []ProductSKU    `json:"skus" gorm:"foreignKey:ProductID"`
}

// ProductCategory 商品分类模型
type ProductCategory struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	MerchantID int       `json:"merchant_id" gorm:"not null"`
	Name       string    `json:"name" gorm:"size:50;not null"`
	ParentID   int       `json:"parent_id" gorm:"default:0"`
	Level      int       `json:"level" gorm:"default:1"`
	Sort       int       `json:"sort" gorm:"default:0"`
	Status     string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Products []Product `json:"products" gorm:"foreignKey:CategoryID"`
}

// ProductSKU 商品SKU模型
type ProductSKU struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ProductID  int       `json:"product_id" gorm:"not null"`
	MerchantID int       `json:"merchant_id" gorm:"not null"`
	SKUCode    string    `json:"sku_code" gorm:"size:50;not null"`
	Attributes string    `json:"attributes" gorm:"type:json;not null"` // JSON格式存储SKU属性
	Price      float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock      int       `json:"stock" gorm:"not null"`
	Status     string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}

// ProductImage 商品图片模型
type ProductImage struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	ProductID int       `json:"product_id" gorm:"not null"`
	ImageURL  string    `json:"image_url" gorm:"size:255;not null"`
	IsMain    bool      `json:"is_main" gorm:"default:false"`
	Sort      int       `json:"sort" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}
