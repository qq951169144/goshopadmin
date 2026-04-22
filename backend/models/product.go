package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Product 商品模型
type Product struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:longtext"`
	Detail      string    `json:"detail" gorm:"type:text"`
	Price       decimal.Decimal   `json:"price" gorm:"type:decimal(10,2);not null"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	CategoryID  int       `json:"category_id" gorm:"not null"`
	MerchantID  int       `json:"merchant_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	IsActivity  int       `json:"is_activity" gorm:"type:tinyint;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:datetime(3)"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:datetime(3)"`

	// 关联
	Category ProductCategory `json:"category" gorm:"foreignKey:CategoryID"`
	Images   []ProductImage  `json:"images" gorm:"foreignKey:ProductID"`
	Skus     []ProductSku    `json:"skus" gorm:"foreignKey:ProductID"`
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

// ProductSku 商品SKU模型
type ProductSku struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	ProductID     int       `json:"product_id" gorm:"not null"`
	MerchantID    int       `json:"merchant_id" gorm:"not null"`
	SkuCode       string    `json:"sku_code" gorm:"size:50;not null;uniqueIndex"`
	Attributes    string    `json:"attributes" gorm:"type:json;"` // [已废弃] 请使用Specs关联表存储规格关系
	Price         decimal.Decimal   `json:"price" gorm:"type:decimal(10,2);not null"`
	OriginalPrice decimal.Decimal   `json:"original_price" gorm:"type:decimal(10,2);default:0"`
	Stock         int       `json:"stock" gorm:"not null"`
	IsActivity    int       `json:"is_activity" gorm:"type:tinyint(1);default:0"` // 0-普通SKU，1-活动专用SKU
	ActivityID    int       `json:"activity_id" gorm:"default:0"`                 // 关联活动ID
	Status        string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Product Product          `json:"product" gorm:"foreignKey:ProductID"`
	Specs   []ProductSkuSpec `json:"specs" gorm:"foreignKey:SkuID"`
}

// ProductSpecification 商品规格表
type ProductSpecification struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	ProductID int       `json:"product_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"size:50;not null"`
	Sort      int       `json:"sort" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Product Product                     `json:"product" gorm:"foreignKey:ProductID"`
	Values  []ProductSpecificationValue `json:"values" gorm:"foreignKey:SpecID"`
}

// TableName 设置表名
func (ProductSpecification) TableName() string {
	return "product_specifications"
}

// ProductSpecificationValue 规格值表
type ProductSpecificationValue struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	SpecID    int       `json:"spec_id" gorm:"not null;index"`
	Value     string    `json:"value" gorm:"size:50;not null"`
	Sort      int       `json:"sort" gorm:"default:0"`
	Status    string    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Spec ProductSpecification `json:"spec" gorm:"foreignKey:SpecID"`
}

// TableName 设置表名
func (ProductSpecificationValue) TableName() string {
	return "product_specification_values"
}

// ProductSkuSpec SKU规格关联表
type ProductSkuSpec struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	SkuID       int       `json:"sku_id" gorm:"not null;index"`
	SpecID      int       `json:"spec_id" gorm:"not null;index"`
	SpecValueID int       `json:"spec_value_id" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`

	// 关联
	Sku       ProductSku                `json:"sku" gorm:"foreignKey:SkuID"`
	Spec      ProductSpecification      `json:"spec" gorm:"foreignKey:SpecID"`
	SpecValue ProductSpecificationValue `json:"spec_value" gorm:"foreignKey:SpecValueID"`
}

// TableName 设置表名
func (ProductSkuSpec) TableName() string {
	return "product_sku_specs"
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
