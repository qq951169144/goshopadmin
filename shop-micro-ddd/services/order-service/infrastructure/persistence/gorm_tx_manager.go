package persistence

import (
	"context"

	"order-service/domain/order"

	"gorm.io/gorm"
)

// GormTransaction 基于 GORM 的事务实现
type GormTransaction struct {
	db *gorm.DB
}

// Commit 提交事务
func (t *GormTransaction) Commit() error {
	return t.db.Commit().Error
}

// Rollback 回滚事务
func (t *GormTransaction) Rollback() error {
	return t.db.Rollback().Error
}

// GormTransactionManager 基于 GORM 的事务管理器
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager 创建 GORM 事务管理器
func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// BeginTx 开启事务
func (m *GormTransactionManager) BeginTx(ctx context.Context) (order.Transaction, error) {
	tx := m.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &GormTransaction{db: tx}, nil
}
