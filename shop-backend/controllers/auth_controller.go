package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shop-backend/services"
)

// AuthController 认证控制器
type AuthController struct {
	BaseController
	authService *services.AuthService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
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
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	token, user, err := c.authService.Register(services.RegisterRequest{
		Username:  req.Username,
		Password:  req.Password,
		CaptchaID: req.CaptchaID,
		Captcha:   req.Captcha,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Register success",
		"token":   token,
		"user_id": user.ID,
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
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	token, user, err := c.authService.Login(services.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		CaptchaID: req.CaptchaID,
		Captcha:   req.Captcha,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"token":   token,
		"user_id": user.ID,
	})
}

// Logout 登出
func (c *AuthController) Logout(ctx *gin.Context) {
	// 前端清除token即可
	c.ResponseSuccess(ctx, gin.H{"message": "Logout success"})
}
