package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ProductPO 商品持久化对象
type ProductPO struct {
	ID        int            `gorm:"primaryKey;type:int;autoIncrement"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	Status    string         `gorm:"column:status;type:enum('active','inactive');not null;default:active"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	SKUs      []SKUPo        `gorm:"foreignKey:ProductID"`
}

// TableName 商品表名
func (ProductPO) TableName() string { return "products" }

// SKUPo SKU持久化对象
type SKUPo struct {
	ID         int             `gorm:"primaryKey;type:int;autoIncrement"`
	ProductID  int             `gorm:"column:product_id;type:int;not null;index"`
	Attrs      string          `gorm:"column:attrs;type:varchar(500)"`
	Price      decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null"`
	Stock      int             `gorm:"column:stock;type:int;not null;default:0"`
	MerchantID int             `gorm:"column:merchant_id;type:int;not null"`
}

// TableName SKU表名
func (SKUPo) TableName() string { return "product_skus" }
