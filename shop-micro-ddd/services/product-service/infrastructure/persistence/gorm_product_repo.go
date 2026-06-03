package persistence

import (
	"context"

	"product-service/domain/product"
	"product-service/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// GormProductRepository 基于 GORM 的商品仓库实现
type GormProductRepository struct {
	db *gorm.DB
}

// NewGormProductRepository 创建 GORM 商品仓库
func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

// FindByID 根据ID查询商品（包含SKU列表）
func (r *GormProductRepository) FindByID(ctx context.Context, id int) (*product.Product, error) {
	var po models.ProductPO
	err := r.db.Preload("SKUs").Where("id = ?", id).First(&po).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&po), nil
}

// FindSKUByID 根据SKU ID查询SKU
func (r *GormProductRepository) FindSKUByID(ctx context.Context, skuID int) (*product.SKU, error) {
	var po models.SKUPo
	err := r.db.Where("id = ?", skuID).First(&po).Error
	if err != nil {
		return nil, err
	}

	sku := product.NewSKU(po.ID, po.Attrs, po.Price, po.Stock, po.MerchantID)
	return &sku, nil
}

// SaveSKU 保存SKU（更新库存等）
func (r *GormProductRepository) SaveSKU(ctx context.Context, sku *product.SKU) error {
	return r.db.Model(&models.SKUPo{}).Where("id = ?", sku.ID()).
		Updates(map[string]interface{}{
			"stock": sku.Stock(),
		}).Error
}

// toDomain 将 PO 转换为领域实体
func (r *GormProductRepository) toDomain(po *models.ProductPO) *product.Product {
	skus := make([]product.SKU, 0, len(po.SKUs))
	for _, skuPO := range po.SKUs {
		skus = append(skus, product.NewSKU(
			skuPO.ID, skuPO.Attrs, skuPO.Price, skuPO.Stock, skuPO.MerchantID,
		))
	}
	return product.NewProduct(po.ID, po.Name, product.ProductStatus(po.Status), skus)
}
