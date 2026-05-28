package utils

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// GoroutineMetrics 协程监控指标
type GoroutineMetrics struct {
	Total           int               `json:"total"`              // 总协程数
	System          int               `json:"system"`             // 系统协程数（runtime内部）
	Application     int               `json:"application"`        // 应用协程数
	ByModule        map[string]int    `json:"by_module"`          // 按模块统计
	Timestamp       time.Time         `json:"timestamp"`          // 时间戳
}

// Monitor 协程监控器
type Monitor struct {
	mu              sync.Mutex
	isRunning       bool
	metrics         GoroutineMetrics
	moduleCounters  map[string]int
	ticker          *time.Ticker
	stopChan        chan struct{}
}

// NewMonitor 创建协程监控器实例
func NewMonitor() *Monitor {
	return &Monitor{
		moduleCounters: make(map[string]int),
		stopChan:       make(chan struct{}),
	}
}

// Start 启动监控器
func (m *Monitor) Start(interval time.Duration) {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = true
	m.mu.Unlock()

	m.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.updateMetrics()
			case <-m.stopChan:
				return
			}
		}
	}()
}

// Stop 停止监控器
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return
	}

	m.isRunning = false
	m.stopChan <- struct{}{}
	m.ticker.Stop()
}

// updateMetrics 更新监控指标
func (m *Monitor) updateMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := runtime.NumGoroutine()
	
	// 估算系统协程数（通常为 runtime 内部协程，如垃圾回收等）
	system := estimateSystemGoroutines()
	application := total - system

	m.metrics = GoroutineMetrics{
		Total:       total,
		System:      system,
		Application: application,
		ByModule:    make(map[string]int),
		Timestamp:   time.Now(),
	}

	for module, count := range m.moduleCounters {
		m.metrics.ByModule[module] = count
	}
}

// estimateSystemGoroutines 估算系统协程数
// 系统协程包括：主协程、GC协程、网络poller协程等
func estimateSystemGoroutines() int {
	// 基础系统协程数：主协程 + GC协程 + 网络poller
	return 4
}

// GetCurrentMetrics 获取当前监控指标
func (m *Monitor) GetCurrentMetrics() GoroutineMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metrics
}

// IncrementModuleCounter 增加模块协程计数
func (m *Monitor) IncrementModuleCounter(module string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moduleCounters[module]++
}

// DecrementModuleCounter 减少模块协程计数
func (m *Monitor) DecrementModuleCounter(module string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.moduleCounters[module] > 0 {
		m.moduleCounters[module]--
	}
}

// RegisterHTTPHandlers 注册监控API路由
func (m *Monitor) RegisterHTTPHandlers(r *gin.RouterGroup) {
	r.GET("/goroutines", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    m.GetCurrentMetrics(),
		})
	})

	r.GET("/goroutines/total", func(c *gin.Context) {
		metrics := m.GetCurrentMetrics()
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"total": metrics.Total,
			},
		})
	})

	r.GET("/goroutines/module", func(c *gin.Context) {
		metrics := m.GetCurrentMetrics()
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    metrics.ByModule,
		})
	})
}

// GetGoroutineCount 获取当前协程总数
func (m *Monitor) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}