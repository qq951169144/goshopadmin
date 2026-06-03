package persistence

import (
	"context"

	"customer-service/domain/customer"
	"customer-service/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// GormCustomerRepository 基于 GORM 的客户仓库实现
type GormCustomerRepository struct {
	db *gorm.DB
}

// NewGormCustomerRepository 创建 GORM 客户仓库
func NewGormCustomerRepository(db *gorm.DB) *GormCustomerRepository {
	return &GormCustomerRepository{db: db}
}

// FindByID 根据ID查询客户（包含地址列表）
func (r *GormCustomerRepository) FindByID(ctx context.Context, id int) (*customer.Customer, error) {
	var po models.CustomerPO
	err := r.db.Preload("Addresses").Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&po), nil
}

// toDomain 将 PO 转换为领域实体
func (r *GormCustomerRepository) toDomain(po *models.CustomerPO) *customer.Customer {
	addresses := make([]customer.Address, 0, len(po.Addresses))
	for _, addrPO := range po.Addresses {
		addresses = append(addresses, customer.NewAddress(
			addrPO.ID, addrPO.MerchantID, addrPO.Detail,
		))
	}
	return customer.NewCustomer(po.ID, po.Username, addresses)
}
