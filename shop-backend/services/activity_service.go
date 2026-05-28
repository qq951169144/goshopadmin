package services

import (
	"errors"
	"shop-backend/constants"
	"shop-backend/models"
	"shop-backend/utils"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ActivitySkuResponse struct {
	SkuID       int             `json:"sku_id"`
	SkuCode     string          `json:"sku_code"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	ProductID   int             `json:"product_id"`
	ProductName string          `json:"product_name"`
	ProductDesc string          `json:"description"`
	IsActivity  int             `json:"is_activity"`
	MainImage   string          `json:"main_image"`
}

type ActivitySkuDetailResponse struct {
	ActivityName        string          `json:"activity_name"`
	ActivityID          int             `json:"activity_id"`
	SkuID               int             `json:"sku_id"`
	SkuCode             string          `json:"sku_code"`
	Price               decimal.Decimal `json:"price"`
	Stock               int             `json:"stock"`
	SkuStatus           string          `json:"sku_status"`
	IsActivity          int             `json:"is_activity"`
	ProductID           int             `json:"product_id"`
	ProductName         string          `json:"product_name"`
	ProductDescription  string          `json:"product_description"`
	MainImage           string          `json:"main_image"`
}

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
	now := time.Now()

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

	now := time.Now()
	isActivityActive := activity.Status == constants.ActivityStatusActive && activity.StartTime.Before(now) && activity.EndTime.After(now)
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
func (s *ActivityService) GetActivityProductSkus(activityID int) ([]ActivitySkuResponse, error) {
	var results []ActivitySkuResponse

	err := s.DB.Table("activity_products").
		Select(`
			activity_products.sku_id,
			activity_products.status,
			product_skus.sku_code,
			product_skus.price,
			product_skus.stock,
			activity_products.product_id,
			products.name as product_name,
			products.description,
			products.is_activity,
			product_images.image_url as main_image
		`).
		Joins("JOIN product_skus ON activity_products.sku_id = product_skus.id").
		Joins("JOIN products ON activity_products.product_id = products.id").
		Joins(`
			LEFT JOIN product_images
				ON product_images.id = (
					SELECT pi.id
					FROM product_images pi
					WHERE pi.product_id = activity_products.product_id
						AND pi.is_main = 1
					ORDER BY pi.sort ASC
					LIMIT 1
				)
		`).
		Where("activity_products.activity_id = ? AND activity_products.status = ?", activityID, constants.ActivityStatusActive).
		Scan(&results).Error

	if err != nil {
		utils.Error("获取活动商品SKU失败: %v", err)
		return nil, err
	}

	return results, nil
}

// GetActivitySkuDetail 获取活动商品SKU详情
func (s *ActivityService) GetActivitySkuDetail(activityID int, skuID int) (*ActivitySkuDetailResponse, error) {
	var result ActivitySkuDetailResponse

	err := s.DB.Table("activity_products").
		Select(`
			activities.name as activity_name,
			activity_products.activity_id,
			activity_products.sku_id,
			product_skus.sku_code,
			product_skus.price,
			product_skus.stock,
			product_skus.status as sku_status,
			product_skus.is_activity,
			product_skus.product_id,
			products.name as product_name,
			products.description as product_description,
			product_images.image_url as main_image
		`).
		Joins("JOIN activities ON activity_products.activity_id = activities.id").
		Joins("JOIN product_skus ON activity_products.sku_id = product_skus.id").
		Joins("JOIN products ON activity_products.product_id = products.id").
		Joins(`
			LEFT JOIN product_images
				ON product_images.id = (
					SELECT pi.id
					FROM product_images pi
					WHERE pi.product_id = activity_products.product_id
						AND pi.is_main = 1
					ORDER BY pi.sort ASC
					LIMIT 1
				)
		`).
		Where("activity_products.activity_id = ? AND activity_products.sku_id = ? AND activity_products.status = ?", activityID, skuID, constants.ActivityStatusActive).
		Scan(&result).Error

	if err != nil {
		utils.Error("获取活动商品SKU详情失败: %v", err)
		return nil, err
	}

	return &result, nil
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
