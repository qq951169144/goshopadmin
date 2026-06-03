package models

import (
	"github.com/shopspring/decimal"
)

// ProductPO 商品持久化对象
type ProductPO struct {
	ID          int             `gorm:"primaryKey;autoIncrement"`
	Name        string          `gorm:"column:name;size:100;not null"`
	Description string          `gorm:"column:description;type:longtext"`
	Price       decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null"`
	MerchantID  int             `gorm:"column:merchant_id;not null"`
	Status      string          `gorm:"column:status;type:enum('active','inactive');default:'active'"`
}

func (ProductPO) TableName() string { return "products" }

// SkuPO SKU持久化对象
type SkuPO struct {
	ID         int             `gorm:"primaryKey;autoIncrement"`
	ProductID  int             `gorm:"column:product_id;not null;index"`
	SkuCode    string          `gorm:"column:sku_code;size:50;not null;uniqueIndex"`
	Price      decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null"`
	Stock      int             `gorm:"column:stock;not null"`
	Attributes string          `gorm:"column:attributes;type:json"`
	Status     string          `gorm:"column:status;type:enum('active','inactive');default:'active'"`
}

func (SkuPO) TableName() string { return "product_skus" }
