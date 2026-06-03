package order

import (
	"context"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 手写 Mock
// ============================================================

// MockTransaction Mock 事务
type MockTransaction struct {
	Committed  bool
	RolledBack bool
}

func (t *MockTransaction) Commit() error   { t.Committed = true; return nil }
func (t *MockTransaction) Rollback() error { t.RolledBack = true; return nil }

// MockTransactionManager Mock 事务管理器
type MockTransactionManager struct{ Tx *MockTransaction }

func (m *MockTransactionManager) BeginTx(ctx context.Context) (Transaction, error) {
	m.Tx = &MockTransaction{}
	return m.Tx, nil
}

// MockOrderRepository Mock 订单仓库
type MockOrderRepository struct{ SavedOrder *Order }

func (m *MockOrderRepository) FindByOrderNoAndCustomer(ctx context.Context, orderNo string, customerID int) (*Order, error) {
	return nil, ErrOrderNotFound
}

func (m *MockOrderRepository) Save(ctx context.Context, tx Transaction, order *Order) error {
	m.SavedOrder = order
	return nil
}

// MockProductServiceProvider Mock 商品服务（模拟 HTTP 调用）
type MockProductServiceProvider struct {
	Name       string
	SkuAttrs   string
	Price      decimal.Decimal
	Stock      int
	deductErr  error
	restoreErr error
}

func (m *MockProductServiceProvider) GetProductAndSKU(ctx context.Context, productID, skuID int) (string, string, decimal.Decimal, int, error) {
	return m.Name, m.SkuAttrs, m.Price, m.Stock, nil
}

func (m *MockProductServiceProvider) DeductStock(ctx context.Context, skuID int, quantity int) error {
	return m.deductErr
}

func (m *MockProductServiceProvider) RestoreStock(ctx context.Context, skuID int, quantity int) error {
	return m.restoreErr
}

// MockCustomerServiceProvider Mock 客户服务（模拟 HTTP 调用）
type MockCustomerServiceProvider struct {
	MerchantID int
	VerifyErr  error
}

func (m *MockCustomerServiceProvider) VerifyAddress(ctx context.Context, customerID, addressID int) (int, error) {
	if m.VerifyErr != nil {
		return 0, m.VerifyErr
	}
	return m.MerchantID, nil
}

// ============================================================
// 测试用例
// ============================================================

// TestOrderService_CreateOrder_Success 创建订单成功
func TestOrderService_CreateOrder_Success(t *testing.T) {
	mockOrders := &MockOrderRepository{}
	mockProducts := &MockProductServiceProvider{
		Name: "测试商品", SkuAttrs: `{"颜色":"红色"}`,
		Price: decimal.NewFromFloat(99.90), Stock: 100,
	}
	mockCustomers := &MockCustomerServiceProvider{MerchantID: 1}
	mockTxManager := &MockTransactionManager{}

	svc := NewOrderService(mockOrders, mockProducts, mockCustomers, mockTxManager)
	items := []ItemInput{{ProductID: 1, SkuID: 1, Quantity: 2}}

	order, err := svc.CreateOrder(context.Background(), 1, 1, items, "测试备注")

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, decimal.NewFromFloat(199.80), order.TotalAmount())
	assert.Equal(t, StatusPending, order.Status())
	assert.True(t, mockTxManager.Tx.Committed)
}

// TestOrderService_CreateOrder_StockInsufficient 库存不足
func TestOrderService_CreateOrder_StockInsufficient(t *testing.T) {
	mockOrders := &MockOrderRepository{}
	mockProducts := &MockProductServiceProvider{
		Name: "测试商品", Price: decimal.NewFromFloat(99.90), Stock: 1,
	}
	mockCustomers := &MockCustomerServiceProvider{MerchantID: 1}
	mockTxManager := &MockTransactionManager{}

	svc := NewOrderService(mockOrders, mockProducts, mockCustomers, mockTxManager)
	items := []ItemInput{{ProductID: 1, SkuID: 1, Quantity: 2}}

	order, err := svc.CreateOrder(context.Background(), 1, 1, items, "")

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrStockInsufficient, err)
}

// TestOrderService_CreateOrder_DeductStockFailed_SagaCompensation 扣减库存失败触发 Saga 补偿
func TestOrderService_CreateOrder_DeductStockFailed_SagaCompensation(t *testing.T) {
	mockOrders := &MockOrderRepository{}
	mockProducts := &MockProductServiceProvider{
		Name: "测试商品", Price: decimal.NewFromFloat(99.90), Stock: 100,
		deductErr: fmt.Errorf("扣减库存服务异常"),
	}
	mockCustomers := &MockCustomerServiceProvider{MerchantID: 1}
	mockTxManager := &MockTransactionManager{}

	svc := NewOrderService(mockOrders, mockProducts, mockCustomers, mockTxManager)
	items := []ItemInput{{ProductID: 1, SkuID: 1, Quantity: 2}}

	order, err := svc.CreateOrder(context.Background(), 1, 1, items, "")

	assert.Error(t, err)
	assert.Nil(t, order)
}

// TestOrder_Cancel_Success 取消订单成功
func TestOrder_Cancel_Success(t *testing.T) {
	items := []OrderItem{mustNewItem(1, 1, "商品A", "{}", decimal.NewFromFloat(50), 2)}
	order, _ := NewOrder(1, 1, 1, items, "")

	err := order.Cancel()

	assert.NoError(t, err)
	assert.Equal(t, StatusCancelled, order.Status())
}

// mustNewItem 辅助函数：创建订单项，出错则 panic
func mustNewItem(pid, sid int, name, attrs string, price decimal.Decimal, qty int) OrderItem {
	item, err := NewOrderItem(pid, sid, name, attrs, price, qty)
	if err != nil {
		panic(err)
	}
	return item
}
