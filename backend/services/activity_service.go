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
		if err := tx.Create(&products[i]).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 如果是活动专用SKU，更新SKU的活动标识
		if products[i].SKUID > 0 {
			updates := map[string]interface{}{
				"is_activity": 1,
				"activity_id": activity.ID,
			}
			err := tx.Model(&models.ProductSKU{}).
				Where("id = ?", products[i].SKUID).
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
		Preload("Products.SKU").
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
func (s *ActivityService) UpdateActivity(activity *models.Activity, merchantID int) error {
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

	if err := s.db.Save(activity).Error; err != nil {
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
