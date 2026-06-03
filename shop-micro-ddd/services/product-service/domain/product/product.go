package product

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	// ErrProductNotFound 商品不存在
	ErrProductNotFound = errors.New("商品不存在")
	// ErrSKUNotFound SKU不存在
	ErrSKUNotFound = errors.New("SKU不存在")
	// ErrStockInsufficient 库存不足
	ErrStockInsufficient = errors.New("库存不足")
)

// ProductStatus 商品状态类型
type ProductStatus string

const (
	// StatusActive 上架
	StatusActive ProductStatus = "active"
	// StatusInactive 下架
	StatusInactive ProductStatus = "inactive"
)

// Product 商品实体（充血模型）
type Product struct {
	id     int
	name   string
	status ProductStatus
	skus   []SKU
}

// SKU 商品SKU值对象
type SKU struct {
	id       int
	attrs    string
	price    decimal.Decimal
	stock    int
	merchantID int
}

// NewProduct 创建商品
func NewProduct(id int, name string, status ProductStatus, skus []SKU) *Product {
	return &Product{id: id, name: name, status: status, skus: skus}
}

// NewSKU 创建SKU
func NewSKU(id int, attrs string, price decimal.Decimal, stock int, merchantID int) SKU {
	return SKU{id: id, attrs: attrs, price: price, stock: stock, merchantID: merchantID}
}

// FindSKU 根据ID查找SKU
func (p *Product) FindSKU(skuID int) (SKU, bool) {
	for _, sku := range p.skus {
		if sku.id == skuID {
			return sku, true
		}
	}
	return SKU{}, false
}

// DeductStock 扣减库存
func (s *SKU) DeductStock(quantity int) error {
	if s.stock < quantity {
		return ErrStockInsufficient
	}
	s.stock -= quantity
	return nil
}

// RestoreStock 恢复库存
func (s *SKU) RestoreStock(quantity int) {
	s.stock += quantity
}

// Getter 方法
func (p *Product) ID() int               { return p.id }
func (p *Product) Name() string           { return p.name }
func (p *Product) Status() ProductStatus  { return p.status }
func (p *Product) SKUs() []SKU            { return p.skus }

func (s SKU) ID() int                        { return s.id }
func (s SKU) Attrs() string                  { return s.attrs }
func (s SKU) Price() decimal.Decimal         { return s.price }
func (s SKU) Stock() int                     { return s.stock }
func (s SKU) MerchantID() int                { return s.merchantID }
