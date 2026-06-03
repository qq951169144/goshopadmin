package persistence

import (
	"context"
	"fmt"

	"shop-backend-ddd/domain/customer"
	"shop-backend-ddd/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// GormCustomerRepository CustomerRepository 的 GORM 实现
// 同时实现 order.CustomerQuerier 接口
type GormCustomerRepository struct {
	db *gorm.DB
}

// NewGormCustomerRepository 创建 GORM 客户仓库
func NewGormCustomerRepository(db *gorm.DB) *GormCustomerRepository {
	return &GormCustomerRepository{db: db}
}

// FindByID 根据ID查找客户
func (r *GormCustomerRepository) FindByID(ctx context.Context, id int) (*customer.Customer, error) {
	var po models.CustomerPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return customer.NewCustomer(po.ID, po.Username), nil
}

// FindAddressByID 根据ID查找地址
func (r *GormCustomerRepository) FindAddressByID(ctx context.Context, id int) (*customer.Address, error) {
	var po models.AddressPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	addr := customer.NewAddress(po.ID, po.CustomerID, po.Name, po.Phone, po.Province, po.City, po.District, po.DetailAddress)
	return &addr, nil
}

// VerifyAddress 验证地址是否属于该客户，返回商户ID（实现 order.CustomerQuerier 接口）
func (r *GormCustomerRepository) VerifyAddress(ctx context.Context, customerID, addressID int) (int, error) {
	var addr models.AddressPO
	if err := r.db.WithContext(ctx).Where("id = ? AND customer_id = ?", addressID, customerID).First(&addr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("地址不存在或不属于该客户")
		}
		return 0, err
	}
	// 简化：返回默认商户ID
	return 1, nil
}
