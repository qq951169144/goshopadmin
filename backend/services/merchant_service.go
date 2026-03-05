package services

import (
	"errors"
	"goshopadmin/models"
	"time"

	"gorm.io/gorm"
)

// MerchantService 商户服务
type MerchantService struct {
	DB *gorm.DB
}

// NewMerchantService 创建商户服务实例
func NewMerchantService(db *gorm.DB) *MerchantService {
	return &MerchantService{DB: db}
}

// GetMerchants 获取商户列表
func (s *MerchantService) GetMerchants() ([]*models.Merchant, error) {
	var merchants []*models.Merchant
	result := s.DB.Find(&merchants)
	if result.Error != nil {
		return nil, result.Error
	}
	return merchants, nil
}

// GetMerchantByID 根据ID获取商户
func (s *MerchantService) GetMerchantByID(id int) (*models.Merchant, error) {
	var merchant models.Merchant
	result := s.DB.First(&merchant, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &merchant, nil
}

// CreateMerchant 创建商户
func (s *MerchantService) CreateMerchant(name, contactPerson, contactPhone, address string, createdBy int) (*models.Merchant, error) {
	// 检查商户名称是否已存在
	var existingMerchant models.Merchant
	result := s.DB.Where("name = ?", name).First(&existingMerchant)
	if result.Error == nil {
		return nil, errors.New("商户名称已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 创建商户
	merchant := &models.Merchant{
		Name:          name,
		ContactPerson: contactPerson,
		ContactPhone:  contactPhone,
		Address:       address,
		AuditStatus:   "pending",
		Status:        "inactive",
	}

	// 开始事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建商户
	if err := tx.Create(merchant).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建审核记录
	audit := &models.MerchantAudit{
		MerchantID: merchant.ID,
		AuditType:  "registration",
		NewData: &models.JSON{
			"name":            name,
			"contact_person":  contactPerson,
			"contact_phone":   contactPhone,
			"address":         address,
		},
		Status:    "pending",
		CreatedBy: createdBy,
	}

	if err := tx.Create(audit).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return merchant, nil
}

// UpdateMerchant 更新商户信息
func (s *MerchantService) UpdateMerchant(id int, name, contactPerson, contactPhone, address, status string, updatedBy int) (*models.Merchant, error) {
	// 获取商户
	var merchant models.Merchant
	result := s.DB.First(&merchant, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 开始事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存旧数据
	oldData := &models.JSON{
		"name":           merchant.Name,
		"contact_person": merchant.ContactPerson,
		"contact_phone":  merchant.ContactPhone,
		"address":        merchant.Address,
		"status":         merchant.Status,
	}

	// 更新商户信息
	if name != "" {
		merchant.Name = name
	}
	if contactPerson != "" {
		merchant.ContactPerson = contactPerson
	}
	if contactPhone != "" {
		merchant.ContactPhone = contactPhone
	}
	if address != "" {
		merchant.Address = address
	}
	if status != "" {
		merchant.Status = status
	}

	// 重置审核状态
	merchant.AuditStatus = "pending"

	if err := tx.Save(&merchant).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建审核记录
	audit := &models.MerchantAudit{
		MerchantID: merchant.ID,
		AuditType:  "update",
		OldData:    oldData,
		NewData: &models.JSON{
			"name":           merchant.Name,
			"contact_person": merchant.ContactPerson,
			"contact_phone":  merchant.ContactPhone,
			"address":        merchant.Address,
			"status":         merchant.Status,
		},
		Status:    "pending",
		CreatedBy: updatedBy,
	}

	if err := tx.Create(audit).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

// AuditMerchant 审核商户
func (s *MerchantService) AuditMerchant(id int, status, remark string, auditedBy int) error {
	// 获取商户
	var merchant models.Merchant
	result := s.DB.First(&merchant, id)
	if result.Error != nil {
		return result.Error
	}

	// 开始事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新商户状态
	merchant.AuditStatus = status
	if status == "approved" {
		merchant.Status = "active"
		now := time.Now()
		merchant.ApprovedAt = &now
		merchant.ApprovedBy = &auditedBy
	} else if status == "rejected" {
		merchant.Status = "inactive"
	}

	if err := tx.Save(&merchant).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新审核记录
	var audit models.MerchantAudit
	result = tx.Where("merchant_id = ? AND status = ?", id, "pending").Order("created_at DESC").First(&audit)
	if result.Error == nil {
		audit.Status = status
		audit.Remark = remark
		audit.AuditedBy = &auditedBy
		now := time.Now()
		audit.AuditedAt = &now
		if err := tx.Save(&audit).Error; err != nil {
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

// GetMerchantUsers 获取商户用户列表
func (s *MerchantService) GetMerchantUsers(merchantID int) ([]*models.MerchantUser, error) {
	var merchantUsers []*models.MerchantUser
	result := s.DB.Where("merchant_id = ?", merchantID).Preload("User").Find(&merchantUsers)
	if result.Error != nil {
		return nil, result.Error
	}
	return merchantUsers, nil
}

// AddMerchantUser 添加商户用户
func (s *MerchantService) AddMerchantUser(merchantID, userID int, role string) error {
	// 检查是否已存在
	var existingMerchantUser models.MerchantUser
	result := s.DB.Where("merchant_id = ? AND user_id = ?", merchantID, userID).First(&existingMerchantUser)
	if result.Error == nil {
		return errors.New("该用户已关联到商户")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// 创建关联
	merchantUser := &models.MerchantUser{
		MerchantID: merchantID,
		UserID:     userID,
		Role:       role,
	}

	result = s.DB.Create(merchantUser)
	return result.Error
}

// RemoveMerchantUser 移除商户用户
func (s *MerchantService) RemoveMerchantUser(merchantID, userID int) error {
	result := s.DB.Where("merchant_id = ? AND user_id = ?", merchantID, userID).Delete(&models.MerchantUser{})
	return result.Error
}

// DeleteMerchant 禁用商户
func (s *MerchantService) DeleteMerchant(id int) error {
	// 获取商户
	var merchant models.Merchant
	result := s.DB.First(&merchant, id)
	if result.Error != nil {
		return result.Error
	}

	// 更新商户状态为禁用
	merchant.Status = "inactive"

	// 保存更新
	result = s.DB.Save(&merchant)
	return result.Error
}
