package services

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
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	redis *redis.Client
}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService(redis *redis.Client) *CaptchaService {
	return &CaptchaService{redis: redis}
}

// CaptchaImage 验证码图片结构
type CaptchaImage struct {
	ID     string
	Value  string
	Image []byte
}

const captchaLength = 4

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha() (*CaptchaImage, error) {
	// 生成随机验证码
	captcha := ""
	for i := 0; i < captchaLength; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		captcha += strconv.Itoa(int(n.Int64()))
	}

	// 生成唯一的验证码ID
	captchaID, err := generateRandomID(16)
	if err != nil {
		return nil, err
	}

	// 存储验证码到Redis，设置5分钟过期
	ctx := context.Background()
	err = s.redis.Set(ctx, "captcha:"+captchaID, captcha, 5*time.Minute).Err()
	if err != nil {
		return nil, err
	}

	// 生成验证码图片
	img := generateCaptchaImage(captcha)

	// 将图片编码为PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return &CaptchaImage{
		ID:    captchaID,
		Value: captcha,
		Image: buf.Bytes(),
	}, nil
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(captchaID, value string) bool {
	ctx := context.Background()
	storedCaptcha, err := s.redis.Get(ctx, "captcha:"+captchaID).Result()
	if err != nil || storedCaptcha != value {
		return false
	}

	// 验证成功后删除验证码
	s.redis.Del(ctx, "captcha:"+captchaID)
	return true
}

// GetCaptchaValue 获取验证码值（用于图片生成）
func (s *CaptchaService) GetCaptchaValue(captchaID string) (string, error) {
	ctx := context.Background()
	return s.redis.Get(ctx, "captcha:"+captchaID).Result()
}

// StoreCaptcha 存储验证码
func (s *CaptchaService) StoreCaptcha(captchaID, value string) error {
	ctx := context.Background()
	return s.redis.Set(ctx, "captcha:"+captchaID, value, 5*time.Minute).Err()
}

func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateCaptchaImage(captcha string) image.Image {
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

	return img
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

// EncodeCaptchaImage 将验证码图片编码为PNG字节
func EncodeCaptchaImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CaptchaHandler 处理验证码HTTP请求
func (s *CaptchaService) CaptchaHandler(w http.ResponseWriter, r *http.Request) {
	captcha, err := s.GenerateCaptcha()
	if err != nil {
		http.Error(w, "Failed to generate captcha", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Captcha-ID", captcha.ID)
	w.Write(captcha.Image)
}
