package services

import (
	"goshopadmin/utils"
	"sync"
	"time"
)

// CaptchaService 验证码服务
type CaptchaService struct {
	captchaStore map[string]int
	mutex        sync.RWMutex
}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		captchaStore: make(map[string]int),
	}
}

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha() (*utils.Captcha, error) {
	// 生成验证码
	captcha, err := utils.GenerateCaptcha(300, 150, 50)
	if err != nil {
		return nil, err
	}

	// 存储验证码答案
	s.mutex.Lock()
	s.captchaStore[captcha.ID] = captcha.Answer
	s.mutex.Unlock()

	// 设置过期时间
	go func(id string) {
		time.Sleep(5 * time.Minute)
		s.mutex.Lock()
		delete(s.captchaStore, id)
		s.mutex.Unlock()
	}(captcha.ID)

	return captcha, nil
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(id string, answer int) bool {
	s.mutex.RLock()
	correctAnswer, exists := s.captchaStore[id]
	s.mutex.RUnlock()

	if !exists {
		utils.Error("验证码不存在: id=%s", id)
		return false
	}

	// 验证答案
	isCorrect := utils.VerifyCaptcha(answer, correctAnswer, 20)

	// 记录验证信息
	utils.Info("验证码验证: id=%s, 用户答案=%d, 正确答案=%d, 验证结果=%t", id, answer, correctAnswer, isCorrect)

	// 验证后删除验证码
	if isCorrect {
		utils.Info("验证码验证成功，删除验证码: id=%s", id)
		s.mutex.Lock()
		delete(s.captchaStore, id)
		s.mutex.Unlock()
	}

	return isCorrect
}
