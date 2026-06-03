package order

import "errors"

var (
	// ErrInvalidCustomerID 客户ID无效
	ErrInvalidCustomerID = errors.New("客户ID无效")
	// ErrInvalidAddressID 地址ID无效
	ErrInvalidAddressID = errors.New("地址ID无效")
	// ErrEmptyOrderItems 订单项不能为空
	ErrEmptyOrderItems = errors.New("订单项不能为空")
	// ErrInvalidProductID 商品ID无效
	ErrInvalidProductID = errors.New("商品ID无效")
	// ErrInvalidSKUID SKU ID无效
	ErrInvalidSKUID = errors.New("SKU ID无效")
	// ErrInvalidQuantity 商品数量无效
	ErrInvalidQuantity = errors.New("商品数量无效")
	// ErrInvalidStatusForCancel 当前订单状态不允许取消
	ErrInvalidStatusForCancel = errors.New("当前订单状态不允许取消")
	// ErrInvalidStatusForConfirm 当前订单状态不允许确认收货
	ErrInvalidStatusForConfirm = errors.New("当前订单状态不允许确认收货")
	// ErrInvalidStatusForPay 当前订单状态不允许支付
	ErrInvalidStatusForPay = errors.New("当前订单状态不允许支付")
	// ErrOrderNotFound 订单不存在
	ErrOrderNotFound = errors.New("订单不存在")
	// ErrStockInsufficient 库存不足
	ErrStockInsufficient = errors.New("库存不足")
	// ErrProductServiceUnavailable 商品服务不可用
	ErrProductServiceUnavailable = errors.New("商品服务不可用")
	// ErrCustomerServiceUnavailable 客户服务不可用
	ErrCustomerServiceUnavailable = errors.New("客户服务不可用")
)
