package order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatus 订单状态类型
type OrderStatus string

const (
	// StatusPending 待支付
	StatusPending OrderStatus = "pending"
	// StatusPaid 已支付
	StatusPaid OrderStatus = "paid"
	// StatusShipped 已发货
	StatusShipped OrderStatus = "shipped"
	// StatusCompleted 已完成
	StatusCompleted OrderStatus = "completed"
	// StatusCancelled 已取消
	StatusCancelled OrderStatus = "cancelled"
)

// Order 订单实体（充血模型）
// 业务规则封装在实体内部，外部不能直接修改状态
type Order struct {
	id          int
	orderNo     string
	customerID  int
	merchantID  int
	totalAmount decimal.Decimal
	status      OrderStatus
	items       []OrderItem
	addressID   int
	remark      string
	createdAt   time.Time
	cancelledAt *time.Time
}

// NewOrder 创建新订单（工厂方法）
// 封装订单创建的业务规则：必须有客户、地址、订单项
func NewOrder(customerID, merchantID, addressID int, items []OrderItem, remark string) (*Order, error) {
	if customerID <= 0 {
		return nil, ErrInvalidCustomerID
	}
	if addressID <= 0 {
		return nil, ErrInvalidAddressID
	}
	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}

	// 计算总金额
	var total decimal.Decimal
	for _, item := range items {
		total = total.Add(item.TotalAmount())
	}

	// 生成订单号
	now := time.Now()
	orderNo := fmt.Sprintf("ORD%s%04d", now.Format("20060102"), now.UnixNano()%10000)

	return &Order{
		orderNo:     orderNo,
		customerID:  customerID,
		merchantID:  merchantID,
		totalAmount: total,
		status:      StatusPending,
		items:       items,
		addressID:   addressID,
		remark:      remark,
		createdAt:   now,
	}, nil
}

// ReconstructFromDB 从数据库重建订单实体
// 只有基础设施层调用，用于将 PO 转换为领域实体
func ReconstructFromDB(id int, orderNo string, customerID, merchantID, addressID int,
	totalAmount decimal.Decimal, status OrderStatus, items []OrderItem,
	remark string, createdAt time.Time, cancelledAt *time.Time) *Order {
	return &Order{
		id:          id,
		orderNo:     orderNo,
		customerID:  customerID,
		merchantID:  merchantID,
		totalAmount: totalAmount,
		status:      status,
		items:       items,
		addressID:   addressID,
		remark:      remark,
		createdAt:   createdAt,
		cancelledAt: cancelledAt,
	}
}

// Cancel 取消订单
// 业务规则：只有 pending 和 paid 状态可以取消
func (o *Order) Cancel() error {
	if o.status != StatusPending && o.status != StatusPaid {
		return ErrInvalidStatusForCancel
	}
	o.status = StatusCancelled
	now := time.Now()
	o.cancelledAt = &now
	return nil
}

// Confirm 确认收货
// 业务规则：只有 shipped 状态可以确认
func (o *Order) Confirm() error {
	if o.status != StatusShipped {
		return ErrInvalidStatusForConfirm
	}
	o.status = StatusCompleted
	return nil
}

// Pay 标记为已支付
// 业务规则：只有 pending 状态可以支付
func (o *Order) Pay() error {
	if o.status != StatusPending {
		return ErrInvalidStatusForPay
	}
	o.status = StatusPaid
	return nil
}

// CanCancel 判断是否可取消
func (o *Order) CanCancel() bool {
	return o.status == StatusPending || o.status == StatusPaid
}

// Getter 方法（实体字段私有，通过 getter 访问）
func (o *Order) ID() int                          { return o.id }
func (o *Order) OrderNo() string                  { return o.orderNo }
func (o *Order) CustomerID() int                  { return o.customerID }
func (o *Order) MerchantID() int                  { return o.merchantID }
func (o *Order) TotalAmount() decimal.Decimal     { return o.totalAmount }
func (o *Order) Status() OrderStatus              { return o.status }
func (o *Order) Items() []OrderItem               { return o.items }
func (o *Order) AddressID() int                   { return o.addressID }
func (o *Order) Remark() string                   { return o.remark }
func (o *Order) CreatedAt() time.Time             { return o.createdAt }
func (o *Order) CancelledAt() *time.Time          { return o.cancelledAt }

// SetID 设置ID（由持久化层在 Save 后调用）
func (o *Order) SetID(id int) { o.id = id }
