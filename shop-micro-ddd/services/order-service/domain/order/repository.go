package order

import "context"

// Transaction 事务接口
type Transaction interface {
	Commit() error
	Rollback() error
}

// TransactionManager 事务管理器接口
type TransactionManager interface {
	BeginTx(ctx context.Context) (Transaction, error)
}

// OrderRepository 订单仓库接口
type OrderRepository interface {
	// FindByOrderNoAndCustomer 根据订单号和客户ID查询订单
	FindByOrderNoAndCustomer(ctx context.Context, orderNo string, customerID int) (*Order, error)
	// Save 保存订单
	Save(ctx context.Context, tx Transaction, order *Order) error
}
