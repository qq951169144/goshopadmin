package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	// shop_goroutine_count 当前协程总数（Gauge：可增可减的当前值）
	goroutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_goroutine_count",
		Help: "Current number of goroutines",
	})

	// shop_goroutine_module_count 按模块分组的协程数量（GaugeVec：带标签的当前值）
	// 标签 module 标识协程所属模块（如 services、controllers 等）
	goroutineModuleCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shop_goroutine_module_count",
		Help: "Number of goroutines per module",
	}, []string{"module"})

	// shop_memory_alloc_bytes 已分配的堆对象字节数（Gauge）
	memoryAllocBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_alloc_bytes",
		Help: "Bytes of allocated heap objects",
	})

	// shop_memory_sys_bytes 从 OS 获取的总内存字节数（Gauge）
	memorySysBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_sys_bytes",
		Help: "Total bytes of memory obtained from the OS",
	})

	// shop_memory_heap_alloc_bytes 已分配的堆对象字节数（Gauge）
	memoryHeapAllocBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_heap_alloc_bytes",
		Help: "Bytes of allocated heap objects",
	})

	// shop_memory_stack_inuse_bytes 栈区使用字节数（Gauge）
	memoryStackInuseBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_stack_inuse_bytes",
		Help: "Bytes in stack spans",
	})

	// shop_gc_count_total GC 总次数（Counter：单调递增的累计值）
	// 使用 Counter 而非 Gauge，因为 GC 次数只会递增
	// 每次采集时 Add 增量值，而非 Set 绝对值
	gcCountTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_gc_count_total",
		Help: "Total number of GC cycles",
	})

	// shop_thread_count 当前 OS 线程数量（Gauge）
	threadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_thread_count",
		Help: "Current number of OS threads",
	})

	// shop_cgo_call_count_total CGO 调用总次数（Counter：单调递增的累计值）
	// runtime.NumCgoCall() 返回累积值，需要计算增量后 Add
	// 命名遵循 Prometheus Counter 规范：使用 _total 后缀
	cgoCallCountTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_cgo_call_count_total",
		Help: "Total number of CGO calls",
	})

	// shop_mutex_contentions_total 互斥锁竞争总次数（Counter）
	mutexContentionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_mutex_contentions_total",
		Help: "Total number of mutex contentions",
	})

	// shop_mutex_delay_seconds_total 互斥锁等待延迟总秒数（Counter）
	mutexDelaySecondsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_mutex_delay_seconds_total",
		Help: "Total time spent waiting for mutex in seconds",
	})

	// prometheusRegistered 确保指标只注册一次的标志位
	// MustRegister 重复注册会 panic，此标志位实现幂等注册
	prometheusRegistered = false
)

// registerPrometheusMetrics 注册所有自定义 Prometheus 指标
// 使用 prometheusRegistered 标志位保证幂等性，多次调用不会重复注册
// 同时注册 Go 运行时默认指标（go_goroutines、go_memstats_* 等）
func registerPrometheusMetrics() {
	if prometheusRegistered {
		return
	}
	prometheusRegistered = true

	prometheus.MustRegister(
		collectors.NewGoCollector(),
		goroutineCount,
		goroutineModuleCount,
		memoryAllocBytes,
		memorySysBytes,
		memoryHeapAllocBytes,
		memoryStackInuseBytes,
		gcCountTotal,
		threadCount,
		cgoCallCountTotal,
		mutexContentionsTotal,
		mutexDelaySecondsTotal,
	)
}

// updatePrometheusMetrics 将采集到的运行时指标推送到 Prometheus
// stats: 本次采集的运行时统计快照
// gcDelta: 自上次采集以来的 GC 增量次数
// cgoDelta: 自上次采集以来的 CGO 调用增量次数
func updatePrometheusMetrics(stats *RuntimeStats, gcDelta uint32, cgoDelta int64) {
	// Gauge 类型指标：直接 Set 当前值（可增可减）
	goroutineCount.Set(float64(stats.GoroutineCount))

	// GaugeVec 标签泄漏修复：
	// 先 Reset() 清除所有标签，再重新设置当前存在的模块
	// 避免模块协程数降为 0 后标签仍然残留在 Prometheus 中
	goroutineModuleCount.Reset()
	for module, count := range stats.ModuleStats {
		goroutineModuleCount.WithLabelValues(module).Set(float64(count))
	}

	// 内存指标：直接 Set 当前值
	memoryAllocBytes.Set(float64(stats.MemoryStats.Alloc))
	memorySysBytes.Set(float64(stats.MemoryStats.Sys))
	memoryHeapAllocBytes.Set(float64(stats.MemoryStats.HeapAlloc))
	memoryStackInuseBytes.Set(float64(stats.MemoryStats.StackInuse))

	// 线程数：直接 Set 当前值
	threadCount.Set(float64(stats.ThreadStats.ThreadCount))

	// Counter 类型指标：Add 增量值（不能 Set，Counter 只能单调递增）
	// GC 次数增量
	gcCountTotal.Add(float64(gcDelta))
	// CGO 调用次数增量
	cgoCallCountTotal.Add(float64(cgoDelta))

	// 互斥锁竞争指标：Add 增量值
	// Contentions 已经是增量值（在 collectMutexStats 中计算过 delta）
	mutexContentionsTotal.Add(float64(stats.MutexStats.Contentions))
	// Delay 单位从纳秒转换为秒（1e9 = 10^9 纳秒 = 1 秒）
	mutexDelaySecondsTotal.Add(stats.MutexStats.Delay / 1e9)
}
