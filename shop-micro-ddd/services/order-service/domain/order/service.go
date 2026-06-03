package order

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// ItemInput 创建订单项的输入参数
type ItemInput struct {
	ProductID int
	SkuID     int
	Quantity  int
}

// ProductServiceProvider 商品服务提供者接口（跨服务调用）
// 微服务版的关键区别：不再直接调用 ProductRepository，而是通过 HTTP 调用商品服务
type ProductServiceProvider interface {
	// GetProductAndSKU 获取商品和SKU信息（HTTP 调用商品服务）
	GetProductAndSKU(ctx context.Context, productID, skuID int) (name string, skuAttrs string, price decimal.Decimal, stock int, err error)
	// DeductStock 扣减库存（HTTP 调用商品服务）
	DeductStock(ctx context.Context, skuID int, quantity int) error
	// RestoreStock 恢复库存（HTTP 调用商品服务）
	RestoreStock(ctx context.Context, skuID int, quantity int) error
}

// CustomerServiceProvider 客户服务提供者接口（跨服务调用）
type CustomerServiceProvider interface {
	// VerifyAddress 验证地址（HTTP 调用客户服务）
	VerifyAddress(ctx context.Context, customerID, addressID int) (merchantID int, err error)
}

// OrderService 订单领域服务
type OrderService struct {
	orders    OrderRepository
	products  ProductServiceProvider
	customers CustomerServiceProvider
	txManager TransactionManager
}

// NewOrderService 创建订单服务
func NewOrderService(
	orders OrderRepository,
	products ProductServiceProvider,
	customers CustomerServiceProvider,
	txManager TransactionManager,
) *OrderService {
	return &OrderService{
		orders:    orders,
		products:  products,
		customers: customers,
		txManager: txManager,
	}
}

// CreateOrder 创建订单
// 微服务版的关键区别：
// 1. 不在一个事务中扣减库存（跨服务事务用 Saga 模式）
// 2. 先扣减库存（HTTP 调用），再创建订单
// 3. 如果创建订单失败，需要补偿（恢复库存）
func (s *OrderService) CreateOrder(ctx context.Context, customerID, addressID int, items []ItemInput, remark string) (*Order, error) {
	// 1. 验证地址（HTTP 调用客户服务）
	merchantID, err := s.customers.VerifyAddress(ctx, customerID, addressID)
	if err != nil {
		return nil, fmt.Errorf("验证地址失败: %w", err)
	}

	// 2. 验证商品、扣减库存（HTTP 调用商品服务）
	// 注意：微服务版中，库存扣减是独立的服务调用，不在订单事务中
	orderItems := make([]OrderItem, 0, len(items))
	deductedSKUs := make([]struct{ skuID, qty int }, 0) // 用于补偿回滚

	for _, input := range items {
		name, skuAttrs, price, stock, err := s.products.GetProductAndSKU(ctx, input.ProductID, input.SkuID)
		if err != nil {
			// 补偿：恢复已扣减的库存
			s.compensateDeduction(ctx, deductedSKUs)
			return nil, fmt.Errorf("查询商品失败: %w", err)
		}

		if stock < input.Quantity {
			s.compensateDeduction(ctx, deductedSKUs)
			return nil, ErrStockInsufficient
		}

		// 扣减库存（HTTP 调用）
		if err := s.products.DeductStock(ctx, input.SkuID, input.Quantity); err != nil {
			s.compensateDeduction(ctx, deductedSKUs)
			return nil, fmt.Errorf("扣减库存失败: %w", err)
		}
		deductedSKUs = append(deductedSKUs, struct{ skuID, qty int }{input.SkuID, input.Quantity})

		orderItem, err := NewOrderItem(input.ProductID, input.SkuID, name, skuAttrs, price, input.Quantity)
		if err != nil {
			s.compensateDeduction(ctx, deductedSKUs)
			return nil, err
		}
		orderItems = append(orderItems, orderItem)
	}

	// 3. 创建订单实体
	orderEntity, err := NewOrder(customerID, merchantID, addressID, orderItems, remark)
	if err != nil {
		s.compensateDeduction(ctx, deductedSKUs)
		return nil, err
	}

	// 4. 持久化订单（本地事务）
	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		s.compensateDeduction(ctx, deductedSKUs)
		return nil, err
	}

	if err := s.orders.Save(ctx, tx, orderEntity); err != nil {
		tx.Rollback()
		s.compensateDeduction(ctx, deductedSKUs)
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.compensateDeduction(ctx, deductedSKUs)
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return orderEntity, nil
}

// compensateDeduction 补偿已扣减的库存（Saga 补偿模式）
func (s *OrderService) compensateDeduction(ctx context.Context, deducted []struct{ skuID, qty int }) {
	for _, d := range deducted {
		if err := s.products.RestoreStock(ctx, d.skuID, d.qty); err != nil {
			// 补偿失败，记录日志，需要人工介入
			// 实际项目中会发送到死信队列
			fmt.Printf("补偿失败: skuID=%d, qty=%d, err=%v\n", d.skuID, d.qty, err)
		}
	}
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderNo string, customerID int) error {
	order, err := s.orders.FindByOrderNoAndCustomer(ctx, orderNo, customerID)
	if err != nil {
		return ErrOrderNotFound
	}

	if err := order.Cancel(); err != nil {
		return err
	}

	// 恢复库存（HTTP 调用商品服务）
	for _, item := range order.Items() {
		if err := s.products.RestoreStock(ctx, item.SkuID(), item.Quantity()); err != nil {
			// 恢复失败，记录日志，需要人工介入
			fmt.Printf("恢复库存失败: skuID=%d, err=%v\n", item.SkuID(), err)
		}
	}

	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}

	if err := s.orders.Save(ctx, tx, order); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetOrderDetail 获取订单详情
func (s *OrderService) GetOrderDetail(ctx context.Context, orderNo string, customerID int) (*Order, error) {
	order, err := s.orders.FindByOrderNoAndCustomer(ctx, orderNo, customerID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}
