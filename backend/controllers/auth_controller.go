package controllers

import (
	"encoding/base64"
	"goshopadmin/errors"
	"goshopadmin/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthController 认证控制器
type AuthController struct {
	BaseController
	authService    *services.AuthService
	captchaService *services.CaptchaService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(db *gorm.DB, jwtSecret string, jwtExpireHour int) *AuthController {
	return &AuthController{
		authService:    services.NewAuthService(db, jwtSecret, jwtExpireHour),
		captchaService: services.NewCaptchaService(),
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CaptchaID  string `json:"captcha_id" binding:"required"`
	CaptchaAns int    `json:"captcha_ans" binding:"required"`
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 验证验证码
	// 验证验证码
	// 暂时注释掉验证码验证，先调试登录功能
	/*
		if !c.captchaService.VerifyCaptcha(req.CaptchaID, req.CaptchaAns) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "验证码错误"})
			return
		}
	*/

	// 登录验证
	token, user, err := c.authService.Login(req.Username, req.Password)
	if err != nil {
		c.ResponseError(ctx, errors.CodeLoginFailed, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"role_id":   user.RoleID,
			"role_name": user.Role.Name,
		},
	})
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *gin.Context) {
	// 由于使用JWT，登出只需客户端删除token即可
	c.ResponseSuccess(ctx, nil)
}

// RefreshToken 刷新token
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	token, err := c.authService.RefreshToken(userID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"token": token})
}

// GetCurrentUser 获取当前用户信息
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return
	}

	user, err := c.authService.GetUserByID(userID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"role_id":     user.RoleID,
		"role_name":   user.Role.Name,
		"permissions": user.Role.Permissions,
	})
}

// GetCaptcha 获取验证码图片
func (c *AuthController) GetCaptcha(ctx *gin.Context) {
	captcha, err := c.captchaService.GenerateCaptcha()
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	// 将图片字节转换为base64字符串
	imageBase64 := base64.StdEncoding.EncodeToString(captcha.Image)

	c.ResponseSuccess(ctx, gin.H{
		"id":    captcha.ID,
		"image": imageBase64,
		"ans":   captcha.Answer,
	})
}

// VerifyCaptchaRequest 验证码验证请求结构
type VerifyCaptchaRequest struct {
	CaptchaID  string `json:"captcha_id" binding:"required"`
	CaptchaAns int    `json:"captcha_ans" binding:"required"`
}

// VerifyCaptcha 验证验证码
func (c *AuthController) VerifyCaptcha(ctx *gin.Context) {
	var req VerifyCaptchaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	isValid := c.captchaService.VerifyCaptcha(req.CaptchaID, req.CaptchaAns)
	if !isValid {
		c.ResponseError(ctx, errors.CodeParamInvalid, nil)
		return
	}

	c.ResponseSuccess(ctx, nil)
}
