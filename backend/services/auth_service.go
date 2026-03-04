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
