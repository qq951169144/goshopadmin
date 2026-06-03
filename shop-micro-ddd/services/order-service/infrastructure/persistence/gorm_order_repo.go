package persistence

import (
	"context"

	"order-service/domain/order"
	"order-service/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// GormOrderRepository 基于 GORM 的订单仓库实现
type GormOrderRepository struct {
	db *gorm.DB
}

// NewGormOrderRepository 创建 GORM 订单仓库
func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

// FindByOrderNoAndCustomer 根据订单号和客户ID查询订单
func (r *GormOrderRepository) FindByOrderNoAndCustomer(ctx context.Context, orderNo string, customerID int) (*order.Order, error) {
	var po models.OrderPO
	err := r.db.Preload("Items").Where("order_no = ? AND customer_id = ?", orderNo, customerID).First(&po).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&po), nil
}

// Save 保存订单
func (r *GormOrderRepository) Save(ctx context.Context, tx order.Transaction, orderEntity *order.Order) error {
	db := r.getDB(tx)

	// 转换为 PO
	orderPO := r.toPO(orderEntity)

	// 保存订单主表
	if err := db.Create(orderPO).Error; err != nil {
		return err
	}

	// 回填ID
	orderEntity.SetID(orderPO.ID)
	return nil
}

// toDomain 将 PO 转换为领域实体
func (r *GormOrderRepository) toDomain(po *models.OrderPO) *order.Order {
	items := make([]order.OrderItem, 0, len(po.Items))
	for _, itemPO := range po.Items {
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

// toPO 将领域实体转换为 PO
func (r *GormOrderRepository) toPO(o *order.Order) *models.OrderPO {
	itemPOs := make([]models.OrderItemPO, 0, len(o.Items()))
	for _, item := range o.Items() {
		itemPOs = append(itemPOs, models.OrderItemPO{
			ProductID:   item.ProductID(),
			SkuID:       item.SkuID(),
			ProductName: item.ProductName(),
			SkuAttrs:    item.SkuAttrs(),
			Price:       item.Price(),
			Quantity:    item.Quantity(),
			TotalAmount: item.TotalAmount(),
		})
	}

	return &models.OrderPO{
		OrderNo:     o.OrderNo(),
		CustomerID:  o.CustomerID(),
		MerchantID:  o.MerchantID(),
		TotalAmount: o.TotalAmount(),
		Status:      string(o.Status()),
		AddressID:   o.AddressID(),
		Remark:      o.Remark(),
		CancelledAt: o.CancelledAt(),
		Items:       itemPOs,
	}
}

// getDB 获取数据库连接，优先使用事务连接
func (r *GormOrderRepository) getDB(tx order.Transaction) *gorm.DB {
	if gormTx, ok := tx.(*GormTransaction); ok {
		return gormTx.db
	}
	return r.db
}
