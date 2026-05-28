package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	goroutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_goroutine_count",
		Help: "Current number of goroutines",
	})

	goroutineModuleCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shop_goroutine_module_count",
		Help: "Number of goroutines per module",
	}, []string{"module"})

	memoryAllocBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_alloc_bytes",
		Help: "Bytes of allocated heap objects",
	})

	memorySysBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_sys_bytes",
		Help: "Total bytes of memory obtained from the OS",
	})

	memoryHeapAllocBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_heap_alloc_bytes",
		Help: "Bytes of allocated heap objects",
	})

	memoryStackInuseBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_memory_stack_inuse_bytes",
		Help: "Bytes in stack spans",
	})

	gcCountTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_gc_count_total",
		Help: "Total number of GC cycles",
	})

	threadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_thread_count",
		Help: "Current number of OS threads",
	})

	cgoCallCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shop_cgo_call_count",
		Help: "Current number of CGO calls",
	})

	mutexContentionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_mutex_contentions_total",
		Help: "Total number of mutex contentions",
	})

	mutexDelaySecondsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "shop_mutex_delay_seconds_total",
		Help: "Total time spent waiting for mutex in seconds",
	})

	prometheusRegistered = false
)

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
		cgoCallCount,
		mutexContentionsTotal,
		mutexDelaySecondsTotal,
	)
}

func updatePrometheusMetrics(stats *RuntimeStats, gcDelta uint32) {
	goroutineCount.Set(float64(stats.GoroutineCount))

	for module, count := range stats.ModuleStats {
		goroutineModuleCount.WithLabelValues(module).Set(float64(count))
	}

	memoryAllocBytes.Set(float64(stats.MemoryStats.Alloc))
	memorySysBytes.Set(float64(stats.MemoryStats.Sys))
	memoryHeapAllocBytes.Set(float64(stats.MemoryStats.HeapAlloc))
	memoryStackInuseBytes.Set(float64(stats.MemoryStats.StackInuse))

	threadCount.Set(float64(stats.ThreadStats.ThreadCount))
	cgoCallCount.Set(float64(stats.ThreadStats.CgoCallCount))

	gcCountTotal.Add(float64(gcDelta))

	mutexContentionsTotal.Add(float64(stats.MutexStats.Contentions))
	mutexDelaySecondsTotal.Add(stats.MutexStats.Delay / 1e9)
}
