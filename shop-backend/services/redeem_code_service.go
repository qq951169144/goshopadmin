package services

import (
	"errors"
	"shop-backend/constants"
	"shop-backend/models"
	"shop-backend/utils"
	"time"

	"gorm.io/gorm"
)

// RedeemCodeService 兑换码服务
type RedeemCodeService struct {
	DB *gorm.DB
}

// NewRedeemCodeService 创建兑换码服务实例
func NewRedeemCodeService(db *gorm.DB) *RedeemCodeService {
	return &RedeemCodeService{DB: db}
}

// VerifyRedeemCode 验证兑换码
func (s *RedeemCodeService) VerifyRedeemCode(code string) (*models.RedeemCode, error) {
	var redeemCode models.RedeemCode

	result := s.DB.Where("code = ?", code).First(&redeemCode)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("兑换码不存在")
		}
		utils.Error("验证兑换码失败: %v", result.Error)
		return nil, result.Error
	}

	if redeemCode.Status != constants.RedeemCodeStatusActive {
		return nil, errors.New("兑换码已使用或已过期")
	}

	if time.Now().After(redeemCode.ExpireTime) {
		s.DB.Model(&redeemCode).Update("status", constants.RedeemCodeStatusExpired)
		return nil, errors.New("兑换码已过期")
	}

	var activity models.Activity
	result = s.DB.First(&activity, redeemCode.ActivityID)
	if result.Error != nil {
		return nil, errors.New("活动不存在")
	}

	now := time.Now()
	isActivityActive := activity.Status == constants.ActivityStatusActive && activity.StartTime.Before(now) && activity.EndTime.After(now)
	if !isActivityActive {
		return nil, errors.New("活动已结束或未开始")
	}

	return &redeemCode, nil
}

// RedeemCode 兑换码兑换
func (s *RedeemCodeService) RedeemCode(code string, customerID int) (*models.RedeemCode, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var redeemCode models.RedeemCode
	result := tx.Where("code = ?", code).First(&redeemCode)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("兑换码不存在")
		}
		return nil, result.Error
	}

	if redeemCode.Status != constants.RedeemCodeStatusActive {
		tx.Rollback()
		return nil, errors.New("兑换码已使用或已过期")
	}

	if time.Now().After(redeemCode.ExpireTime) {
		tx.Rollback()
		s.DB.Model(&redeemCode).Update("status", constants.RedeemCodeStatusExpired)
		return nil, errors.New("兑换码已过期")
	}

	var activity models.Activity
	result = tx.First(&activity, redeemCode.ActivityID)
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

	var count int64
	tx.Model(&models.RedeemCodeLog{}).Where("activity_id = ? AND customer_id = ?", redeemCode.ActivityID, customerID).Count(&count)
	if count > 0 {
		tx.Rollback()
		return nil, errors.New("您已经使用过该活动的兑换码")
	}

	if err := tx.Model(&redeemCode).Updates(map[string]interface{}{
		"status":      constants.RedeemCodeStatusUsed,
		"used_at":     time.Now(),
		"customer_id": customerID,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	redeemLog := &models.RedeemCodeLog{
		ActivityID:   redeemCode.ActivityID,
		RedeemCodeID: redeemCode.ID,
		CustomerID:   customerID,
		Code:         redeemCode.Code,
		Value:        redeemCode.Value,
		RedeemTime:   time.Now(),
		Status:       constants.RedeemLogStatusSuccess,
	}

	if err := tx.Create(redeemLog).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &redeemCode, nil
}

// GetRedeemCodeLogs 获取用户兑换码使用记录
func (s *RedeemCodeService) GetRedeemCodeLogs(customerID int, page, pageSize int) ([]models.RedeemCodeLog, int64, error) {
	var logs []models.RedeemCodeLog
	var total int64

	s.DB.Model(&models.RedeemCodeLog{}).Where("customer_id = ?", customerID).Count(&total)

	offset := (page - 1) * pageSize
	result := s.DB.Where("customer_id = ?", customerID).Order("redeem_time DESC").Offset(offset).Limit(pageSize).Find(&logs)
	if result.Error != nil {
		utils.Error("获取兑换记录失败: %v", result.Error)
		return nil, 0, result.Error
	}

	return logs, total, nil
}

// GetRedeemCodeByActivity 获取活动的兑换码信息
func (s *RedeemCodeService) GetRedeemCodeByActivity(activityID int) (*models.ActivityRedeemSetting, error) {
	var setting models.ActivityRedeemSetting

	result := s.DB.Where("activity_id = ?", activityID).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动兑换设置不存在")
		}
		return nil, result.Error
	}

	return &setting, nil
}
