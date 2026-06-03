package middleware

import (
	"net/http"
	"sync"
	"time"

	svcErrors "search-service/errors"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 限流中间件
// 使用令牌桶算法限制请求频率，防止恶意请求压垮搜索服务
// 默认限制为 50 QPS（每秒请求数）
// ============================================================

// RateLimiter 令牌桶限流器
// 令牌桶算法原理：
// 1. 桶中初始有 maxTokens 个令牌
// 2. 每秒往桶中添加 rate 个令牌，不超过 maxTokens
// 3. 每个请求消耗一个令牌
// 4. 桶为空时拒绝请求
type RateLimiter struct {
	// rate 每秒添加的令牌数
	rate int

	// maxTokens 桶中最大令牌数
	maxTokens int

	// currentTokens 当前令牌数
	currentTokens int

	// lastRefillTime 上次补充令牌的时间
	lastRefillTime time.Time

	// mu 互斥锁，保证并发安全
	mu sync.Mutex
}

// NewRateLimiter 创建令牌桶限流器
// rate: 每秒允许的请求数
// maxTokens: 桶中最大令牌数（允许短时突发）
func NewRateLimiter(rate, maxTokens int) *RateLimiter {
	return &RateLimiter{
		rate:           rate,
		maxTokens:      maxTokens,
		currentTokens:  maxTokens, // 初始满桶
		lastRefillTime: time.Now(),
	}
}

// globalLimiter 全局限流器实例，限制 50 QPS
var globalLimiter = NewRateLimiter(50, 100)

// Allow 判断是否允许请求通过
// 返回 true 表示允许，false 表示被限流
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()
	tokensToAdd := int(elapsed * float64(rl.rate))

	if tokensToAdd > 0 {
		rl.currentTokens += tokensToAdd
		if rl.currentTokens > rl.maxTokens {
			rl.currentTokens = rl.maxTokens
		}
		rl.lastRefillTime = now
	}

	// 检查是否有可用令牌
	if rl.currentTokens > 0 {
		rl.currentTokens--
		return true
	}

	return false
}

// RateLimit 限流中间件
// 检查全局限流器是否允许当前请求
// 被限流的请求返回 429 Too Many Requests
func RateLimit() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !globalLimiter.Allow() {
			utils.Warn("请求被限流: %s %s", ctx.Request.Method, ctx.Request.URL.Path)

			ctx.JSON(http.StatusOK, gin.H{
				"code":    svcErrors.CodeSearchRateLimited,
				"message": svcErrors.GetErrorMessage(svcErrors.CodeSearchRateLimited),
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
