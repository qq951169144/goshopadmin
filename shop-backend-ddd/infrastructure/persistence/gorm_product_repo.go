package persistence

import (
	"context"
	"fmt"

	"shop-backend-ddd/domain/order"
	"shop-backend-ddd/domain/product"
	"shop-backend-ddd/infrastructure/persistence/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GormProductRepository ProductRepository 的 GORM 实现
// 同时实现 order.ProductQuerier 接口（跨聚合查询）
type GormProductRepository struct {
	db *gorm.DB
}

// NewGormProductRepository 创建 GORM 商品仓库
func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

// FindProductByID 根据ID查找商品
func (r *GormProductRepository) FindProductByID(ctx context.Context, id int) (*product.Product, error) {
	var po models.ProductPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return product.ReconstructProduct(po.ID, po.Name, po.Description, po.Price, po.Status), nil
}

// FindSKUByID 根据ID查找SKU
func (r *GormProductRepository) FindSKUByID(ctx context.Context, id int) (*product.SKU, error) {
	var po models.SkuPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	sku := product.NewSKU(po.ID, po.ProductID, po.SkuCode, po.Price, po.Stock, po.Attributes)
	return &sku, nil
}

// DeductStock 扣减SKU库存
func (r *GormProductRepository) DeductStock(ctx context.Context, skuID int, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&models.SkuPO{}).
		Where("id = ? AND stock >= ?", skuID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	if result.RowsAffected == 0 {
		return fmt.Errorf("SKU %d 库存不足", skuID)
	}
	return nil
}

// RestoreStock 恢复SKU库存
func (r *GormProductRepository) RestoreStock(ctx context.Context, skuID int, quantity int) error {
	return r.db.WithContext(ctx).
		Model(&models.SkuPO{}).
		Where("id = ?", skuID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}

// ============================================================
// 实现 order.ProductQuerier 接口（跨聚合查询适配）
// ============================================================

// FindProductAndSKU 查询商品和SKU信息（实现 order.ProductQuerier 接口）
func (r *GormProductRepository) FindProductAndSKU(ctx context.Context, productID, skuID int) (string, string, decimal.Decimal, int, error) {
	var productPO models.ProductPO
	if err := r.db.WithContext(ctx).First(&productPO, productID).Error; err != nil {
		return "", "", decimal.Zero, 0, fmt.Errorf("商品不存在: %w", err)
	}

	var skuPO models.SkuPO
	if err := r.db.WithContext(ctx).First(&skuPO, skuID).Error; err != nil {
		return "", "", decimal.Zero, 0, fmt.Errorf("SKU不存在: %w", err)
	}

	return productPO.Name, skuPO.Attributes, skuPO.Price, skuPO.Stock, nil
}

// DeductStockTx 为订单扣减库存（在事务中执行，实现 order.ProductQuerier 接口）
func (r *GormProductRepository) DeductStockTx(ctx context.Context, tx order.Transaction, skuID int, quantity int) error {
	db := r.db
	if gormTx, ok := tx.(*GormTransaction); ok {
		db = gormTx.db
	}

	result := db.WithContext(ctx).
		Model(&models.SkuPO{}).
		Where("id = ? AND stock >= ?", skuID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	if result.RowsAffected == 0 {
		return fmt.Errorf("SKU %d 库存不足", skuID)
	}
	return nil
}

// RestoreStockTx 为订单恢复库存（在事务中执行，实现 order.ProductQuerier 接口）
func (r *GormProductRepository) RestoreStockTx(ctx context.Context, tx order.Transaction, skuID int, quantity int) error {
	db := r.db
	if gormTx, ok := tx.(*GormTransaction); ok {
		db = gormTx.db
	}

	return db.WithContext(ctx).
		Model(&models.SkuPO{}).
		Where("id = ?", skuID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}
