package customer

import "context"

// CustomerRepository 客户仓库接口
type CustomerRepository interface {
	// FindByID 根据ID查询客户（包含地址列表）
	FindByID(ctx context.Context, id int) (*Customer, error)
}
