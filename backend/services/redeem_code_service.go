package services

import (
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"goshopadmin/constants"
	"goshopadmin/models"

	"gorm.io/gorm"
)

// RedeemCodeService 兑换码服务
type RedeemCodeService struct {
	db *gorm.DB
}

// NewRedeemCodeService 创建兑换码服务实例
func NewRedeemCodeService(db *gorm.DB) *RedeemCodeService {
	return &RedeemCodeService{db: db}
}

// GenerateRedeemCodes 生成兑换码
func (s *RedeemCodeService) GenerateRedeemCodes(activityID, merchantID, createdBy, quantity int, codeType string, codeLength int, excludeChars string) ([]string, error) {
	// 验证活动是否存在且属于该商户
	var activity models.Activity
	err := s.db.Where("id = ? AND merchant_id = ?", activityID, merchantID).
		First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在或不属于该商户")
		}
		return nil, err
	}

	// 验证活动类型
	if activity.Type != constants.ActivityTypeRedeemCode {
		return nil, errors.New("该活动不是兑换码活动")
	}

	// 生成兑换码
	codes, err := s.generateCodes(quantity, codeType, codeLength, excludeChars)
	if err != nil {
		return nil, err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 批量创建兑换码
	var redeemCodes []models.RedeemCode
	for _, code := range codes {
		redeemCodes = append(redeemCodes, models.RedeemCode{
			ActivityID:     activityID,
			MerchantID:     merchantID,
			Code:           code,
			Status:         constants.RedeemCodeStatusUnused,
			ValidStartTime: activity.StartTime,
			ValidEndTime:   activity.EndTime,
			CreatedBy:      createdBy,
		})
	}

	if err := tx.Create(&redeemCodes).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 更新兑换码配置中的总数量
	err = tx.Model(&models.ActivityRedeemSetting{}).
		Where("activity_id = ?", activityID).
		Update("total_quantity", gorm.Expr("total_quantity + ?", quantity)).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return codes, nil
}

// generateCodes 生成兑换码
func (s *RedeemCodeService) generateCodes(quantity int, codeType string, length int, excludeChars string) ([]string, error) {
	var codes []string
	codeSet := make(map[string]bool)

	// 根据类型确定字符集
	var charset string
	switch codeType {
	case constants.RedeemCodeTypeNumeric:
		charset = "23456789" // 排除0,1
	case constants.RedeemCodeTypeAlpha:
		charset = "ABCDEFGHJKLMNPQRSTUVWXYZ" // 排除I,O
	case constants.RedeemCodeTypeAlphanumeric:
		charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除I,O,0,1
	default:
		charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	}

	// 移除排除的字符
	for _, char := range excludeChars {
		charset = strings.ReplaceAll(charset, string(char), "")
	}

	maxAttempts := quantity * 10 // 最大尝试次数
	attempts := 0

	for len(codes) < quantity && attempts < maxAttempts {
		code := s.generateRandomCode(length, charset)
		if !codeSet[code] {
			codeSet[code] = true
			codes = append(codes, code)
		}
		attempts++
	}

	if len(codes) < quantity {
		return nil, errors.New("兑换码生成失败，可能存在重复")
	}

	return codes, nil
}

// generateRandomCode 生成单个随机兑换码
func (s *RedeemCodeService) generateRandomCode(length int, charset string) string {
	result := make([]byte, length)
	charsetLen := len(charset)

	for i := 0; i < length; i++ {
		result[i] = charset[rand.IntN(charsetLen)]
	}

	return string(result)
}

// GetRedeemCodes 获取兑换码列表
func (s *RedeemCodeService) GetRedeemCodes(activityID int, merchantID int, status string, page, size int) ([]models.RedeemCode, int64, map[string]int, error) {
	var redeemCodes []models.RedeemCode
	var total int64

	// 验证活动是否属于该商户
	var activity models.Activity
	err := s.db.Where("id = ? AND merchant_id = ?", activityID, merchantID).
		First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil, errors.New("活动不存在或不属于该商户")
		}
		return nil, 0, nil, err
	}

	query := s.db.Model(&models.RedeemCode{}).Where("activity_id = ?", activityID)

	// 按状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&redeemCodes).Error; err != nil {
		return nil, 0, nil, err
	}

	// 统计各状态数量
	stats := make(map[string]int)
	var statusCounts []struct {
		Status string
		Count  int
	}

	err = s.db.Model(&models.RedeemCode{}).
		Where("activity_id = ?", activityID).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, 0, nil, err
	}

	for _, sc := range statusCounts {
		stats[sc.Status] = sc.Count
	}

	return redeemCodes, total, stats, nil
}

