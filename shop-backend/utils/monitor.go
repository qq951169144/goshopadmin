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

// RuntimeStats 运行时统计快照，包含单次采集的所有监控指标
type RuntimeStats struct {
	GoroutineCount int         `json:"goroutine_count"`
	ModuleStats    ModuleStats `json:"module_stats"`
	MemoryStats    MemoryStats `json:"memory_stats"`
	ThreadStats    ThreadStats `json:"thread_stats"`
	MutexStats     MutexStats  `json:"mutex_stats"`
	Timestamp      time.Time   `json:"timestamp"`
	ServiceName    string      `json:"service_name"`
}

// ModuleStats 按模块分组的协程统计，key 为模块名，value 为该模块的协程数量
type ModuleStats map[string]int

// MemoryStats 内存使用统计，数据来源于 runtime.ReadMemStats
type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapSys    uint64 `json:"heap_sys"`
	StackInuse uint64 `json:"stack_inuse"`
	NumGC      uint32 `json:"num_gc"`
}

// ThreadStats 线程与 CGO 调用统计
type ThreadStats struct {
	ThreadCount  int   `json:"thread_count"`
	CgoCallCount int64 `json:"cgo_call_count"`
}

// MutexStats 互斥锁竞争统计（增量值，每次采集返回自上次以来的增量）
type MutexStats struct {
	Contentions int64   `json:"contentions"`
	Delay       float64 `json:"delay"`
}

// Monitor 运行时监控器，定时采集 Go 运行时指标并推送到 Prometheus
type Monitor struct {
	stats                []RuntimeStats
	mu                   sync.RWMutex
	alertThreshold       int
	memoryThreshold      uint64
	checkInterval        time.Duration
	quit                 chan struct{}
	maxHistorySize       int
	serviceName          string
	stackBuf             []byte
	lastGCCount          uint32
	lastCgoCallCount     int64
	lastMutexContentions int64
	lastMutexDelay       float64
}

// NewMonitor 创建监控器实例
// alertThreshold: 协程数量告警阈值，超过此值输出警告日志
// checkInterval: 采集间隔，每隔多久采集一次运行时指标
// maxHistorySize: 历史记录最大条数，超过后自动淘汰最早记录
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

	// 采样率设置：
	// SetMutexProfileFraction(5) 表示每 5 次 mutex 竞争事件采样 1 次，降低性能开销
	// 值越大采样越稀疏，0 表示关闭，1 表示全量采样（高并发下影响性能）
	runtime.SetMutexProfileFraction(5)
	// SetBlockProfileRate(100) 表示阻塞时间 >= 100 纳秒的操作才会被记录
	// 值为纳秒阈值，0 表示关闭，1 表示记录所有阻塞（高并发下影响性能）
	runtime.SetBlockProfileRate(100)

	registerPrometheusMetrics()

	return &Monitor{
		stats:           make([]RuntimeStats, 0, maxHistorySize),
		alertThreshold:  alertThreshold,
		memoryThreshold: 512 * 1024 * 1024,
		checkInterval:   checkInterval,
		quit:            make(chan struct{}),
		maxHistorySize:  maxHistorySize,
		serviceName:     "shop-backend",
		// 预分配 1MB 缓冲区用于 runtime.Stack() 读取所有协程堆栈
		// 复用此缓冲区避免每次采集都分配 1MB 内存造成 GC 压力
		stackBuf: make([]byte, 1<<20),
	}
}

// Start 启动监控器，在后台协程中定时采集指标
func (m *Monitor) Start() {
	go m.collectMetrics()
}

// Stop 停止监控器，关闭退出信号通道，后台采集协程会安全退出
func (m *Monitor) Stop() {
	close(m.quit)
}

// collectMetrics 采集循环，定时触发采集或响应退出信号
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

// collectOnce 执行一次完整的指标采集、存储和推送流程
func (m *Monitor) collectOnce() {
	stats := m.collectRuntimeStats()

	// 将采集结果存入历史记录（环形缓冲区策略）
	m.mu.Lock()
	m.stats = append(m.stats, stats)
	if len(m.stats) > m.maxHistorySize {
		m.stats = m.stats[1:]
	}
	m.mu.Unlock()

	// 计算 GC 增量：本次采集的 GC 次数减去上次记录的 GC 次数
	// Prometheus Counter 只能递增，因此需要计算增量后 Add
	// 转为 int 避免无符号整数下溢导致负数判断失效
	gcDelta := int(stats.MemoryStats.NumGC) - int(m.lastGCCount)
	if gcDelta < 0 {
		gcDelta = 0
	}
	m.lastGCCount = stats.MemoryStats.NumGC

	// 计算 CGO 调用增量：runtime.NumCgoCall() 返回累积值
	// 与 GC 相同，需要计算增量后通过 Counter.Add 推送
	cgoDelta := stats.ThreadStats.CgoCallCount - m.lastCgoCallCount
	if cgoDelta < 0 {
		cgoDelta = 0
	}
	m.lastCgoCallCount = stats.ThreadStats.CgoCallCount

	// 将本次采集的指标推送到 Prometheus
	updatePrometheusMetrics(&stats, uint32(gcDelta), cgoDelta)

	// 检查是否触发告警条件
	m.checkAlerts(&stats)

	Info("协程数量: %d, 堆内存: %.2f MB, 线程数: %d",
		stats.GoroutineCount,
		float64(stats.MemoryStats.HeapAlloc)/1024/1024,
		stats.ThreadStats.ThreadCount)
}

