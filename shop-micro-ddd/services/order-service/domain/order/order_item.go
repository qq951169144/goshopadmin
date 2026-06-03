package order

import "github.com/shopspring/decimal"

// OrderItem 订单项值对象
type OrderItem struct {
	productID   int
	skuID       int
	productName string
	skuAttrs    string
	price       decimal.Decimal
	quantity    int
	totalAmount decimal.Decimal
}

// NewOrderItem 创建订单项
func NewOrderItem(productID, skuID int, productName, skuAttrs string, price decimal.Decimal, quantity int) (OrderItem, error) {
	if productID <= 0 {
		return OrderItem{}, ErrInvalidProductID
	}
	if skuID <= 0 {
		return OrderItem{}, ErrInvalidSKUID
	}
	if quantity <= 0 {
		return OrderItem{}, ErrInvalidQuantity
	}
	return OrderItem{
		productID:   productID,
		skuID:       skuID,
		productName: productName,
		skuAttrs:    skuAttrs,
		price:       price,
		quantity:    quantity,
		totalAmount: price.Mul(decimal.NewFromInt(int64(quantity))),
	}, nil
}

// ProductID 获取商品ID
func (i OrderItem) ProductID() int { return i.productID }

// SkuID 获取SKU ID
func (i OrderItem) SkuID() int { return i.skuID }

// ProductName 获取商品名称
func (i OrderItem) ProductName() string { return i.productName }

// SkuAttrs 获取SKU属性
func (i OrderItem) SkuAttrs() string { return i.skuAttrs }

// Price 获取单价
func (i OrderItem) Price() decimal.Decimal { return i.price }

// Quantity 获取数量
func (i OrderItem) Quantity() int { return i.quantity }

// TotalAmount 获取小计金额
func (i OrderItem) TotalAmount() decimal.Decimal { return i.totalAmount }
