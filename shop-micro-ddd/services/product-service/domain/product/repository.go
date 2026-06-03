package product

import "context"

// ProductRepository 商品仓库接口
type ProductRepository interface {
	// FindByID 根据ID查询商品（包含SKU列表）
	FindByID(ctx context.Context, id int) (*Product, error)
	// FindSKUByID 根据SKU ID查询SKU
	FindSKUByID(ctx context.Context, skuID int) (*SKU, error)
	// SaveSKU 保存SKU（更新库存等）
	SaveSKU(ctx context.Context, sku *SKU) error
}
