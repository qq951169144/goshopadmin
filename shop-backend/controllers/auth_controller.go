package controllers

import (
	"strings"

	"shop-backend/errors"
	"shop-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// AuthController 认证控制器
type AuthController struct {
	BaseController
	authService *services.AuthService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(db *gorm.DB, redisClient *redis.Client, jwtSecret string, jwtExpireHour int) *AuthController {
	return &AuthController{
		authService: services.NewAuthService(db, services.NewCaptchaService(redisClient), jwtSecret, jwtExpireHour),
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

// Register 注册
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	token, _, err := c.authService.Register(services.RegisterRequest{
		Username:  req.Username,
		Password:  req.Password,
		CaptchaID: req.CaptchaID,
		Captcha:   req.Captcha,
	})
	if err != nil {
		// 根据错误信息判断错误类型，返回对应的错误码
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "验证码错误"):
			c.ResponseError(ctx, errors.CodeCaptchaError, err)
		case strings.Contains(errStr, "用户名已存在"):
			c.ResponseError(ctx, errors.CodeUserExists, err)
		default:
			c.ResponseError(ctx, errors.CodeDuplicate, err)
		}
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Register success",
		"token":   token,
	})
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

// Login 登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	token, _, err := c.authService.Login(services.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		CaptchaID: req.CaptchaID,
		Captcha:   req.Captcha,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeLoginFailed, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"token": token,
	})
}

// Logout 登出
func (c *AuthController) Logout(ctx *gin.Context) {
	// 前端清除token即可
	c.ResponseSuccess(ctx, gin.H{"message": "Logout success"})
}
