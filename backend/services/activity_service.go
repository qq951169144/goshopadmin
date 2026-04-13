package services

import (
	"errors"
	"goshopadmin/constants"
	"goshopadmin/models"
	"time"

	"gorm.io/gorm"
)

// ActivityService 活动服务
type ActivityService struct {
	db *gorm.DB
}

// NewActivityService 创建活动服务实例
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

// CheckSkuBoundToOtherActivity 检查SKU是否已绑定其他活动
func (s *ActivityService) CheckSkuBoundToOtherActivity(skuID int, activityID int) error {
	if skuID <= 0 {
		return nil
	}
	var count int64
	err := s.db.Model(&models.ActivityProduct{}).
		Where("sku_id = ? AND activity_id != ? AND status = 'active'", skuID, activityID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该SKU商品已绑定其他活动，请先解绑")
	}
	return nil
}

// CreateActivity 创建活动
func (s *ActivityService) CreateActivity(activity *models.Activity, products []models.ActivityProduct, redeemSetting *models.ActivityRedeemSetting) error {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建活动
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 关联活动商品
	for i := range products {
		products[i].ActivityID = activity.ID

		if err := s.CheckSkuBoundToOtherActivity(products[i].SkuID, activity.ID); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Create(&products[i]).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 更新商品的 is_activity 字段为 1
		if products[i].ProductID > 0 {
			err := tx.Model(&models.Product{}).
				Where("id = ?", products[i].ProductID).
				Update("is_activity", 1).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// 如果是活动专用SKU，更新SKU的活动标识
		if products[i].SkuID > 0 {
			updates := map[string]interface{}{
				"is_activity": 1,
				"activity_id": activity.ID,
			}
			err := tx.Model(&models.ProductSku{}).
				Where("id = ?", products[i].SkuID).
				Updates(updates).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 如果是兑换码活动，创建兑换码配置
	if activity.Type == constants.ActivityTypeRedeemCode && redeemSetting != nil {
		redeemSetting.ActivityID = activity.ID
		redeemSetting.MerchantID = activity.MerchantID
		if err := tx.Create(redeemSetting).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetActivityByID 根据ID获取活动
func (s *ActivityService) GetActivityByID(id int, merchantID int) (*models.Activity, error) {
	var activity models.Activity
	query := s.db.Where("merchant_id = ?", merchantID).
		Preload("Products").
		Preload("Products.Product").
		Preload("Products.Sku").
		Preload("RedeemSetting")

	err := query.First(&activity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		return nil, err
	}
	return &activity, nil
}

// GetActivities 获取活动列表
func (s *ActivityService) GetActivities(merchantID int, activityType string, status string, page, size int) ([]models.Activity, int64, error) {
	var activities []models.Activity
	var total int64

	query := s.db.Model(&models.Activity{}).Where("merchant_id = ?", merchantID)

	// 按类型筛选
	if activityType != "" {
		query = query.Where("type = ?", activityType)
	}

	// 按状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	query = query.Preload("Products").Offset(offset).Limit(size).Order("created_at DESC")
	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// UpdateActivity 更新活动
func (s *ActivityService) UpdateActivity(activity *models.Activity, products []models.ActivityProduct, redeemSetting *models.ActivityRedeemSetting, merchantID int) error {
	// 检查活动是否属于该商户
	var existingActivity models.Activity
	err := s.db.Where("id = ? AND merchant_id = ?", activity.ID, merchantID).
		First(&existingActivity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("活动不存在或不属于该商户")
		}
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存活动基本信息
	if err := tx.Save(activity).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果有商品信息，更新商品关联
	if len(products) > 0 {
		// 获取原有商品ID列表（用于后续重置 is_activity 状态）
		var oldProducts []models.ActivityProduct
		if err := tx.Where("activity_id = ?", activity.ID).Find(&oldProducts).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 删除原有商品关联
		if err := tx.Where("activity_id = ?", activity.ID).Delete(&models.ActivityProduct{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 重置原有商品的 is_activity 状态
		for _, oldProduct := range oldProducts {
			// 重置商品的 is_activity 字段
			if oldProduct.ProductID > 0 {
				if err := tx.Model(&models.Product{}).
					Where("id = ?", oldProduct.ProductID).
					Update("is_activity", 0).Error; err != nil {
					tx.Rollback()
					return err
				}
			}

			// 重置活动专用SKU的状态
			if oldProduct.SkuID > 0 {
				updates := map[string]interface{}{
					"is_activity": 0,
					"activity_id": 0,
				}
				if err := tx.Model(&models.ProductSku{}).
					Where("id = ?", oldProduct.SkuID).
					Updates(updates).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		// 创建新的商品关联
		for i := range products {
			products[i].ActivityID = activity.ID

			if err := s.CheckSkuBoundToOtherActivity(products[i].SkuID, activity.ID); err != nil {
				tx.Rollback()
				return err
			}

			if err := tx.Create(&products[i]).Error; err != nil {
				tx.Rollback()
				return err
			}

			// 更新商品的 is_activity 字段为 1
			if products[i].ProductID > 0 {
				err := tx.Model(&models.Product{}).
					Where("id = ?", products[i].ProductID).
					Update("is_activity", 1).Error
				if err != nil {
					tx.Rollback()
					return err
				}
			}

			// 如果是活动专用SKU，更新SKU的活动标识
			if products[i].SkuID > 0 {
				updates := map[string]interface{}{
					"is_activity": 1,
					"activity_id": activity.ID,
				}
				err := tx.Model(&models.ProductSku{}).
					Where("id = ?", products[i].SkuID).
					Updates(updates).Error
				if err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	// 如果是兑换码活动，更新兑换码配置
	if activity.Type == constants.ActivityTypeRedeemCode && redeemSetting != nil {
		// 删除原有兑换码配置
		if err := tx.Where("activity_id = ?", activity.ID).Delete(&models.ActivityRedeemSetting{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 创建新的兑换码配置
		redeemSetting.ActivityID = activity.ID
		redeemSetting.MerchantID = merchantID
		if err := tx.Create(redeemSetting).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// DeleteActivity 删除活动
func (s *ActivityService) DeleteActivity(id int, merchantID int) error {
	// 检查活动是否属于该商户
	var existingActivity models.Activity
	err := s.db.Where("id = ? AND merchant_id = ?", id, merchantID).
		First(&existingActivity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("活动不存在或不属于该商户")
		}
		return err
	}

	// 获取活动关联的所有商品和SKU
	var activityProducts []models.ActivityProduct
	if err := s.db.Where("activity_id = ?", id).Find(&activityProducts).Error; err != nil {
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除活动商品关联
	if err := tx.Where("activity_id = ?", id).Delete(&models.ActivityProduct{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除兑换码配置
	if err := tx.Where("activity_id = ?", id).Delete(&models.ActivityRedeemSetting{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除兑换码
	if err := tx.Where("activity_id = ?", id).Delete(&models.RedeemCode{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除活动
	if err := tx.Delete(&models.Activity{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 重置关联商品的 is_activity 字段为 0
	for _, product := range activityProducts {
		// 重置商品的 is_activity 字段
		if product.ProductID > 0 {
			if err := tx.Model(&models.Product{}).
				Where("id = ?", product.ProductID).
				Update("is_activity", 0).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		// 重置活动专用SKU的状态
		if product.SkuID > 0 {
			updates := map[string]interface{}{
				"is_activity": 0,
				"activity_id": 0,
			}
			if err := tx.Model(&models.ProductSku{}).
				Where("id = ?", product.SkuID).
				Updates(updates).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// UpdateActivityStatus 更新活动状态
func (s *ActivityService) UpdateActivityStatus(id int, status string, merchantID int) error {
	// 检查活动是否属于该商户
	var existingActivity models.Activity
	err := s.db.Where("id = ? AND merchant_id = ?", id, merchantID).
		First(&existingActivity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("活动不存在或不属于该商户")
		}
		return err
	}

	err = s.db.Model(&models.Activity{}).
		Where("id = ? AND merchant_id = ?", id, merchantID).
		Update("status", status).Error
	if err != nil {
		return err
	}
	return nil
}

// CheckActivityStatus 检查活动状态
func (s *ActivityService) CheckActivityStatus(activity *models.Activity) error {
	// 检查活动是否存在
	if activity == nil {
		return errors.New("活动不存在")
	}

	// 检查活动状态
	if activity.Status != constants.StatusActive {
		return errors.New("活动状态无效")
	}

	// 检查活动时间
	now := time.Now()
	if now.Before(activity.StartTime) {
		return errors.New("活动未开始")
	}
	if now.After(activity.EndTime) {
		return errors.New("活动已过期")
	}

	return nil
}
