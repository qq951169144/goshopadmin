package order

import (
	"context"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 手写 Mock（不需要 mockgen，简单直接）
// ============================================================

// MockTransaction Mock 事务
type MockTransaction struct {
	Committed   bool
	RolledBack  bool
	commitErr   error
	rollbackErr error
}

func (t *MockTransaction) Commit() error {
	t.Committed = true
	return t.commitErr
}

func (t *MockTransaction) Rollback() error {
	t.RolledBack = true
	return t.rollbackErr
}

// MockTransactionManager Mock 事务管理器
type MockTransactionManager struct {
	Tx *MockTransaction
}

func (m *MockTransactionManager) BeginTx(ctx context.Context) (Transaction, error) {
	m.Tx = &MockTransaction{}
	return m.Tx, nil
}

// MockOrderRepository Mock 订单仓库
type MockOrderRepository struct {
	SavedOrder *Order
	saveErr    error
}

func (m *MockOrderRepository) FindByOrderNoAndCustomer(ctx context.Context, orderNo string, customerID int) (*Order, error) {
	return nil, ErrOrderNotFound
}

func (m *MockOrderRepository) Save(ctx context.Context, tx Transaction, order *Order) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.SavedOrder = order
	return nil
}

// MockProductQuerier Mock 商品查询
type MockProductQuerier struct {
	Name       string
	SkuAttrs   string
	Price      decimal.Decimal
	Stock      int
	findErr    error
	deductErr  error
	restoreErr error
}

func (m *MockProductQuerier) FindProductAndSKU(ctx context.Context, productID, skuID int) (string, string, decimal.Decimal, int, error) {
	if m.findErr != nil {
		return "", "", decimal.Zero, 0, m.findErr
	}
	return m.Name, m.SkuAttrs, m.Price, m.Stock, nil
}

func (m *MockProductQuerier) DeductStockTx(ctx context.Context, tx Transaction, skuID int, quantity int) error {
	return m.deductErr
}

func (m *MockProductQuerier) RestoreStockTx(ctx context.Context, tx Transaction, skuID int, quantity int) error {
	return m.restoreErr
}

// MockCustomerQuerier Mock 客户查询
type MockCustomerQuerier struct {
	MerchantID int
	verifyErr  error
}

func (m *MockCustomerQuerier) VerifyAddress(ctx context.Context, customerID, addressID int) (int, error) {
	if m.verifyErr != nil {
		return 0, m.verifyErr
	}
	return m.MerchantID, nil
}

// ============================================================
// 测试用例
// ============================================================

func TestOrderService_CreateOrder_Success(t *testing.T) {
	// Arrange
	mockOrders := &MockOrderRepository{}
	mockProducts := &MockProductQuerier{
		Name:     "测试商品",
		SkuAttrs: `{"颜色":"红色","尺码":"XL"}`,
		Price:    decimal.NewFromFloat(99.90),
		Stock:    100,
	}
	mockCustomers := &MockCustomerQuerier{MerchantID: 1}
	mockTxManager := &MockTransactionManager{}

	svc := NewOrderService(mockOrders, mockProducts, mockCustomers, mockTxManager)

	items := []ItemInput{
		{ProductID: 1, SkuID: 1, Quantity: 2},
	}

	// Act
	order, err := svc.CreateOrder(context.Background(), 1, 1, items, "测试备注")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, decimal.NewFromFloat(199.80), order.TotalAmount())
	assert.Equal(t, StatusPending, order.Status())
	assert.Equal(t, 1, order.CustomerID())
	assert.Len(t, order.Items(), 1)
	assert.True(t, mockTxManager.Tx.Committed)
	assert.NotNil(t, mockOrders.SavedOrder)
}

func TestOrderService_CreateOrder_StockInsufficient(t *testing.T) {
	// Arrange
	mockOrders := &MockOrderRepository{}
	mockProducts := &MockProductQuerier{
		Name:  "测试商品",
		Price: decimal.NewFromFloat(99.90),
		Stock: 1, // 库存只有1件
	}
	mockCustomers := &MockCustomerQuerier{MerchantID: 1}
	mockTxManager := &MockTransactionManager{}

	svc := NewOrderService(mockOrders, mockProducts, mockCustomers, mockTxManager)

	items := []ItemInput{
		{ProductID: 1, SkuID: 1, Quantity: 2}, // 请求2件
	}

	// Act
	order, err := svc.CreateOrder(context.Background(), 1, 1, items, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrStockInsufficient, err)
	assert.True(t, mockTxManager.Tx.RolledBack)
}

func TestOrderService_CreateOrder_InvalidAddress(t *testing.T) {
	// Arrange
	mockOrders := &MockOrderRepository{}
	mockProducts := &MockProductQuerier{}
	mockCustomers := &MockCustomerQuerier{
		verifyErr: fmt.Errorf("地址不存在"),
	}
	mockTxManager := &MockTransactionManager{}

	svc := NewOrderService(mockOrders, mockProducts, mockCustomers, mockTxManager)

	// Act
	order, err := svc.CreateOrder(context.Background(), 1, 999, nil, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
}

func TestOrder_Cancel_Success(t *testing.T) {
	// Arrange — 直接测试实体方法
	items := []OrderItem{
		mustNewOrderItem(1, 1, "商品A", "{}", decimal.NewFromFloat(50), 2),
	}
	order, _ := NewOrder(1, 1, 1, items, "")

	// Act
	err := order.Cancel()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, StatusCancelled, order.Status())
	assert.NotNil(t, order.CancelledAt())
}

func TestOrder_Cancel_InvalidStatus(t *testing.T) {
	// Arrange
	items := []OrderItem{
		mustNewOrderItem(1, 1, "商品A", "{}", decimal.NewFromFloat(50), 2),
	}
	order2, _ := NewOrder(1, 1, 1, items, "")
	order2.Pay()
	order2.status = StatusCompleted // 直接设置状态模拟已完成

	// Act
	err := order2.Cancel()

	// Assert
	assert.Equal(t, ErrInvalidStatusForCancel, err)
}

// mustNewOrderItem 辅助函数，创建订单项（测试用）
func mustNewOrderItem(productID, skuID int, name, skuAttrs string, price decimal.Decimal, quantity int) OrderItem {
	item, err := NewOrderItem(productID, skuID, name, skuAttrs, price, quantity)
	if err != nil {
		panic(err)
	}
	return item
}
