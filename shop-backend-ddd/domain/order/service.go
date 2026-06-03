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

// ProductQuerier 商品查询接口（由商品聚合提供）
// 注意：接口由消费者（订单服务）定义，不是由提供者定义
type ProductQuerier interface {
	// FindProductAndSKU 查询商品和SKU信息
	FindProductAndSKU(ctx context.Context, productID, skuID int) (name string, skuAttrs string, price decimal.Decimal, stock int, err error)
	// DeductStockTx 扣减库存（在事务中执行）
	DeductStockTx(ctx context.Context, tx Transaction, skuID int, quantity int) error
	// RestoreStockTx 恢复库存（取消订单时调用）
	RestoreStockTx(ctx context.Context, tx Transaction, skuID int, quantity int) error
}

// CustomerQuerier 客户查询接口（由客户聚合提供）
type CustomerQuerier interface {
	// VerifyAddress 验证地址是否属于该客户，返回商户ID
	VerifyAddress(ctx context.Context, customerID, addressID int) (merchantID int, err error)
}

// OrderService 订单领域服务
// 编排订单相关的业务流程，但不包含业务规则（规则在实体内部）
type OrderService struct {
	orders    OrderRepository
	products  ProductQuerier
	customers CustomerQuerier
	txManager TransactionManager
}

// NewOrderService 创建订单服务（构造器注入）
func NewOrderService(
	orders OrderRepository,
	products ProductQuerier,
	customers CustomerQuerier,
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
// 编排流程：验证地址 → 验证商品/库存 → 创建订单实体 → 持久化
func (s *OrderService) CreateOrder(ctx context.Context, customerID, addressID int, items []ItemInput, remark string) (*Order, error) {
	// 1. 验证地址，获取商户ID
	merchantID, err := s.customers.VerifyAddress(ctx, customerID, addressID)
	if err != nil {
		return nil, fmt.Errorf("验证地址失败: %w", err)
	}

	// 2. 开启事务
	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}

	// 3. 验证商品、SKU、库存，构建订单项
	orderItems := make([]OrderItem, 0, len(items))
	for _, input := range items {
		name, skuAttrs, price, stock, err := s.products.FindProductAndSKU(ctx, input.ProductID, input.SkuID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("查询商品失败: %w", err)
		}

		if stock < input.Quantity {
			tx.Rollback()
			return nil, ErrStockInsufficient
		}

		// 扣减库存
		if err := s.products.DeductStockTx(ctx, tx, input.SkuID, input.Quantity); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("扣减库存失败: %w", err)
		}

		// 创建订单项值对象
		orderItem, err := NewOrderItem(input.ProductID, input.SkuID, name, skuAttrs, price, input.Quantity)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		orderItems = append(orderItems, orderItem)
	}

	// 4. 创建订单实体（业务规则在 NewOrder 工厂方法中）
	order, err := NewOrder(customerID, merchantID, addressID, orderItems, remark)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. 持久化
	if err := s.orders.Save(ctx, tx, order); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}

	// 6. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return order, nil
}

// CancelOrder 取消订单
// 业务规则在 Order.Cancel() 实体方法中
func (s *OrderService) CancelOrder(ctx context.Context, orderNo string, customerID int) error {
	// 1. 查找订单
	order, err := s.orders.FindByOrderNoAndCustomer(ctx, orderNo, customerID)
	if err != nil {
		return ErrOrderNotFound
	}

	// 2. 执行取消（业务规则在实体内部）
	if err := order.Cancel(); err != nil {
		return err
	}

	// 3. 开启事务
	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}

	// 4. 恢复库存
	for _, item := range order.Items() {
		if err := s.products.RestoreStockTx(ctx, tx, item.SkuID(), item.Quantity()); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 5. 保存订单
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
