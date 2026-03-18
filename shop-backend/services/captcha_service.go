package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mojocn/base64Captcha"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	redis  *redis.Client
	driver *base64Captcha.DriverDigit
	store  base64Captcha.Store
}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService(redis *redis.Client) *CaptchaService {
	// 创建数字验证码驱动
	// 参数：高度、宽度、验证码长度、最大倾斜角度、干扰线条数
	// 增大尺寸让数字更清晰，减少干扰线数量
	driver := base64Captcha.NewDriverDigit(80, 200, 4, 0.5, 10)

	return &CaptchaService{
		redis:  redis,
		driver: driver,
		store:  &redisStore{redis: redis},
	}
}

// redisStore 实现 base64Captcha.Store 接口
type redisStore struct {
	redis *redis.Client
}

func (s *redisStore) Set(id string, value string) error {
	ctx := context.Background()
	return s.redis.Set(ctx, "captcha:"+id, value, 5*time.Minute).Err()
}

func (s *redisStore) Get(id string, clear bool) string {
	ctx := context.Background()
	key := "captcha:" + id
	value, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		s.redis.Del(ctx, key)
	}
	return value
}

func (s *redisStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == answer
}

// CaptchaImage 验证码图片结构
type CaptchaImage struct {
	ID         string
	Value      string
	Image      []byte
	Base64Code string
}

const captchaLength = 4

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha() (*CaptchaImage, error) {
	// 生成唯一的验证码ID
	captchaID, err := generateRandomID(16)
	if err != nil {
		return nil, err
	}

	// 使用 base64Captcha 生成验证码
	c := base64Captcha.NewCaptcha(s.driver, s.store)
	_, content, answer, err := c.Generate()
	if err != nil {
		return nil, err
	}

	// 存储验证码到Redis
	err = s.store.Set(captchaID, answer)
	if err != nil {
		return nil, err
	}

	// 将 base64 内容转换为图片字节
	// content 格式为 "data:image/png;base64,xxxxx"
	imageBytes, err := base64ImageToBytes(content)
	if err != nil {
		return nil, err
	}

	return &CaptchaImage{
		ID:         captchaID,
		Value:      answer,
		Image:      imageBytes,
		Base64Code: content,
	}, nil
}

// base64ImageToBytes 将 base64 图片字符串转换为字节数组
func base64ImageToBytes(base64String string) ([]byte, error) {
	// 移除 data:image/png;base64, 前缀
	if idx := strings.Index(base64String, ","); idx != -1 {
		base64String = base64String[idx+1:]
	}
	return base64.StdEncoding.DecodeString(base64String)
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(captchaID, value string) bool {
	return s.store.Verify(captchaID, value, true)
}

// GetCaptchaValue 获取验证码值（用于图片生成）
func (s *CaptchaService) GetCaptchaValue(captchaID string) (string, error) {
	return s.store.Get(captchaID, false), nil
}

// StoreCaptcha 存储验证码
func (s *CaptchaService) StoreCaptcha(captchaID, value string) error {
	return s.store.Set(captchaID, value)
}

func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