// collectRuntimeStats 采集一次完整的运行时指标快照
func (m *Monitor) collectRuntimeStats() RuntimeStats {
	// runtime.ReadMemStats 会触发 STW（Stop-The-World），
	// 但时间极短（微秒级），对业务影响可忽略
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 解析所有协程的堆栈信息，按模块分组统计
	moduleStats := m.collectModuleStats()

	// 采集互斥锁竞争统计（增量值）
	mutexStats := m.collectMutexStats()

	// 通过 pprof 获取线程创建数量
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

// getThreadCount 通过 pprof 获取当前 OS 线程数量
// 使用 pprof.Lookup("threadcreate").Count() 而非 runtime 直接 API
// 因为 Go 标准库没有直接暴露线程计数的 API
func getThreadCount() int {
	p := pprof.Lookup("threadcreate")
	if p != nil {
		return p.Count()
	}
	return 0
}

// collectMutexStats 采集互斥锁竞争统计，返回自上次采集以来的增量值
// 使用 pprof.Lookup("mutex") 获取 mutex profile 数据
func (m *Monitor) collectMutexStats() MutexStats {
	p := pprof.Lookup("mutex")
	if p == nil {
		return MutexStats{}
	}

	// 使用 p.Count() 获取竞争总次数（比解析文本更可靠）
	currentContentions := int64(p.Count())

	// 计算增量：本次总竞争次数 - 上次记录的总竞争次数
	deltaContentions := currentContentions - m.lastMutexContentions
	if deltaContentions < 0 {
		deltaContentions = 0
	}
	m.lastMutexContentions = currentContentions

	// 解析 mutex profile 文本获取延迟数据（纳秒级精度）
	var buf bytes.Buffer
	p.WriteTo(&buf, 1)
	_, totalDelay := parseMutexProfileText(buf.String())

	// 计算延迟增量
	deltaDelay := totalDelay - m.lastMutexDelay
	if deltaDelay < 0 {
		deltaDelay = 0
	}
	m.lastMutexDelay = totalDelay

	return MutexStats{
		Contentions: deltaContentions,
		Delay:       deltaDelay,
	}
}

// parseMutexProfileText 解析 mutex pprof 文本输出，提取竞争次数和延迟总和
// pprof mutex 文本格式每行包含：竞争次数 延迟周期数，后跟堆栈信息
// 例如：5 1234567 \n stack_trace...
func parseMutexProfileText(data string) (totalContentions int64, totalDelay float64) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行、分隔线、注释行和标题行
		if line == "" || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "cycles") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// 第一个字段：竞争次数
		count, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}

		// 第二个字段：延迟周期数（cycles），需要转换为纳秒
		delay, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		totalContentions += count
		totalDelay += delay
	}

	return totalContentions, totalDelay
}

// checkAlerts 检查是否触发告警条件，当前仅输出日志
func (m *Monitor) checkAlerts(stats *RuntimeStats) {
	if stats.GoroutineCount > m.alertThreshold {
		Warn("协程数量超过阈值: 当前=%d, 阈值=%d", stats.GoroutineCount, m.alertThreshold)
	}
	if stats.MemoryStats.HeapAlloc > m.memoryThreshold {
		Warn("堆内存使用超过阈值: 当前=%d MB, 阈值=%d MB",
			stats.MemoryStats.HeapAlloc/1024/1024, m.memoryThreshold/1024/1024)
	}
}

// GetCurrentStats 获取最新一次采集的运行时统计快照
// 如果历史记录为空（监控器刚启动尚未采集），则立即执行一次采集
func (m *Monitor) GetCurrentStats() RuntimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.stats) == 0 {
		return m.collectRuntimeStats()
	}

	return m.stats[len(m.stats)-1]
}

// GetHistoryStats 获取所有历史统计记录的副本
// 返回副本而非引用，避免外部修改影响内部数据
func (m *Monitor) GetHistoryStats() []RuntimeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]RuntimeStats, len(m.stats))
	copy(result, m.stats)
	return result
}
