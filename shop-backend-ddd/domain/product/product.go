package product

import "github.com/shopspring/decimal"

// Product 商品实体
type Product struct {
	id          int
	name        string
	description string
	price       decimal.Decimal
	status      string
}

// NewProduct 创建商品实体
func NewProduct(id int, name string, price decimal.Decimal) *Product {
	return &Product{id: id, name: name, price: price, status: "active"}
}

// ReconstructProduct 从数据库重建商品
func ReconstructProduct(id int, name, description string, price decimal.Decimal, status string) *Product {
	return &Product{id: id, name: name, description: description, price: price, status: status}
}

func (p *Product) ID() int                          { return p.id }
func (p *Product) Name() string                     { return p.name }
func (p *Product) Description() string              { return p.description }
func (p *Product) Price() decimal.Decimal           { return p.price }
func (p *Product) Status() string                   { return p.status }

// SKU 商品SKU值对象
type SKU struct {
	id        int
	productID int
	skuCode   string
	price     decimal.Decimal
	stock     int
	attrs     string
}

// NewSKU 创建SKU
func NewSKU(id, productID int, skuCode string, price decimal.Decimal, stock int, attrs string) SKU {
	return SKU{id: id, productID: productID, skuCode: skuCode, price: price, stock: stock, attrs: attrs}
}

func (s SKU) ID() int                        { return s.id }
func (s SKU) ProductID() int                 { return s.productID }
func (s SKU) SkuCode() string                { return s.skuCode }
func (s SKU) Price() decimal.Decimal         { return s.price }
func (s SKU) Stock() int                     { return s.stock }
func (s SKU) Attrs() string                  { return s.attrs }
