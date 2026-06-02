# 系统监控增强 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 增强 shop-backend 运行时监控能力，通过 Prometheus + Grafana 实现可视化监控，在管理后台侧边栏链接跳转 Grafana。

**架构：** shop-backend 暴露 `/metrics` 端点供 Prometheus 采集运行时统计，暴露认证保护的 `/debug/pprof/*` 端点供按需 Profile 分析，Prometheus + Grafana 以 Docker 容器部署，管理后台侧边栏添加链接跳转 Grafana。

**技术栈：** Go / Gin / prometheus/client_golang / gin-contrib/pprof / Prometheus / Grafana / Docker Compose / Vue 3 / Element Plus

---

## 文件结构

| 文件 | 职责 |
|:---|:---|
| `shop-backend/utils/monitor.go` | Monitor 核心结构，采集运行时统计，管理历史数据 |
| `shop-backend/utils/monitor_prometheus.go` | Prometheus 指标定义和更新逻辑 |
| `shop-backend/utils/monitor_module.go` | ModuleStats 堆栈解析逻辑 |
| `shop-backend/controllers/monitor_controller.go` | 监控 API 控制器（stats/history） |
| `shop-backend/routes/routes.go` | 注册监控和 pprof 路由到 Gin |
| `shop-backend/main.go` | 调整 Monitor 初始化，移除旧 RegisterHTTPHandlers |
| `shop-backend/go.mod` | 添加 prometheus/client_golang 和 gin-contrib/pprof 依赖 |
| `docker/docker-compose.yml` | 添加 Prometheus 和 Grafana 服务 |
| `docker/prometheus/prometheus.yml` | Prometheus 配置 |
| `docker/grafana/provisioning/datasources/datasources.yml` | Grafana 数据源自动配置 |
| `docker/grafana/provisioning/dashboards/dashboards.yml` | Grafana Dashboard 自动配置 |
| `docker/grafana/provisioning/dashboards/go_runtime.json` | Go Runtime Dashboard 模板 |
| `frontend/src/views/Home.vue` | 添加系统监控侧边栏菜单项 |
| `frontend/.env.example` | 添加 VITE_GRAFANA_URL |

---

### 任务 1：添加 Go 依赖

**文件：**
- 修改：`shop-backend/go.mod`

- [ ] **步骤 1：添加 prometheus 和 pprof 依赖**

在 `shop-backend` 目录下执行：

```bash
cd d:\code\goshopadmin\shop-backend
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/prometheus/client_golang/prometheus/collectors
go get github.com/gin-contrib/pprof
go mod tidy
```

- [ ] **步骤 2：验证依赖安装**

```bash
cd d:\code\goshopadmin\shop-backend
go mod download
```

预期：无报错，依赖下载成功。

- [ ] **步骤 3：Commit**

```bash
git add shop-backend/go.mod shop-backend/go.sum
git commit -m "chore: add prometheus and pprof dependencies to shop-backend"
```

---

### 任务 2：创建 ModuleStats 堆栈解析模块

**文件：**
- 创建：`shop-backend/utils/monitor_module.go`

- [ ] **步骤 1：创建 monitor_module.go**

```go
package utils

import (
	"bytes"
	"runtime"
	"strings"
)

func collectModuleStats() ModuleStats {
	stats := make(ModuleStats)
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	stackData := buf[:n]

	sep := []byte("\n\n")
	goroutines := bytes.Split(stackData, sep)

	for _, goroutine := range goroutines {
		module := extractModule(goroutine)
		stats[module]++
	}

	return stats
}

func extractModule(goroutineStack []byte) string {
	lines := bytes.Split(goroutineStack, []byte("\n"))
	for i := 1; i < len(lines); i += 2 {
		line := strings.TrimSpace(string(lines[i]))
		module := parseModuleFromStackLine(line)
		if module != "" {
			return module
		}
	}
	return "other"
}

func parseModuleFromStackLine(line string) string {
	knownModules := []string{
		"shop-backend/services",
		"shop-backend/controllers",
		"shop-backend/mq",
		"shop-backend/utils",
		"shop-backend/middleware",
		"shop-backend/cache",
		"shop-backend/config",
		"shop-backend/routes",
		"shop-backend/pkg",
	}

	for _, mod := range knownModules {
		if strings.Contains(line, mod) {
			parts := strings.Split(mod, "/")
			return parts[len(parts)-1]
		}
	}

	if strings.Contains(line, "runtime.") {
		return "runtime"
	}

	return ""
}
```

