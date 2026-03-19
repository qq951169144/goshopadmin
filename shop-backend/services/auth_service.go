package services

import (
	"errors"
	"fmt"
	"time"

	"shop-backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db             *gorm.DB
	captchaService *CaptchaService
	jwtSecret      string
	jwtExpireHour  int
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, captchaService *CaptchaService, jwtSecret string, jwtExpireHour int) *AuthService {
	return &AuthService{
		db:             db,
		captchaService: captchaService,
		jwtSecret:      jwtSecret,
		jwtExpireHour:  jwtExpireHour,
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username  string
	Password  string
	CaptchaID string
	Captcha   string
}

// Register 用户注册
func (s *AuthService) Register(req RegisterRequest) (string, *models.Customer, error) {
	// 验证验证码
	if !s.captchaService.VerifyCaptcha(req.CaptchaID, req.Captcha) {
		return "", nil, errors.New("验证码错误")
	}

	// 检查用户名是否已存在
	var existingUser models.Customer
	result := s.db.Where("username = ?", req.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		return "", nil, fmt.Errorf("用户名已存在: %s", req.Username)
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := models.Customer{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return "", nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("生成token失败: %w", err)
	}

	return token, &user, nil
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username  string
	Password  string
	CaptchaID string
	Captcha   string
}

// Login 用户登录
func (s *AuthService) Login(req LoginRequest) (string, *models.Customer, error) {
	// 验证验证码
	if !s.captchaService.VerifyCaptcha(req.CaptchaID, req.Captcha) {
		return "", nil, errors.New("验证码错误")
	}

	// 验证用户名和密码
	var user models.Customer
	result := s.db.Where("username = ?", req.Username).First(&user)
	if result.RowsAffected == 0 {
		return "", nil, errors.New("用户名或密码错误")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 生成token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, errors.New("生成token失败")
	}

	return token, &user, nil
}

// generateToken 生成JWT token
func (s *AuthService) generateToken(customerID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"customer_id": customerID,
		"exp":         time.Now().Add(time.Hour * time.Duration(s.jwtExpireHour)).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
