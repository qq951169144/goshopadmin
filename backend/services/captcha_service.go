package services

import (
	"goshopadmin/utils"
	"sync"
	"time"
)

// CaptchaItem 验证码条目，包含答案和过期时间
type CaptchaItem struct {
	Answer      int
	ExpireTime  time.Time
}

// CaptchaService 验证码服务
type CaptchaService struct {
	captchaStore map[string]CaptchaItem
	mutex        sync.RWMutex
	isRunning    bool
	stopChan     chan struct{}
}

// NewCaptchaService 创建验证码服务实例
func NewCaptchaService() *CaptchaService {
	service := &CaptchaService{
		captchaStore: make(map[string]CaptchaItem),
		isRunning:    false,
		stopChan:     make(chan struct{}),
	}
	// 启动定时清理任务
	service.startCleanupTask()
	return service
}

// startCleanupTask 启动定时清理过期验证码的任务
func (s *CaptchaService) startCleanupTask() {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return
	}
	s.isRunning = true
	s.mutex.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpiredCaptchas()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// cleanupExpiredCaptchas 清理过期的验证码
func (s *CaptchaService) cleanupExpiredCaptchas() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for id, item := range s.captchaStore {
		if now.After(item.ExpireTime) {
			delete(s.captchaStore, id)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		utils.Info("清理过期验证码: 共清理 %d 个", expiredCount)
	}
}

// Stop 停止验证码服务
func (s *CaptchaService) Stop() {
	s.mutex.Lock()
	if s.isRunning {
		s.isRunning = false
		s.stopChan <- struct{}{}
	}
	s.mutex.Unlock()
}

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha() (*utils.Captcha, error) {
	// 生成验证码
	captcha, err := utils.GenerateCaptcha(300, 150, 50)
	if err != nil {
		return nil, err
	}

	// 存储验证码答案和过期时间（5分钟后过期）
	s.mutex.Lock()
	s.captchaStore[captcha.ID] = CaptchaItem{
		Answer:     captcha.Answer,
		ExpireTime: time.Now().Add(5 * time.Minute),
	}
	s.mutex.Unlock()

	return captcha, nil
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(id string, answer int) bool {
	s.mutex.RLock()
	item, exists := s.captchaStore[id]
	s.mutex.RUnlock()

	if !exists {
		utils.Error("验证码不存在: id=%s", id)
		return false
	}

	// 检查验证码是否过期
	if time.Now().After(item.ExpireTime) {
		utils.Error("验证码已过期: id=%s", id)
		s.mutex.Lock()
		delete(s.captchaStore, id)
		s.mutex.Unlock()
		return false
	}

	// 验证答案
	isCorrect := utils.VerifyCaptcha(answer, item.Answer, 20)

	// 记录验证信息
	utils.Info("验证码验证: id=%s, 用户答案=%d, 正确答案=%d, 验证结果=%t", id, answer, item.Answer, isCorrect)

	// 验证后删除验证码
	if isCorrect {
		utils.Info("验证码验证成功，删除验证码: id=%s", id)
		s.mutex.Lock()
		delete(s.captchaStore, id)
		s.mutex.Unlock()
	}

	return isCorrect
}