- [ ] **步骤 2：验证编译**

```bash
cd d:\code\goshopadmin\shop-backend
go build ./utils/...
```

预期：编译成功，无报错。

- [ ] **步骤 3：Commit**

```bash
git add shop-backend/utils/monitor_module.go
git commit -m "feat: add ModuleStats goroutine stack parsing"
```

---

### 任务 3：创建 Prometheus 指标模块

**文件：**
- 创建：`shop-backend/utils/monitor_prometheus.go`

- [ ] **步骤 1：创建 monitor_prometheus.go**

```go
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

func updatePrometheusMetrics(stats *RuntimeStats) {
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

	mutexContentionsTotal.Add(float64(stats.MutexStats.Contentions))
	mutexDelaySecondsTotal.Add(stats.MutexStats.Delay / 1e9)
}
```

- [ ] **步骤 2：验证编译**

```bash
cd d:\code\goshopadmin\shop-backend
go build ./utils/...
```

预期：编译成功，无报错。

- [ ] **步骤 3：Commit**

```bash
git add shop-backend/utils/monitor_prometheus.go
git commit -m "feat: add Prometheus metrics definitions and update logic"
```

---

### 任务 4：重构 Monitor 核心模块

**文件：**
- 修改：`shop-backend/utils/monitor.go`

- [ ] **步骤 1：重写 monitor.go**

将现有 `monitor.go` 完整替换为以下内容：

```go
package utils

import (
	"runtime"
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
	stats           []RuntimeStats
	mu              sync.RWMutex
	alertThreshold  int
	memoryThreshold uint64
	checkInterval   time.Duration
	quit            chan struct{}
	maxHistorySize  int
	serviceName     string
	lastMutexStats  MutexStats
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
	runtime.ReadMemStats(memStats)

	moduleStats := collectModuleStats()

	mutexStats := m.collectMutexStats()

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
			ThreadCount:  runtime.NumCgoCall(),
			CgoCallCount: runtime.NumCgoCall(),
		},
		MutexStats:  mutexStats,
		Timestamp:   time.Now(),
		ServiceName: m.serviceName,
	}
}

func (m *Monitor) collectMutexStats() MutexStats {
	var mutexProfile runtime.MutexProfile
	runtime.SetMutexProfileFraction(1)
	mutexProfile.WriteTo(nil, 0)

	profile := runtime.MutexProfile{}
	records := profile.WriteTo(nil, 0)
	_ = records

	current := MutexStats{
		Contentions: 0,
		Delay:       0,
	}

	delta := MutexStats{
		Contentions: current.Contentions - m.lastMutexStats.Contentions,
		Delay:       current.Delay - m.lastMutexStats.Delay,
	}

	if delta.Contentions < 0 {
		delta.Contentions = 0
	}
	if delta.Delay < 0 {
		delta.Delay = 0
	}

	m.lastMutexStats = current

	return delta
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
```

- [ ] **步骤 2：验证编译**

```bash
cd d:\code\goshopadmin\shop-backend
go build ./...
```

预期：编译成功。注意 `RegisterHTTPHandlers` 方法已移除，`main.go` 中的调用会报错，这是预期的，将在任务 6 中修复。

- [ ] **步骤 3：Commit**

```bash
git add shop-backend/utils/monitor.go
git commit -m "feat: refactor Monitor with expanded runtime stats collection"
```

---

### 任务 5：创建 MonitorController

**文件：**
- 创建：`shop-backend/controllers/monitor_controller.go`

- [ ] **步骤 1：创建 monitor_controller.go**

