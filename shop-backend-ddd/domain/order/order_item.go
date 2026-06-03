package order

import "github.com/shopspring/decimal"

// OrderItem 订单项值对象
// 不可变，创建后不能修改
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
// 自动计算总金额 = 单价 × 数量
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

// Getter 方法
func (i OrderItem) ProductID() int              { return i.productID }
func (i OrderItem) SkuID() int                  { return i.skuID }
func (i OrderItem) ProductName() string          { return i.productName }
func (i OrderItem) SkuAttrs() string             { return i.skuAttrs }
func (i OrderItem) Price() decimal.Decimal       { return i.price }
func (i OrderItem) Quantity() int                { return i.quantity }
func (i OrderItem) TotalAmount() decimal.Decimal { return i.totalAmount }
