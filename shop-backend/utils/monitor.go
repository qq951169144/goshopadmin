package utils

import (
	"bufio"
	"bytes"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GoroutineMetrics struct {
	TotalGoroutines int            `json:"total_goroutines"`
	ModuleStats     map[string]int `json:"module_stats"`
	Timestamp       time.Time      `json:"timestamp"`
}

type RuntimeStats struct {
	GoroutineCount int           `json:"goroutine_count"`
	ModuleStats    ModuleStats   `json:"module_stats"`
	MemoryStats    MemoryStats   `json:"memory_stats"`
	ThreadStats    ThreadStats   `json:"thread_stats"`
	MutexStats     MutexStats    `json:"mutex_stats"`
	Timestamp      time.Time     `json:"timestamp"`
	ServiceName    string        `json:"service_name"`
}

type ModuleStats map[string]int

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapSys    uint64 `json:"heap_sys"`
	StackInuse uint64 `json:"stack_inuse"`
	NumGC      uint32 `json:"num_gc"`
}

type ThreadStats struct {
	ThreadCount  int   `json:"thread_count"`
	CgoCallCount int64 `json:"cgo_call_count"`
}

type MutexStats struct {
	Contentions int64   `json:"contentions"`
	Delay       float64 `json:"delay"`
}

type Monitor struct {
	stats                []RuntimeStats
	mu                   sync.RWMutex
	alertThreshold       int
	memoryThreshold      uint64
	checkInterval        time.Duration
	quit                 chan struct{}
	maxHistorySize       int
	serviceName          string
	lastGCCount          uint32
	lastMutexContentions int64
	lastMutexDelay       float64
}

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

	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	registerPrometheusMetrics()

	return &Monitor{
		stats:           make([]RuntimeStats, 0, maxHistorySize),
		alertThreshold:  alertThreshold,
		memoryThreshold: 512 * 1024 * 1024,
		checkInterval:   checkInterval,
		quit:            make(chan struct{}),
		maxHistorySize:  maxHistorySize,
		serviceName:     "shop-backend",
	}
}

func (m *Monitor) Start() {
	go m.collectMetrics()
}

func (m *Monitor) Stop() {
	close(m.quit)
}

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

func (m *Monitor) collectOnce() {
	stats := m.collectRuntimeStats()

	m.mu.Lock()
	m.stats = append(m.stats, stats)
	if len(m.stats) > m.maxHistorySize {
		m.stats = m.stats[1:]
	}
	m.mu.Unlock()

	updatePrometheusMetrics(&stats)

	m.checkAlerts(&stats)

	Info("协程数量: %d, 堆内存: %.2f MB, 线程数: %d",
		stats.GoroutineCount,
		float64(stats.MemoryStats.HeapAlloc)/1024/1024,
		stats.ThreadStats.ThreadCount)
}

func (m *Monitor) collectRuntimeStats() RuntimeStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	moduleStats := collectModuleStats()

	mutexStats := m.collectMutexStats()

	threadCount := getThreadCount()

	return RuntimeStats{
		GoroutineCount: runtime.NumGoroutine(),
		ModuleStats:    moduleStats,
		MemoryStats: MemoryStats{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			HeapAlloc:  memStats.HeapAlloc,
			HeapSys:    memStats.HeapSys,
			StackInuse: memStats.StackInuse,
			NumGC:      memStats.NumGC,
		},
		ThreadStats: ThreadStats{
			ThreadCount:  threadCount,
			CgoCallCount: runtime.NumCgoCall(),
		},
		MutexStats:  mutexStats,
		Timestamp:   time.Now(),
		ServiceName: m.serviceName,
	}
}

func getThreadCount() int {
	p := pprof.Lookup("threadcreate")
	if p != nil {
		return p.Count()
	}
	return 0
}

func (m *Monitor) collectMutexStats() MutexStats {
	p := pprof.Lookup("mutex")
	if p == nil {
		return MutexStats{}
	}

	var buf bytes.Buffer
	p.WriteTo(&buf, 1)

	totalContentions, totalDelay := parseMutexProfileText(buf.String())

	deltaContentions := totalContentions - m.lastMutexContentions
	deltaDelay := totalDelay - m.lastMutexDelay

	if deltaContentions < 0 {
		deltaContentions = 0
	}
	if deltaDelay < 0 {
		deltaDelay = 0
	}

	m.lastMutexContentions = totalContentions
	m.lastMutexDelay = totalDelay

	return MutexStats{
		Contentions: deltaContentions,
		Delay:       deltaDelay,
	}
}

func parseMutexProfileText(data string) (totalContentions int64, totalDelay float64) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "cycles") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		count, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}

		delay, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		totalContentions += count
		totalDelay += delay
	}

	return totalContentions, totalDelay
}

func (m *Monitor) checkAlerts(stats *RuntimeStats) {
	if stats.GoroutineCount > m.alertThreshold {
		Warn("协程数量超过阈值: 当前=%d, 阈值=%d", stats.GoroutineCount, m.alertThreshold)
	}
	if stats.MemoryStats.HeapAlloc > m.memoryThreshold {
		Warn("堆内存使用超过阈值: 当前=%d MB, 阈值=%d MB",
			stats.MemoryStats.HeapAlloc/1024/1024, m.memoryThreshold/1024/1024)
	}
}

func (m *Monitor) GetCurrentStats() RuntimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.stats) == 0 {
		return m.collectRuntimeStats()
	}

	return m.stats[len(m.stats)-1]
}

func (m *Monitor) GetHistoryStats() []RuntimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]RuntimeStats, len(m.stats))
	copy(result, m.stats)
	return result
}
