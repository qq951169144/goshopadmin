package services

import (
	"encoding/json"
	"errors"
	"goshopadmin/constants"
	"goshopadmin/models"
	"goshopadmin/utils"
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
func (s *MerchantService) CreateMerchant(name, contactName, contactPhone, email, address, businessLicense, taxNumber string, createdBy int) (*models.Merchant, error) {
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
		Name:            name,
		ContactName:     contactName,
		ContactPhone:    contactPhone,
		Email:           email,
		Address:         address,
		BusinessLicense: businessLicense,
		TaxNumber:       taxNumber,
		AuditStatus:     constants.AuditStatusPending,
		Status:          constants.StatusInactive,
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
		utils.Info("创建商户失败: %v", err)
		return nil, err
	}

	// 创建审核记录
	newData := map[string]any{
		"name":             name,
		"contact_name":     contactName,
		"contact_phone":    contactPhone,
		"email":            email,
		"address":          address,
		"business_license": businessLicense,
		"tax_number":       taxNumber,
	}
	newDataJSON, err := json.Marshal(newData)
	if err != nil {
		utils.Error("JSON序列化失败: %v", err)
		tx.Rollback()
		return nil, err
	}

	audit := &models.MerchantAudit{
		MerchantID: merchant.ID,
		AuditType:  "registration",
		OldData:    "{}",
		NewData:    string(newDataJSON),
		Status:     constants.AuditStatusPending,
		CreatedBy:  createdBy,
	}

	if err := tx.Create(audit).Error; err != nil {
		utils.Info("创建审核记录失败%v", err)
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
func (s *MerchantService) UpdateMerchant(id int, name, contactName, contactPhone, email, address, businessLicense, taxNumber, status string, updatedBy int) (*models.Merchant, error) {
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
	oldData := map[string]interface{}{
		"name":             merchant.Name,
		"contact_name":     merchant.ContactName,
		"contact_phone":    merchant.ContactPhone,
		"email":            merchant.Email,
		"address":          merchant.Address,
		"business_license": merchant.BusinessLicense,
		"tax_number":       merchant.TaxNumber,
		"status":           merchant.Status,
	}
	oldDataJSON, err := json.Marshal(oldData)
	if err != nil {
		utils.Error("JSON序列化失败: %v", err)
		tx.Rollback()
		return nil, err
	}

	// 更新商户信息
	if name != "" {
		merchant.Name = name
	}
	if contactName != "" {
		merchant.ContactName = contactName
	}
	if contactPhone != "" {
		merchant.ContactPhone = contactPhone
	}
	if email != "" {
		merchant.Email = email
	}
	if address != "" {
		merchant.Address = address
	}
	if businessLicense != "" {
		merchant.BusinessLicense = businessLicense
	}
	if taxNumber != "" {
		merchant.TaxNumber = taxNumber
	}
	if status != "" {
		merchant.Status = status
	}

	// 重置审核状态
	merchant.AuditStatus = constants.AuditStatusPending

	if err := tx.Save(&merchant).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建审核记录
	newData := map[string]interface{}{
		"name":             merchant.Name,
		"contact_name":     merchant.ContactName,
		"contact_phone":    merchant.ContactPhone,
		"email":            merchant.Email,
		"address":          merchant.Address,
		"business_license": merchant.BusinessLicense,
		"tax_number":       merchant.TaxNumber,
		"status":           merchant.Status,
	}
	newDataJSON, err := json.Marshal(newData)
	if err != nil {
		utils.Error("JSON序列化失败: %v", err)
		tx.Rollback()
		return nil, err
	}

	audit := &models.MerchantAudit{
		MerchantID: merchant.ID,
		AuditType:  "update",
		OldData:    string(oldDataJSON),
		NewData:    string(newDataJSON),
		Status:     constants.AuditStatusPending,
		CreatedBy:  updatedBy,
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
	if status == constants.AuditStatusApproved {
		merchant.Status = constants.StatusActive
		now := time.Now()
		merchant.ApprovedAt = &now
		merchant.ApprovedBy = &auditedBy
	} else if status == constants.AuditStatusRejected {
		merchant.Status = constants.StatusInactive
	}

	if err := tx.Save(&merchant).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新审核记录
	var audit models.MerchantAudit
	result = tx.Where("merchant_id = ? AND status = ?", id, constants.AuditStatusPending).Order("created_at DESC").First(&audit)
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
	result := s.DB.Where("merchant_id = ? AND status = ?", merchantID, "active").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, status")
	}).Preload("Merchant", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, status")
	}).Find(&merchantUsers)
	if result.Error != nil {
		return nil, result.Error
	}
	return merchantUsers, nil
}

// AddMerchantUser 添加商户用户
func (s *MerchantService) AddMerchantUser(merchantID, userID int, role string) error {
	// 检查是否已存在活跃的关联
	var existingMerchantUser models.MerchantUser
	result := s.DB.Where("merchant_id = ? AND user_id = ? AND status = ?", merchantID, userID, "active").First(&existingMerchantUser)
	if result.Error == nil {
		return errors.New("该用户已关联到商户")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// 检查是否存在已禁用的关联，如果存在则重新激活
	result = s.DB.Where("merchant_id = ? AND user_id = ? AND status = ?", merchantID, userID, "inactive").First(&existingMerchantUser)
	if result.Error == nil {
		// 重新激活
		existingMerchantUser.Status = "active"
		existingMerchantUser.Role = role
		result = s.DB.Save(&existingMerchantUser)
		return result.Error
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// 创建关联
	merchantUser := &models.MerchantUser{
		MerchantID: merchantID,
		UserID:     userID,
		Role:       role,
		Status:     "active",
	}

	result = s.DB.Create(merchantUser)
	return result.Error
}

// RemoveMerchantUser 移除商户用户
func (s *MerchantService) RemoveMerchantUser(merchantID, userID int) error {
	result := s.DB.Model(&models.MerchantUser{}).Where("merchant_id = ? AND user_id = ?", merchantID, userID).Update("status", "inactive")
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

// GetMerchantIDByUserID 根据用户ID获取商户ID
func (s *MerchantService) GetMerchantIDByUserID(userID int) (int, error) {
	var merchantUser models.MerchantUser
	result := s.DB.Where("user_id = ? AND status = ?", userID, "active").First(&merchantUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, errors.New("用户未关联商户")
		}
		return 0, result.Error
	}
	return merchantUser.MerchantID, nil
}
