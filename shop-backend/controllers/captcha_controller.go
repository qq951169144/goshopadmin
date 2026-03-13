package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"shop-backend/errors"
)

// CaptchaController 验证码控制器
type CaptchaController struct {
	BaseController
	redis *redis.Client
}

// NewCaptchaController 创建验证码控制器实例
func NewCaptchaController(redis *redis.Client) *CaptchaController {
	return &CaptchaController{
		redis: redis,
	}
}

const captchaLength = 4

// GenerateCaptcha 生成验证码
func (c *CaptchaController) GenerateCaptcha(ctx *gin.Context) {
	// 生成随机验证码
	captcha := ""
	for i := 0; i < captchaLength; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		captcha += strconv.Itoa(int(n.Int64()))
	}

	// 生成唯一的验证码ID
	captchaID, err := generateRandomID(16)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}

	// 存储验证码到Redis，设置5分钟过期
	rctx := context.Background()
	err = c.redis.Set(rctx, "captcha:"+captchaID, captcha, 5*time.Minute).Err()
	if err != nil {
		c.ResponseError(ctx, errors.CodeCacheError, err)
		return
	}

	// 生成验证码图片
	width := 100
	height := 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景
	bgColor := color.RGBA{240, 240, 240, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 绘制验证码
	textColor := color.RGBA{0, 0, 0, 255}
	for i := range captcha {
		x := 10 + i*20
		y := 25
		// 简单绘制文字
		for j := 0; j < 5; j++ {
			for k := 0; k < 10; k++ {
				n, _ := rand.Int(rand.Reader, big.NewInt(2))
				if n.Int64() == 0 {
					img.Set(x+j, y-k, textColor)
				}
			}
		}
	}

	// 绘制干扰线
	for i := 0; i < 5; i++ {
		x1, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
		y1, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
		x2, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
		y2, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
		for x := min(int(x1.Int64()), int(x2.Int64())); x <= max(int(x1.Int64()), int(x2.Int64())); x++ {
			for y := min(int(y1.Int64()), int(y2.Int64())); y <= max(int(y1.Int64()), int(y2.Int64())); y++ {
				n, _ := rand.Int(rand.Reader, big.NewInt(10))
				if n.Int64() == 0 {
					img.Set(x, y, textColor)
				}
			}
		}
	}

	// 设置响应头
	ctx.Header("Content-Type", "image/png")
	ctx.Header("X-Captcha-ID", captchaID)

	// 输出图片
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		c.ResponseError(ctx, errors.CodeInternalError, err)
		return
	}
	ctx.Writer.Write(buf.Bytes())
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

	// 从Redis获取验证码并验证
	rctx := context.Background()
	storedCaptcha, err := c.redis.Get(rctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.Value {
		c.ResponseSuccess(ctx, gin.H{"valid": false})
		return
	}

	// 验证成功后删除验证码
	c.redis.Del(rctx, "captcha:"+req.CaptchaID)

	c.ResponseSuccess(ctx, gin.H{"valid": true})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
