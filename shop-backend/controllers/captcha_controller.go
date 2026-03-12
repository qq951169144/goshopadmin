package controllers

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"shop-backend/config"

	"github.com/gin-gonic/gin"
)

// 验证码长度
const captchaLength = 4

// 生成验证码
func GenerateCaptcha(c *gin.Context) {
	// 生成随机验证码
	rand.Seed(time.Now().UnixNano())
	captcha := ""
	for i := 0; i < captchaLength; i++ {
		captcha += strconv.Itoa(rand.Intn(10))
	}

	// 生成唯一的验证码ID
	captchaID, err := generateRandomID(16)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to generate captcha ID")
		return
	}

	// 存储验证码到Redis，设置5分钟过期
	ctx := context.Background()
	err = config.Redis.Set(ctx, "captcha:"+captchaID, captcha, 5*time.Minute).Err()
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to store captcha")
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
				if rand.Intn(2) == 0 {
					img.Set(x+j, y-k, textColor)
				}
			}
		}
	}

	// 绘制干扰线
	for i := 0; i < 5; i++ {
		x1 := rand.Intn(width)
		y1 := rand.Intn(height)
		x2 := rand.Intn(width)
		y2 := rand.Intn(height)
		for x := min(x1, x2); x <= max(x1, x2); x++ {
			for y := min(y1, y2); y <= max(y1, y2); y++ {
				if rand.Intn(10) == 0 {
					img.Set(x, y, textColor)
				}
			}
		}
	}

	// 设置响应头
	c.Header("Content-Type", "image/png")
	c.Header("X-Captcha-ID", captchaID)

	// 输出图片
	png.Encode(c.Writer, img)
}

// 验证验证码
func VerifyCaptcha(c *gin.Context) {
	type VerifyRequest struct {
		CaptchaID string `json:"captcha_id" binding:"required"`
		Value     string `json:"value" binding:"required"`
	}

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从Redis获取验证码并验证
	ctx := context.Background()
	storedCaptcha, err := config.Redis.Get(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || storedCaptcha != req.Value {
		ResponseSuccess(c, gin.H{"valid": false})
		return
	}

	// 验证成功后删除验证码
	config.Redis.Del(ctx, "captcha:"+req.CaptchaID)

	ResponseSuccess(c, gin.H{"valid": true})
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

// 生成随机ID
func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := crand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