// VerifyRedeemCode 核销兑换码
func (s *RedeemCodeService) VerifyRedeemCode(code string, verifyBy int, remark string) (*models.RedeemCode, *models.RedeemCodeLog, error) {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找兑换码
	var redeemCode models.RedeemCode
	err := tx.Where("code = ?", code).
		Preload("Activity").
		Preload("Activity.Products").
		First(&redeemCode).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("兑换码不存在")
		}
		return nil, nil, err
	}

	// 检查兑换码状态
	if redeemCode.Status == constants.RedeemCodeStatusUsed {
		tx.Rollback()
		return nil, nil, errors.New("兑换码已被使用")
	}

	if redeemCode.Status == constants.RedeemCodeStatusExpired {
		tx.Rollback()
		return nil, nil, errors.New("兑换码已过期")
	}

	if redeemCode.Status == constants.RedeemCodeStatusDisabled {
		tx.Rollback()
		return nil, nil, errors.New("兑换码已禁用")
	}

	// 检查活动状态
	if redeemCode.Activity.Status != constants.StatusActive {
		tx.Rollback()
		return nil, nil, errors.New("活动状态无效")
	}

	// 检查活动时间
	now := time.Now()
	if now.Before(redeemCode.Activity.StartTime) {
		tx.Rollback()
		return nil, nil, errors.New("活动未开始")
	}

	if now.After(redeemCode.Activity.EndTime) {
		tx.Rollback()
		return nil, nil, errors.New("活动已过期")
	}

	// 更新兑换码状态
	err = tx.Model(&redeemCode).Updates(map[string]interface{}{
		"status":  constants.RedeemCodeStatusUsed,
		"used_at": now,
		"used_by": redeemCode.UsedBy,
	}).Error
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// 创建核销记录
	redeemCodeLog := &models.RedeemCodeLog{
		RedeemCodeID: redeemCode.ID,
		ActivityID:   redeemCode.ActivityID,
		MerchantID:   redeemCode.MerchantID,
		CustomerID:   redeemCode.UsedBy,
		Code:         redeemCode.Code,
		VerifyBy:     verifyBy,
		VerifyAt:     now,
		Status:       constants.RedeemCodeLogStatusVerified,
		Remark:       remark,
	}

	if err := tx.Create(redeemCodeLog).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return &redeemCode, redeemCodeLog, nil
}

// UpdateRedeemCodeStatus 更新兑换码状态
func (s *RedeemCodeService) UpdateRedeemCodeStatus(id int, status string, merchantID int) error {
	// 检查兑换码是否属于该商户
	var redeemCode models.RedeemCode
	err := s.db.Where("id = ?", id).First(&redeemCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("兑换码不存在")
		}
		return err
	}

	// 验证活动是否属于该商户
	var activity models.Activity
	err = s.db.Where("id = ? AND merchant_id = ?", redeemCode.ActivityID, merchantID).
		First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("活动不存在或不属于该商户")
		}
		return err
	}

	err = s.db.Model(&models.RedeemCode{}).
		Where("id = ?", id).
		Update("status", status).Error
	if err != nil {
		return err
	}
	return nil
}

// GetRedeemCodeLogs 获取核销记录
func (s *RedeemCodeService) GetRedeemCodeLogs(activityID int, merchantID int, page, size int) ([]models.RedeemCodeLog, int64, error) {
	var logs []models.RedeemCodeLog
	var total int64

	query := s.db.Model(&models.RedeemCodeLog{}).Where("merchant_id = ?", merchantID)

	// 按活动ID筛选
	if activityID > 0 {
		// 验证活动是否属于该商户
		var activity models.Activity
		err := s.db.Where("id = ? AND merchant_id = ?", activityID, merchantID).
			First(&activity).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, errors.New("活动不存在或不属于该商户")
			}
			return nil, 0, err
		}
		query = query.Where("activity_id = ?", activityID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
