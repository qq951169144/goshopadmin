package services

import (
	"errors"
	"goshopadmin/config"
	"goshopadmin/models"
	"goshopadmin/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	DB *gorm.DB
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{DB: db}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	// 记录登录请求
	utils.Info("用户登录请求: username=%s", username)

	// 查找用户
	var user models.User
	result := s.DB.Preload("Role").Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.Info("用户不存在: username=%s", username)
			return "", nil, errors.New("用户名或密码错误")
		}
		utils.Error("查询用户失败: %v", result.Error)
		return "", nil, result.Error
	}

	utils.Info("用户查询成功: id=%d, username=%s, status=%s", user.ID, user.Username, user.Status)

	// 检查用户状态
	if user.Status == "inactive" {
		utils.Info("账号已被禁用: username=%s", username)
		return "", nil, errors.New("账号已被禁用")
	}

	// 验证密码
	utils.Info("验证密码: username=%s", username)
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		utils.Info("密码验证失败: username=%s", username)
		return "", nil, errors.New("用户名或密码错误")
	}

	utils.Info("密码验证成功: username=%s", username)

	// 生成JWT token
	utils.Info("生成JWT token: user_id=%d, username=%s", user.ID, user.Username)
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.RoleID,
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTExpireHour,
	)
	if err != nil {
		utils.Error("生成token失败: %v", err)
		return "", nil, err
	}

	utils.Info("登录成功: username=%s", username)
	return token, &user, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	var user models.User
	result := s.DB.Preload("Role").Preload("Role.Permissions").First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(userID int) (string, error) {
	// 获取用户信息
	user, err := s.GetUserByID(userID)
	if err != nil {
		return "", err
	}

	// 生成新token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.RoleID,
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTExpireHour,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID int, oldPassword, newPassword string) error {
	// 获取用户
	var user models.User
	result := s.DB.First(&user, userID)
	if result.Error != nil {
		return result.Error
	}

	// 验证旧密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()
	result = s.DB.Save(&user)
	return result.Error
}

