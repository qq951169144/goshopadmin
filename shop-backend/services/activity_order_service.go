package services

import (
	"errors"
	"fmt"
	"math/rand"
	"shop-backend/constants"
	"shop-backend/models"
	"shop-backend/utils"
	"time"

	"gorm.io/gorm"
)

// ActivityOrderService 活动订单服务
type ActivityOrderService struct {
	DB              *gorm.DB
	activityService *ActivityService
}

// NewActivityOrderService 创建活动订单服务实例
func NewActivityOrderService(db *gorm.DB) *ActivityOrderService {
	return &ActivityOrderService{
		DB:              db,
		activityService: NewActivityService(db),
	}
}

// ActivityOrderItem 活动订单商品项
type ActivityOrderItem struct {
	ProductID int
	SkuID     int
	Quantity  int
}

// CreateActivityOrder 创建活动订单
func (s *ActivityOrderService) CreateActivityOrder(customerID int, activityID int, items []ActivityOrderItem) (*models.Order, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var activity models.Activity
	result := tx.First(&activity, activityID)
	if result.Error != nil {
		tx.Rollback()
		return nil, errors.New("活动不存在")
	}

	now := time.Now()
	isActivityActive := activity.Status == constants.ActivityStatusActive && activity.StartTime.Before(now) && activity.EndTime.After(now)
	if !isActivityActive {
		tx.Rollback()
		return nil, errors.New("活动已结束或未开始")
	}

	var totalAmount float64
	orderItems := make([]models.OrderItem, 0, len(items))

	for _, item := range items {
		err := s.activityService.CheckActivityStock(activityID, item.ProductID, item.SkuID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		var sku models.ProductSku
		result := tx.First(&sku, item.SkuID)
		if result.Error != nil {
			tx.Rollback()
			return nil, errors.New("商品SKU不存在")
		}

		itemAmount := sku.Price * float64(item.Quantity)
		totalAmount += itemAmount

		orderItem := models.OrderItem{
			ProductID:   item.ProductID,
			SkuID:       item.SkuID,
			Quantity:    item.Quantity,
			Price:       sku.Price,
			TotalAmount: itemAmount,
		}
		orderItems = append(orderItems, orderItem)

		err = s.activityService.ReduceActivityStock(activityID, item.ProductID, item.SkuID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	orderNo := s.generateOrderNo()

	order := &models.Order{
		CustomerID:     customerID,
		OrderNo:        orderNo,
		ActivityID:     activityID,
		TotalAmount:    totalAmount,
		Status:         constants.OrderStatusPending,
		PaymentStatus:  constants.PaymentStatusPending,
		ShippingStatus: constants.ShippingStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return order, nil
}

// GetActivityOrders 获取用户活动订单列表
func (s *ActivityOrderService) GetActivityOrders(customerID int, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	s.DB.Model(&models.Order{}).Where("customer_id = ? AND activity_id > 0", customerID).Count(&total)

	offset := (page - 1) * pageSize
	result := s.DB.Where("customer_id = ? AND activity_id > 0", customerID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders)
	if result.Error != nil {
		utils.Error("获取活动订单失败: %v", result.Error)
		return nil, 0, result.Error
	}

	for i := range orders {
		var items []models.OrderItem
		s.DB.Where("order_id = ?", orders[i].ID).Find(&items)
		orders[i].Items = items
	}

	return orders, total, nil
}

// GetActivityOrderByID 根据ID获取活动订单详情
func (s *ActivityOrderService) GetActivityOrderByID(orderID int, customerID int) (*models.Order, error) {
	var order models.Order

	result := s.DB.Where("id = ? AND customer_id = ? AND activity_id > 0", orderID, customerID).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		utils.Error("获取订单详情失败: %v", result.Error)
		return nil, result.Error
	}

	var items []models.OrderItem
	s.DB.Where("order_id = ?", order.ID).Find(&items)
	order.Items = items

	return &order, nil
}

// generateOrderNo 生成订单号
func (s *ActivityOrderService) generateOrderNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(10000)
	return fmt.Sprintf("ACT%s%04d", timestamp, random)
}

// CancelActivityOrder 取消活动订单
func (s *ActivityOrderService) CancelActivityOrder(orderID int, customerID int) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order models.Order
	result := tx.Where("id = ? AND customer_id = ? AND activity_id > 0 AND status = ?", orderID, customerID, constants.OrderStatusPending).First(&order)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在或状态不正确")
		}
		return result.Error
	}

	var items []models.OrderItem
	result = tx.Where("order_id = ?", orderID).Find(&items)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, item := range items {
		var sku models.ProductSku
		result := tx.First(&sku, item.SkuID)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		if err := tx.Model(&sku).Update("stock", sku.Stock+item.Quantity).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":         constants.OrderStatusCancelled,
		"payment_status": constants.PaymentStatusFailed,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