```go
package controllers

import (
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
)

type MonitorController struct {
	monitor *utils.Monitor
}

func NewMonitorController(monitor *utils.Monitor) *MonitorController {
	return &MonitorController{monitor: monitor}
}

// GetCurrentStats 获取最新运行时统计
func (c *MonitorController) GetCurrentStats(ctx *gin.Context) {
	stats := c.monitor.GetCurrentStats()
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetHistoryStats 获取历史统计列表
func (c *MonitorController) GetHistoryStats(ctx *gin.Context) {
	history := c.monitor.GetHistoryStats()
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    history,
	})
}
```

- [ ] **步骤 2：验证编译**

```bash
cd d:\code\goshopadmin\shop-backend
go build ./controllers/...
```

预期：编译成功。

- [ ] **步骤 3：Commit**

```bash
git add shop-backend/controllers/monitor_controller.go
git commit -m "feat: add MonitorController for monitor API endpoints"
```

---

### 任务 6：注册监控和 pprof 路由到 Gin

**文件：**
- 修改：`shop-backend/routes/routes.go`

- [ ] **步骤 1：修改 routes.go — 添加 MonitorController 到 Dependencies**

在 `Dependencies` 结构体中添加 `MonitorController` 字段：

```go
type Dependencies struct {
	AuthController          *controllers.AuthController
	CustomerController      *controllers.CustomerController
	CaptchaController       *controllers.CaptchaController
	ProductController       *controllers.ProductController
	CartController          *controllers.CartController
	OrderController         *controllers.OrderController
	PaymentController       *controllers.PaymentController
	AddressController       *controllers.AddressController
	SpecificationController *controllers.SpecificationController
	ActivityController      *controllers.ActivityController
	RedeemCodeController    *controllers.RedeemCodeController
	ActivityOrderController *controllers.ActivityOrderController
	HealthController        *controllers.HealthController
	MonitorController       *controllers.MonitorController
}
```

- [ ] **步骤 2：修改 routes.go — 修改 SetupRoutes 函数签名**

将 `SetupRoutes` 函数签名改为接收 `*utils.Monitor` 参数：

```go
func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cfg *config.Config, monitor *utils.Monitor) {
```

在 `deps` 初始化中添加 MonitorController：

```go
deps := &Dependencies{
	// ... 现有控制器 ...
	MonitorController: controllers.NewMonitorController(monitor),
}
```

- [ ] **步骤 3：修改 routes.go — 添加监控和 pprof 路由**

在 `SetupRoutes` 函数末尾（`api` 路由组内）添加监控路由：

```go
monitor := api.Group("/monitor")
monitor.Use(middleware.Auth())
{
	monitor.GET("/stats", deps.MonitorController.GetCurrentStats)
	monitor.GET("/stats/history", deps.MonitorController.GetHistoryStats)
}
```

在 `SetupRoutes` 函数末尾（`r` 根级别）添加 Prometheus metrics 和 pprof 路由：

```go
r.GET("/metrics", gin.WrapH(promhttp.Handler()))

pprofGroup := r.Group("/debug/pprof")
pprofGroup.Use(middleware.Auth())
pprof.Register(pprofGroup)
```

需要在文件顶部添加 import：

```go
import (
	"context"

	"shop-backend/cache"
	"shop-backend/config"
	"shop-backend/controllers"
	"shop-backend/middleware"
	"shop-backend/utils"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)
```

- [ ] **步骤 4：验证编译**

```bash
cd d:\code\goshopadmin\shop-backend
go build ./...
```

预期：`main.go` 中 `routes.SetupRoutes` 调用参数不匹配会报错，将在任务 7 中修复。

- [ ] **步骤 5：Commit**

```bash
git add shop-backend/routes/routes.go
git commit -m "feat: register monitor and pprof routes with Gin"
```

---

### 任务 7：更新 main.go

**文件：**
- 修改：`shop-backend/main.go`

- [ ] **步骤 1：修改 main.go**

需要做以下修改：

1. 移除 `monitor.RegisterHTTPHandlers()` 调用
2. 将 `monitor` 传递给 `routes.SetupRoutes`

将 main.go 中的监控初始化部分：

```go
monitor := utils.NewMonitor(1000, 10*time.Second, 100)
monitor.Start()
defer monitor.Stop()
monitor.RegisterHTTPHandlers()
utils.Info("协程监控初始化成功")
```

