package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"shop-backend/config"
	"shop-backend/models"
)

// 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

// 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

// 注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 验证验证码
	ctx := context.Background()
	storedCaptcha, err := config.Redis.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.Captcha {
		ResponseError(c, http.StatusBadRequest, "Invalid captcha")
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	result := config.DB.Where("username = ?", req.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		ResponseError(c, http.StatusBadRequest, "Username already exists")
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// 生成token
	token, err := generateToken(user.ID)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Register success",
		"token":   token,
		"user_id": user.ID,
	})
}

// 登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 验证验证码
	ctx := context.Background()
	storedCaptcha, err := config.Redis.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.Captcha {
		ResponseError(c, http.StatusBadRequest, "Invalid captcha")
		return
	}

	// 验证用户名和密码
	var user models.User
	result := config.DB.Where("username = ?", req.Username).First(&user)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		ResponseError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// 生成token
	token, err := generateToken(user.ID)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	ResponseSuccess(c, gin.H{
		"token":   token,
		"user_id": user.ID,
	})
}

// 登出
func Logout(c *gin.Context) {
	// 前端清除token即可
	ResponseSuccess(c, gin.H{"message": "Logout success"})
}

// 生成JWT token
func generateToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * time.Duration(config.AppConfig.JWTExpireHour)).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}