// GetUsers 获取用户列表
func (s *AuthService) GetUsers() ([]*models.User, error) {
	var users []*models.User
	result := s.DB.Preload("Role").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// CreateUser 创建用户
func (s *AuthService) CreateUser(username, password string, roleID int) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	result := s.DB.Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("用户名已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:  username,
		Password:  string(hashedPassword),
		RoleID:    roleID,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result = s.DB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	// 加载角色信息
	s.DB.Preload("Role").First(user, user.ID)
	return user, nil
}

// UpdateUser 更新用户
func (s *AuthService) UpdateUser(id int, password string, roleID int, status string) (*models.User, error) {
	// 获取用户
	var user models.User
	result := s.DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 更新密码
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	// 更新角色
	if roleID > 0 {
		user.RoleID = roleID
	}

	// 更新状态
	if status != "" {
		user.Status = status
	}

	user.UpdatedAt = time.Now()
	result = s.DB.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	// 加载角色信息
	s.DB.Preload("Role").First(&user, user.ID)
	return &user, nil
}

// DeleteUser 删除用户（修改状态为不可用）
func (s *AuthService) DeleteUser(id int) error {
	result := s.DB.Model(&models.User{}).Where("id = ?", id).Update("status", "inactive")
	return result.Error
}

// GetRoles 获取角色列表
func (s *AuthService) GetRoles() ([]*models.Role, error) {
	var roles []*models.Role
	result := s.DB.Preload("Permissions").Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (s *AuthService) GetRoleByID(id int) (*models.Role, error) {
	var role models.Role
	result := s.DB.Preload("Permissions").First(&role, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

// CreateRole 创建角色
func (s *AuthService) CreateRole(name, description string) (*models.Role, error) {
	// 检查角色名是否已存在
	var existingRole models.Role
	result := s.DB.Where("name = ?", name).First(&existingRole)
	if result.Error == nil {
		return nil, errors.New("角色名已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 创建角色
	role := &models.Role{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result = s.DB.Create(role)
	if result.Error != nil {
		return nil, result.Error
	}

	return role, nil
}

// UpdateRole 更新角色
func (s *AuthService) UpdateRole(id int, name, description string) (*models.Role, error) {
	// 获取角色
	var role models.Role
	result := s.DB.First(&role, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 更新角色信息
	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}

	role.UpdatedAt = time.Now()
	result = s.DB.Save(&role)
	if result.Error != nil {
		return nil, result.Error
	}

	// 加载权限信息
	s.DB.Preload("Permissions").First(&role, role.ID)
	return &role, nil
}

// DeleteRole 删除角色（修改状态为不可用）
func (s *AuthService) DeleteRole(id int) error {
	// 检查是否有用户使用该角色
	var userCount int64
	s.DB.Model(&models.User{}).Where("role_id = ?", id).Count(&userCount)
	if userCount > 0 {
		return errors.New("该角色已被用户使用，无法删除")
	}

	// 修改角色状态为不可用
	result := s.DB.Model(&models.Role{}).Where("id = ?", id).Update("status", "inactive")
	return result.Error
}

// AssignPermissions 为角色分配权限
func (s *AuthService) AssignPermissions(roleID int, permissionIDs []int) error {
	// 开始事务
	tx := s.DB.Begin()
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

// GetPermissions 获取权限列表
func (s *AuthService) GetPermissions() ([]*models.Permission, error) {
	var permissions []*models.Permission
	result := s.DB.Find(&permissions)
	if result.Error != nil {
		return nil, result.Error
	}
	return permissions, nil
}

// GetPermissionByID 根据ID获取权限
func (s *AuthService) GetPermissionByID(id int) (*models.Permission, error) {
	var permission models.Permission
	result := s.DB.First(&permission, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &permission, nil
}

// CreatePermission 创建权限
func (s *AuthService) CreatePermission(name, code, description string) (*models.Permission, error) {
	// 检查权限名是否已存在
	var existingPermission models.Permission
	result := s.DB.Where("name = ?", name).First(&existingPermission)
	if result.Error == nil {
		return nil, errors.New("权限名已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 检查权限代码是否已存在
	result = s.DB.Where("code = ?", code).First(&existingPermission)
	if result.Error == nil {
		return nil, errors.New("权限代码已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 创建权限
	permission := &models.Permission{
		Name:        name,
		Code:        code,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result = s.DB.Create(permission)
	if result.Error != nil {
		return nil, result.Error
	}

	return permission, nil
}

// UpdatePermission 更新权限
func (s *AuthService) UpdatePermission(id int, name, code, description string) (*models.Permission, error) {
	// 获取权限
	var permission models.Permission
	result := s.DB.First(&permission, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 检查权限名是否已存在
	if name != "" && name != permission.Name {
		var existingPermission models.Permission
		result := s.DB.Where("name = ?", name).First(&existingPermission)
		if result.Error == nil && existingPermission.ID != id {
			return nil, errors.New("权限名已存在")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		permission.Name = name
	}

	// 检查权限代码是否已存在
	if code != "" && code != permission.Code {
		var existingPermission models.Permission
		result := s.DB.Where("code = ?", code).First(&existingPermission)
		if result.Error == nil && existingPermission.ID != id {
			return nil, errors.New("权限代码已存在")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		permission.Code = code
	}

	// 更新描述
	if description != "" {
		permission.Description = description
	}

	permission.UpdatedAt = time.Now()
	result = s.DB.Save(&permission)
	if result.Error != nil {
		return nil, result.Error
	}

	return &permission, nil
}

// DeletePermission 删除权限（修改状态为不可用）
func (s *AuthService) DeletePermission(id int) error {
	// 检查是否有角色使用该权限
	// 由于不设外键约束，这里可以直接修改状态

	// 修改权限状态为不可用
	result := s.DB.Model(&models.Permission{}).Where("id = ?", id).Update("status", "inactive")
	return result.Error
}