改为：

```go
monitor := utils.NewMonitor(1000, 10*time.Second, 100)
monitor.Start()
defer monitor.Stop()
utils.Info("协程监控初始化成功")
```

将路由设置调用：

```go
routes.SetupRoutes(r, conn.DB, conn.Redis, cfg)
```

改为：

```go
routes.SetupRoutes(r, conn.DB, conn.Redis, cfg, monitor)
```

- [ ] **步骤 2：验证编译**

```bash
cd d:\code\goshopadmin\shop-backend
go build ./...
```

预期：编译成功，无报错。

- [ ] **步骤 3：Commit**

```bash
git add shop-backend/main.go
git commit -m "feat: update main.go to pass monitor to routes"
```

---

### 任务 8：添加 Docker Compose 服务

**文件：**
- 修改：`docker/docker-compose.yml`
- 创建：`docker/prometheus/prometheus.yml`

- [ ] **步骤 1：创建 Prometheus 配置文件**

创建目录 `docker/prometheus/` 并创建 `prometheus.yml`：

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'shop-backend'
    static_configs:
      - targets: ['goshopadmin-shop-backend:8081']
```

- [ ] **步骤 2：修改 docker-compose.yml — 添加 Prometheus 和 Grafana 服务**

在 `shop-backend` 服务之后、`nginx` 服务之前添加：

```yaml
  # Prometheus监控服务
  prometheus:
    image: prom/prometheus:latest
    container_name: goshopadmin-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    networks:
      - goshopadmin-network
    restart: always
    depends_on:
      - shop-backend
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:9090/-/healthy"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Grafana可视化服务
  grafana:
    image: grafana/grafana:latest
    container_name: goshopadmin-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    networks:
      - goshopadmin-network
    restart: always
    depends_on:
      - prometheus
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

注意：frontend 服务当前使用端口 3000，需要将 frontend 端口改为 5173（Vite 默认开发端口），或改为其他端口。将 frontend 的端口映射改为：

```yaml
  frontend:
    ports:
      - "5173:3000"  # 改为 5173 避免与 Grafana 冲突
```

在 `volumes` 部分添加：

```yaml
  prometheus-data:
    driver: local
    driver_opts:
      type: none
      device: ./data/prometheus
      o: bind
  grafana-data:
    driver: local
    driver_opts:
      type: none
      device: ./data/grafana
      o: bind
```

- [ ] **步骤 3：创建数据目录**

```bash
mkdir -p d:\code\goshopadmin\docker\data\prometheus
mkdir -p d:\code\goshopadmin\docker\data\grafana
```

- [ ] **步骤 4：Commit**

```bash
git add docker/docker-compose.yml docker/prometheus/prometheus.yml
git commit -m "feat: add Prometheus and Grafana to Docker Compose"
```

---

### 任务 9：添加 Grafana 自动配置

**文件：**
- 创建：`docker/grafana/provisioning/datasources/datasources.yml`
- 创建：`docker/grafana/provisioning/dashboards/dashboards.yml`
- 创建：`docker/grafana/provisioning/dashboards/go_runtime.json`

- [ ] **步骤 1：创建 Grafana 数据源配置**

创建目录 `docker/grafana/provisioning/datasources/` 并创建 `datasources.yml`：

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://goshopadmin-prometheus:9090
    isDefault: true
    editable: true
```

- [ ] **步骤 2：创建 Grafana Dashboard 配置**

创建目录 `docker/grafana/provisioning/dashboards/` 并创建 `dashboards.yml`：

```yaml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /etc/grafana/provisioning/dashboards
      foldersFromFilesStructure: false
