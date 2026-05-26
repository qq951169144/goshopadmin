package utils

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// GoroutineMetrics 协程指标快照
type GoroutineMetrics struct {
	TotalGoroutines int              `json:"total_goroutines"` // 当前总协程数
	ModuleStats     map[string]int   `json:"module_stats"`     // 分模块协程统计
	Timestamp       time.Time        `json:"timestamp"`        // 采集时间
}

// Monitor 协程监控器
type Monitor struct {
	metrics        []GoroutineMetrics // 历史指标数据
	mu             sync.RWMutex       // 互斥锁
	alertThreshold int                // 告警阈值
	checkInterval  time.Duration      // 检查间隔
	quit           chan struct{}      // 退出信号
	maxHistorySize int                // 最大历史记录数
}

// NewMonitor 创建协程监控器
func NewMonitor(alertThreshold int, checkInterval time.Duration, maxHistorySize int) *Monitor {
	if alertThreshold <= 0 {
		alertThreshold = 1000
	}
	if checkInterval <= 0 {
		checkInterval = 10 * time.Second
	}
	if maxHistorySize <= 0 {
		maxHistorySize = 100
	}

	return &Monitor{
		metrics:        make([]GoroutineMetrics, 0, maxHistorySize),
		alertThreshold: alertThreshold,
		checkInterval:  checkInterval,
		quit:           make(chan struct{}),
		maxHistorySize: maxHistorySize,
	}
}

// Start 启动监控
func (m *Monitor) Start() {
	go m.collectMetrics()
}

// Stop 停止监控
func (m *Monitor) Stop() {
	close(m.quit)
}

// collectMetrics 定期采集协程指标
func (m *Monitor) collectMetrics() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectOnce()
		case <-m.quit:
			return
		}
	}
}

// collectOnce 采集一次协程指标
func (m *Monitor) collectOnce() {
	metrics := GoroutineMetrics{
		TotalGoroutines: runtime.NumGoroutine(),
		ModuleStats:     make(map[string]int),
		Timestamp:       time.Now(),
	}

	m.mu.Lock()
	m.metrics = append(m.metrics, metrics)
	if len(m.metrics) > m.maxHistorySize {
		m.metrics = m.metrics[1:]
	}
	m.mu.Unlock()

	// 检查是否超过阈值
	if metrics.TotalGoroutines > m.alertThreshold {
		Warn("协程数量超过阈值: 当前=%d, 阈值=%d", metrics.TotalGoroutines, m.alertThreshold)
	}

	Info("当前协程数量: %d", metrics.TotalGoroutines)
}

// GetCurrentMetrics 获取当前协程指标
func (m *Monitor) GetCurrentMetrics() GoroutineMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.metrics) == 0 {
		return GoroutineMetrics{
			TotalGoroutines: runtime.NumGoroutine(),
			ModuleStats:     make(map[string]int),
			Timestamp:       time.Now(),
		}
	}

	return m.metrics[len(m.metrics)-1]
}

// GetHistoryMetrics 获取历史指标数据
func (m *Monitor) GetHistoryMetrics() []GoroutineMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]GoroutineMetrics, len(m.metrics))
	copy(result, m.metrics)
	return result
}

// RegisterHTTPHandlers 注册HTTP接口
func (m *Monitor) RegisterHTTPHandlers() {
	http.HandleFunc("/metrics/goroutines", func(w http.ResponseWriter, r *http.Request) {
		metrics := m.GetCurrentMetrics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	http.HandleFunc("/metrics/goroutines/history", func(w http.ResponseWriter, r *http.Request) {
		history := m.GetHistoryMetrics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metrics": history,
		})
	})

	Info("协程监控HTTP接口已注册")
}
