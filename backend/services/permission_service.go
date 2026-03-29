package services

import (
	"context"
	"fmt"
	"goshopadmin/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// PermissionService 权限服务
type PermissionService struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(db *gorm.DB, redisClient *redis.Client) *PermissionService {
	return &PermissionService{
		db:          db,
		redisClient: redisClient,
	}
}

// GetUserPermissions 获取用户权限
func (s *PermissionService) GetUserPermissions(userID int) ([]models.Permission, error) {
	var user models.User
	result := s.db.Preload("Role").Preload("Role.Permissions").First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return user.Role.Permissions, nil
}

// GetUserPermissionCodes 获取用户权限代码列表
func (s *PermissionService) GetUserPermissionCodes(userID int) ([]string, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	codes := make([]string, len(permissions))
	for i, permission := range permissions {
		codes[i] = permission.Code
	}
	return codes, nil
}

// HasPermission 检查用户是否有指定权限
func (s *PermissionService) HasPermission(userID int, permissionCode string) (bool, error) {
	codes, err := s.GetUserPermissionCodes(userID)
	if err != nil {
		return false, err
	}

	for _, code := range codes {
		if code == permissionCode {
			return true, nil
		}
	}
	return false, nil
}

// GetPermissions 获取权限列表
func (s *PermissionService) GetPermissions() ([]*models.Permission, error) {
	var permissions []*models.Permission
	result := s.db.Find(&permissions)
	if result.Error != nil {
		return nil, result.Error
	}
	return permissions, nil
}

// GetPermissionByID 根据ID获取权限
func (s *PermissionService) GetPermissionByID(id int) (*models.Permission, error) {
	var permission models.Permission
	result := s.db.First(&permission, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &permission, nil
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(name, code, description, status string) (*models.Permission, error) {
	// 检查权限名是否已存在
	var existingPermission models.Permission
	result := s.db.Where("name = ?", name).First(&existingPermission)
	if result.Error == nil {
		return nil, result.Error
	}

	// 检查权限代码是否已存在
	result = s.db.Where("code = ?", code).First(&existingPermission)
	if result.Error == nil {
		return nil, result.Error
	}

	// 创建权限
	permission := &models.Permission{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      status,
	}

	result = s.db.Create(permission)
	if result.Error != nil {
		return nil, result.Error
	}

	return permission, nil
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(id int, name, code, description, status string) (*models.Permission, error) {
	// 获取权限
	var permission models.Permission
	result := s.db.First(&permission, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 检查权限名是否已存在
	if name != "" && name != permission.Name {
		var existingPermission models.Permission
		result := s.db.Where("name = ?", name).First(&existingPermission)
		if result.Error == nil && existingPermission.ID != id {
			return nil, result.Error
		}
		permission.Name = name
	}

	// 检查权限代码是否已存在
	if code != "" && code != permission.Code {
		var existingPermission models.Permission
		result := s.db.Where("code = ?", code).First(&existingPermission)
		if result.Error == nil && existingPermission.ID != id {
			return nil, result.Error
		}
		permission.Code = code
	}

	// 更新描述
	if description != "" {
		permission.Description = description
	}

	// 更新状态
	if status != "" {
		permission.Status = status
	}

	result = s.db.Save(&permission)
	if result.Error != nil {
		return nil, result.Error
	}

	return &permission, nil
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(id int) error {
	result := s.db.Model(&models.Permission{}).Where("id = ?", id).Update("status", "inactive")
	return result.Error
}

// AssignPermissions 为角色分配权限
func (s *PermissionService) AssignPermissions(roleID int, permissionIDs []int) error {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除现有权限
	if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 分配新权限
	for _, permissionID := range permissionIDs {
		if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// ClearUserPermissionCache 清除用户权限缓存
func (s *PermissionService) ClearUserPermissionCache(userID int) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:permissions", userID)

	if s.redisClient != nil {
		s.redisClient.Del(ctx, cacheKey)
	}
}
