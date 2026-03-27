package controllers

import (
	"shop-backend/errors"
	"shop-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// CaptchaController 验证码控制器
type CaptchaController struct {
	BaseController
	captchaService *services.CaptchaService
}

// NewCaptchaController 创建验证码控制器实例
func NewCaptchaController(redisClient *redis.Client) *CaptchaController {
	return &CaptchaController{
		captchaService: services.NewCaptchaService(redisClient),
	}
}

// GenerateCaptcha 生成验证码
func (c *CaptchaController) GenerateCaptcha(ctx *gin.Context) {
	captcha, err := c.captchaService.GenerateCaptcha()
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "image/png")
	ctx.Header("X-Captcha-ID", captcha.ID)

	// 输出图片
	ctx.Writer.Write(captcha.Image)
}

// VerifyCaptchaRequest 验证验证码请求结构
type VerifyCaptchaRequest struct {
	CaptchaID string `json:"captcha_id" binding:"required"`
	Value     string `json:"value" binding:"required"`
}

// VerifyCaptcha 验证验证码
func (c *CaptchaController) VerifyCaptcha(ctx *gin.Context) {
	var req VerifyCaptchaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	valid := c.captchaService.VerifyCaptcha(req.CaptchaID, req.Value)

	c.ResponseSuccess(ctx, gin.H{"valid": valid})
}
