package cache

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker 熔断器结构体
// 熔断器是一种用于防止系统故障级联的设计模式
// 当服务调用失败率达到一定阈值时，熔断器会打开，阻止请求继续发送
// 经过一段时间后，熔断器会进入半开状态，尝试发送少量请求
// 如果这些请求成功，熔断器会关闭；如果失败，熔断器会再次打开
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker // 底层的gobreaker实例
}

// NewCircuitBreaker 创建熔断器实例
// 参数：
// - name: 熔断器名称
// - failureThreshold: 失败率阈值（0-1之间）
// - resetTimeout: 从开路状态到半开状态的超时时间
// - maxRequests: 半开状态下允许的最大请求数
// 返回：
// - *CircuitBreaker: 熔断器实例
// 说明：
// 熔断器的工作原理：
// 1. 闭合状态：所有请求都允许通过
// 2. 开路状态：所有请求都被拒绝
// 3. 半开状态：允许少量请求通过，用于测试服务是否恢复
func NewCircuitBreaker(name string, failureThreshold float64, resetTimeout time.Duration, maxRequests uint32) *CircuitBreaker {
	// 配置熔断器
	settings := gobreaker.Settings{
		Name:        name,               // 熔断器名称
		MaxRequests: maxRequests,          // 半开状态下允许的最大请求数
		Interval:    time.Minute,          // 统计窗口，用于计算失败率
		Timeout:     resetTimeout,         // 从开路状态到半开状态的超时时间
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 失败率达到阈值时触发熔断
			// 当请求数达到10个以上，且失败率超过阈值时，触发熔断
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= failureThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// 状态变化回调，用于记录状态变化
			fmt.Printf("Circuit breaker %s changed from %v to %v\n", name, from, to)
		},
	}

	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute 执行函数，带熔断器保护
// 参数：
// - fn: 要执行的函数
// 返回：
// - interface{}: 函数执行结果
// - error: 执行错误
// 说明：
// 当熔断器处于闭合状态时，执行函数；
// 当熔断器处于开路状态时，直接返回错误；
// 当熔断器处于半开状态时，允许执行函数，但会根据执行结果决定熔断器状态
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(fn)
}

// State 获取当前熔断器状态
// 返回：
// - gobreaker.State: 熔断器状态（Closed/Open/HalfOpen）
// 说明：
// Closed: 闭合状态，允许所有请求通过
// Open: 开路状态，拒绝所有请求
// HalfOpen: 半开状态，允许少量请求通过
func (cb *CircuitBreaker) State() gobreaker.State {
	return cb.cb.State()
}

// Counts 获取当前统计信息
// 返回：
// - gobreaker.Counts: 统计信息，包含请求数、失败数等
// 说明：
// 用于获取熔断器的运行统计数据，包括总请求数、失败数等
func (cb *CircuitBreaker) Counts() gobreaker.Counts {
	return cb.cb.Counts()
}
