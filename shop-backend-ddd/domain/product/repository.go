package product

import "context"

// ProductRepository 商品仓库接口
type ProductRepository interface {
	// FindProductByID 根据ID查找商品
	FindProductByID(ctx context.Context, id int) (*Product, error)
	// FindSKUByID 根据ID查找SKU
	FindSKUByID(ctx context.Context, id int) (*SKU, error)
	// DeductStock 扣减SKU库存
	DeductStock(ctx context.Context, skuID int, quantity int) error
	// RestoreStock 恢复SKU库存
	RestoreStock(ctx context.Context, skuID int, quantity int) error
}
