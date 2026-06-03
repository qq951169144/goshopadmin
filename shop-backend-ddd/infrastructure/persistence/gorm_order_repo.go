package persistence

import (
	"context"
	"fmt"

	"shop-backend-ddd/domain/order"
	"shop-backend-ddd/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// GormOrderRepository OrderRepository 的 GORM 实现
type GormOrderRepository struct {
	db *gorm.DB
}

// NewGormOrderRepository 创建 GORM 订单仓库
func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

// FindByOrderNoAndCustomer 根据订单号和客户ID查找订单
func (r *GormOrderRepository) FindByOrderNoAndCustomer(ctx context.Context, orderNo string, customerID int) (*order.Order, error) {
	var po models.OrderPO
	result := r.db.WithContext(ctx).
		Where("order_no = ? AND customer_id = ?", orderNo, customerID).
		First(&po)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, order.ErrOrderNotFound
		}
		return nil, result.Error
	}

	// 查询订单项
	var itemPOs []models.OrderItemPO
	r.db.WithContext(ctx).Where("order_id = ?", po.ID).Find(&itemPOs)

	// PO → Domain 转换
	return r.toDomain(&po, itemPOs), nil
}

// Save 保存订单
func (r *GormOrderRepository) Save(ctx context.Context, tx order.Transaction, orderEntity *order.Order) error {
	db := r.getDB(tx)

	// Domain → PO 转换
	orderPO := r.toPO(orderEntity)

	// 保存订单
	if err := db.Create(&orderPO).Error; err != nil {
		return fmt.Errorf("创建订单失败: %w", err)
	}

	// 设置领域实体的ID
	orderEntity.SetID(orderPO.ID)

	// 保存订单项
	for _, item := range orderEntity.Items() {
		itemPO := models.OrderItemPO{
			OrderID:     orderPO.ID,
			ProductID:   item.ProductID(),
			SkuID:       item.SkuID(),
			ProductName: item.ProductName(),
			SkuAttrs:    item.SkuAttrs(),
			Price:       item.Price(),
			Quantity:    item.Quantity(),
			TotalAmount: item.TotalAmount(),
		}
		if err := db.Create(&itemPO).Error; err != nil {
			return fmt.Errorf("创建订单项失败: %w", err)
		}
	}

	return nil
}

// toDomain PO → 领域实体
func (r *GormOrderRepository) toDomain(po *models.OrderPO, itemPOs []models.OrderItemPO) *order.Order {
	items := make([]order.OrderItem, 0, len(itemPOs))
	for _, itemPO := range itemPOs {
		item, _ := order.NewOrderItem(
			itemPO.ProductID, itemPO.SkuID,
			itemPO.ProductName, itemPO.SkuAttrs,
			itemPO.Price, itemPO.Quantity,
		)
		items = append(items, item)
	}

	return order.ReconstructFromDB(
		po.ID, po.OrderNo, po.CustomerID, po.MerchantID, po.AddressID,
		po.TotalAmount, order.OrderStatus(po.Status), items,
		po.Remark, po.CreatedAt, po.CancelledAt,
	)
}

// toPO 领域实体 → PO
func (r *GormOrderRepository) toPO(o *order.Order) models.OrderPO {
	return models.OrderPO{
		OrderNo:     o.OrderNo(),
		CustomerID:  o.CustomerID(),
		MerchantID:  o.MerchantID(),
		TotalAmount: o.TotalAmount(),
		Status:      string(o.Status()),
		AddressID:   o.AddressID(),
		Remark:      o.Remark(),
	}
}

// getDB 获取数据库连接（支持事务）
func (r *GormOrderRepository) getDB(tx order.Transaction) *gorm.DB {
	if gormTx, ok := tx.(*GormTransaction); ok {
		return gormTx.db
	}
	return r.db
}
