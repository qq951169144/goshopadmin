package services

import (
	"errors"
	"shop-backend/constants"
	"shop-backend/models"
	"shop-backend/utils"
	"time"

	"gorm.io/gorm"
)

// ActivityService 活动服务
type ActivityService struct {
	DB *gorm.DB
}

// NewActivityService 创建活动服务实例
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{DB: db}
}

// GetActiveActivities 获取当前有效的活动列表
func (s *ActivityService) GetActiveActivities() ([]models.Activity, error) {
	var activities []models.Activity
	now := time.Now().Unix()

	result := s.DB.Where("status = ? AND start_time <= ? AND end_time >= ?", constants.ActivityStatusActive, now, now).Find(&activities)
	if result.Error != nil {
		utils.Error("获取活动列表失败: %v", result.Error)
		return nil, result.Error
	}

	return activities, nil
}

// GetActivityByID 根据ID获取活动详情
func (s *ActivityService) GetActivityByID(activityID int) (*models.Activity, error) {
	var activity models.Activity

	result := s.DB.First(&activity, activityID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		utils.Error("获取活动详情失败: %v", result.Error)
		return nil, result.Error
	}

	now := time.Now().Unix()
	isActivityActive := activity.Status == constants.ActivityStatusActive && activity.StartTime <= now && activity.EndTime >= now
	if !isActivityActive {
		return nil, errors.New("活动已结束或未开始")
	}

	var products []models.ActivityProduct
	s.DB.Where("activity_id = ?", activityID).Find(&products)
	activity.Products = products

	return &activity, nil
}

// GetActivityProducts 获取活动商品列表
func (s *ActivityService) GetActivityProducts(activityID int) ([]models.ActivityProduct, error) {
	var products []models.ActivityProduct

	result := s.DB.Where("activity_id = ?", activityID).Preload("Product").Find(&products)
	if result.Error != nil {
		utils.Error("获取活动商品失败: %v", result.Error)
		return nil, result.Error
	}

	return products, nil
}

// GetActivityProductSkus 获取活动商品的SKU列表
func (s *ActivityService) GetActivityProductSkus(activityID int, productID int) ([]models.ProductSku, error) {
	var skus []models.ProductSku

	result := s.DB.Where("activity_id = ? AND product_id = ? AND status = ?", activityID, productID, constants.StatusActive).Find(&skus)
	if result.Error != nil {
		utils.Error("获取活动商品SKU失败: %v", result.Error)
		return nil, result.Error
	}

	return skus, nil
}

// CheckActivityStock 检查活动商品库存
func (s *ActivityService) CheckActivityStock(activityID int, productID int, skuID int, quantity int) error {
	var sku models.ProductSku

	result := s.DB.Where("id = ? AND activity_id = ? AND status = ?", skuID, activityID, constants.StatusActive).First(&sku)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("活动商品SKU不存在")
		}
		return result.Error
	}

	if sku.Stock < quantity {
		return errors.New("活动商品库存不足")
	}

	return nil
}

// ReduceActivityStock 减少活动商品库存
func (s *ActivityService) ReduceActivityStock(activityID int, productID int, skuID int, quantity int) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sku models.ProductSku
	result := tx.Where("id = ? AND activity_id = ? AND status = ? AND stock >= ?", skuID, activityID, constants.StatusActive, quantity).First(&sku)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("活动商品库存不足")
		}
		return result.Error
	}

	if err := tx.Model(&sku).Update("stock", sku.Stock-quantity).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CheckUserActivityLimit 检查用户活动购买限制
func (s *ActivityService) CheckUserActivityLimit(activityID int, productID int, customerID int) error {
	var activityProduct models.ActivityProduct
	result := s.DB.Where("activity_id = ? AND product_id = ?", activityID, productID).First(&activityProduct)
	if result.Error != nil {
		return result.Error
	}

	if activityProduct.Limit <= 0 {
		return nil
	}

	var count int64
	s.DB.Model(&models.OrderItem{}).Joins("JOIN orders ON order_items.order_id = orders.id").Where("orders.customer_id = ? AND order_items.product_id = ? AND orders.activity_id = ? AND orders.status != ?", customerID, productID, activityID, constants.OrderStatusCancelled).Count(&count)

	if int(count) >= activityProduct.Limit {
		return errors.New("超过活动购买限制")
	}

	return nil
}
