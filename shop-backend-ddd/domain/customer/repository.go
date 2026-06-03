package customer

import "context"

// CustomerRepository 客户仓库接口
type CustomerRepository interface {
	// FindByID 根据ID查找客户
	FindByID(ctx context.Context, id int) (*Customer, error)
	// FindAddressByID 根据ID查找地址
	FindAddressByID(ctx context.Context, id int) (*Address, error)
}