```

- [ ] **步骤 3：创建 Go Runtime Dashboard JSON**

创建 `docker/grafana/provisioning/dashboards/go_runtime.json`，使用 Grafana 官方 Go Runtime Dashboard（Dashboard ID: 14058）的精简版本，包含以下面板：

- 协程数量趋势（`shop_goroutine_count`）
- 模块协程分布（`shop_goroutine_module_count`）
- 堆内存使用趋势（`shop_memory_heap_alloc_bytes`）
- 系统内存趋势（`shop_memory_sys_bytes`）
- GC 次数（`shop_gc_count_total`）
- 线程数趋势（`shop_thread_count`）
- 锁竞争趋势（`shop_mutex_contentions_total`）

Dashboard JSON 内容较长，使用 Grafana 官方模板 ID 14058 作为基础，替换指标名为自定义指标。具体 JSON 文件在实现时通过 Grafana API 导出或手动创建。

- [ ] **步骤 4：Commit**

```bash
git add docker/grafana/
git commit -m "feat: add Grafana provisioning configs and Go Runtime Dashboard"
```

---

### 任务 10：添加前端侧边栏监控链接

**文件：**
- 修改：`frontend/src/views/Home.vue`
- 修改：`frontend/.env.example`

- [ ] **步骤 1：修改 Home.vue — 添加侧边栏菜单项**

在 `Home.vue` 的 `<el-menu>` 中，活动管理菜单项之后添加：

```vue
<el-menu-item index="monitor">
  <el-icon><monitor /></el-icon>
  <span>系统监控</span>
</el-menu-item>
```

在 `<script setup>` 的 import 中添加 Monitor 图标：

```javascript
import { ArrowDown, House, User, Position, Lock, Shop, Goods, Grid, Calendar, Monitor } from '@element-plus/icons-vue';
```

在 `handleMenuSelect` 函数中添加 monitor 处理：

```javascript
const handleMenuSelect = (key) => {
  if (key === 'monitor') {
    const grafanaUrl = import.meta.env.VITE_GRAFANA_URL || 'http://localhost:3000'
    window.open(grafanaUrl, '_blank')
    return
  }
  activeMenu.value = key;
  currentView.value = key;
  currentProduct.value = null;
};
```

- [ ] **步骤 2：修改 .env.example**

添加 Grafana URL 环境变量：

```env
# 开发环境
VITE_API_BASE_URL=/api
VITE_API_PROXY_TARGET=http://backend:8080

# Grafana监控地址
VITE_GRAFANA_URL=http://localhost:3000

# 生产环境（根据实际情况修改）
# VITE_API_BASE_URL=/api
```

- [ ] **步骤 3：验证前端编译**

```bash
cd d:\code\goshopadmin\frontend
npm run build
```

预期：编译成功，无报错。

- [ ] **步骤 4：Commit**

```bash
git add frontend/src/views/Home.vue frontend/.env.example
git commit -m "feat: add system monitor sidebar link to Grafana"
```

---

### 任务 11：端到端验证

- [ ] **步骤 1：重启 Docker 容器**

```bash
cd d:\code\goshopadmin\docker
docker restart goshopadmin-shop-backend
docker restart goshopadmin-frontend
```

启动新服务：

```bash
cd d:\code\goshopadmin\docker
docker compose up -d prometheus grafana
```

- [ ] **步骤 2：验证 Prometheus 采集**

```bash
curl http://localhost:8081/metrics
```

预期：返回 Prometheus 格式的指标数据，包含 `shop_goroutine_count`、`shop_memory_alloc_bytes` 等自定义指标。

```bash
curl http://localhost:9090/api/v1/targets
```

预期：shop-backend target 状态为 UP。

- [ ] **步骤 3：验证 Grafana**

浏览器访问 `http://localhost:3000`，使用 admin/admin 登录，确认 Prometheus 数据源已配置，Dashboard 已加载。

- [ ] **步骤 4：验证前端链接**

浏览器访问管理后台，确认侧边栏出现"系统监控"菜单项，点击后在新标签页打开 Grafana。

- [ ] **步骤 5：验证 pprof 端点**

```bash
curl http://localhost:8081/debug/pprof/
```

预期：返回 401 未认证（因为需要 Auth 中间件）。

使用有效 token 访问：

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8081/debug/pprof/heap
```

预期：返回 heap profile 数据。

- [ ] **步骤 6：最终 Commit**

```bash
git add -A
git commit -m "feat: complete system monitoring enhancement with Prometheus + Grafana"
```
