package order

import "context"

// Transaction 事务接口（抽象 gorm.DB 的事务）
type Transaction interface {
	Commit() error
	Rollback() error
}

// TransactionManager 事务管理器接口
type TransactionManager interface {
	BeginTx(ctx context.Context) (Transaction, error)
}

// OrderRepository 订单仓库接口
// 由领域层定义，基础设施层实现
// 提供类似集合的语义：查找、保存
type OrderRepository interface {
	// FindByOrderNoAndCustomer 根据订单号和客户ID查找订单
	FindByOrderNoAndCustomer(ctx context.Context, orderNo string, customerID int) (*Order, error)
	// Save 保存订单（新建或更新）
	Save(ctx context.Context, tx Transaction, order *Order) error
}
