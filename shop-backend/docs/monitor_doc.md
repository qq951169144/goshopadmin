# 运行时监控器技术文档

> 本文档对项目中 monitor/pprof 监控体系的代码流程、结构、统计内容进行完整总结，涵盖 shop-backend 完整版和 backend 简化版两套实现。

---

## 概述

项目中有 **两套** Monitor 实现，分属不同子项目：

| 子项目 | Monitor 文件 | 复杂度 | 特点 |
|:---|:---|:---|:---|
| **shop-backend** (C端商城) | `utils/monitor.go` + `monitor_prometheus.go` + `monitor_module.go` + `controllers/monitor_controller.go` | **完整版** | 5类指标采集 + Prometheus 推送 + pprof 端点 + 告警 |
| **backend** (后台管理) | `utils/monitor.go` | **简化版** | 仅协程统计，手动计数器，无 Prometheus |

`monitor.go` + `monitor_prometheus.go` + `monitor_module.go` 实现了一个完整的 Go 运行时监控器，用于实时采集和监控应用程序的协程、内存、线程、锁竞争等运行时指标，并通过 Prometheus + Grafana 进行可视化展示。

### 监控架构

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          监控架构总览                                     │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  shop-backend                                                            │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │  Monitor 采集层                                                     │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ │  │
│  │  │ 协程统计  │ │ 内存统计  │ │ 线程统计  │ │ 锁竞争   │ │ 模块统计 │ │  │
│  │  │ runtime.  │ │ runtime.  │ │ pprof.    │ │ pprof.    │ │ runtime. │ │  │
│  │  │ NumGorou- │ │ ReadMem-  │ │ Lookup    │ │ Lookup    │ │ Stack()  │ │  │
│  │  │ tine()    │ │ Stats()   │ │("thread") │ │("mutex")  │ │ 解析堆栈 │ │  │
│  │  └─────┬────┘ └─────┬────┘ └─────┬────┘ └─────┬────┘ └────┬────┘ │  │
│  │        └─────────────┴────────────┴────────────┴────────────┘      │  │
│  │                              │                                      │  │
│  │                    RuntimeStats 快照                                │  │
│  │                              │                                      │  │
│  │              ┌───────────────┼───────────────┐                      │  │
│  │              ↓               ↓               ↓                      │  │
│  │      历史记录存储     Prometheus 推送     告警检查                     │  │
│  │      (环形缓冲区)     (Counter/Gauge)    (日志输出)                   │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                              │                                           │
│                    /metrics 端点 (IP 白名单)                              │
│                              │                                           │
├──────────────────────────────┼───────────────────────────────────────────┤
│  Docker 内部网络              │                                           │
│                              ↓                                           │
│  ┌─────────────────┐   ┌─────────────┐   ┌─────────────────┐            │
│  │  Prometheus      │←──│  scrape     │   │  Grafana         │            │
│  │  :9090           │   │  /metrics   │   │  :3000           │            │
│  │  15s 采集间隔     │   │             │   │  admin/admin     │            │
│  └────────┬────────┘   └─────────────┘   └────────┬────────┘            │
│           │                                        │                     │
│           └────────── PromQL 查询 ─────────────────┘                     │
│                                                                          │
│  backend 前端                                                            │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │  侧边栏 "系统监控" → 新窗口打开 Grafana                              │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 一、核心数据结构

### 1.1 RuntimeStats - 运行时统计快照

```go
type RuntimeStats struct {
    GoroutineCount int           `json:"goroutine_count"`
    ModuleStats    ModuleStats   `json:"module_stats"`
    MemoryStats    MemoryStats   `json:"memory_stats"`
    ThreadStats    ThreadStats   `json:"thread_stats"`
    MutexStats     MutexStats    `json:"mutex_stats"`
    Timestamp      time.Time     `json:"timestamp"`
    ServiceName    string        `json:"service_name"`
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `GoroutineCount` | `int` | 当前系统中运行的协程总数，来源于 `runtime.NumGoroutine()` |
| `ModuleStats` | `ModuleStats` | 按模块分组的协程统计，通过解析协程堆栈获得 |
| `MemoryStats` | `MemoryStats` | 内存使用统计，来源于 `runtime.ReadMemStats()` |
| `ThreadStats` | `ThreadStats` | 线程与 CGO 调用统计 |
| `MutexStats` | `MutexStats` | 互斥锁竞争统计（增量值） |
| `Timestamp` | `time.Time` | 指标采集的时间戳 |
| `ServiceName` | `string` | 服务名称标识，默认 "shop-backend" |

### 1.2 ModuleStats - 模块协程统计

```go
type ModuleStats map[string]int
```

key 为模块名（如 `services`、`controllers`、`runtime`、`other`），value 为该模块的协程数量。通过 `runtime.Stack()` 获取所有协程堆栈，解析函数路径中的模块前缀来分类。

### 1.3 MemoryStats - 内存统计

```go
type MemoryStats struct {
    Alloc      uint64 `json:"alloc"`
    TotalAlloc uint64 `json:"total_alloc"`
    Sys        uint64 `json:"sys"`
    HeapAlloc  uint64 `json:"heap_alloc"`
    HeapSys    uint64 `json:"heap_sys"`
    StackInuse uint64 `json:"stack_inuse"`
    NumGC      uint32 `json:"num_gc"`
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `Alloc` | `uint64` | 当前已分配的堆对象字节数 |
| `TotalAlloc` | `uint64` | 历史累计分配的堆对象字节数 |
| `Sys` | `uint64` | 从操作系统获取的总内存字节数 |
| `HeapAlloc` | `uint64` | 已分配的堆对象字节数（与 Alloc 相同） |
| `HeapSys` | `uint64` | 从操作系统获取的堆内存字节数 |
| `StackInuse` | `uint64` | 栈区使用的字节数 |
| `NumGC` | `uint32` | 完成的 GC 循环次数（累积值） |

### 1.4 ThreadStats - 线程统计

```go
type ThreadStats struct {
    ThreadCount  int   `json:"thread_count"`
    CgoCallCount int64 `json:"cgo_call_count"`
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `ThreadCount` | `int` | 当前 OS 线程数量，来源于 `pprof.Lookup("threadcreate").Count()` |
| `CgoCallCount` | `int64` | CGO 调用累积次数，来源于 `runtime.NumCgoCall()` |

### 1.5 MutexStats - 互斥锁竞争统计

```go
type MutexStats struct {
    Contentions int64   `json:"contentions"`
    Delay       float64 `json:"delay"`
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `Contentions` | `int64` | 自上次采集以来的锁竞争增量次数 |
| `Delay` | `float64` | 自上次采集以来的锁等待延迟增量（纳秒） |

**注意**：MutexStats 中的值是增量值（delta），不是累积值。每次采集时计算 `当前累积值 - 上次累积值`。

### 1.6 Monitor - 监控器

```go
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
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `stats` | `[]RuntimeStats` | 历史指标数据，环形缓冲区策略 |
| `mu` | `sync.RWMutex` | 读写互斥锁，保证并发安全 |
| `alertThreshold` | `int` | 协程数量告警阈值，默认 1000 |
| `memoryThreshold` | `uint64` | 堆内存告警阈值，默认 512MB |
| `checkInterval` | `time.Duration` | 采集间隔，默认 10 秒 |
| `quit` | `chan struct{}` | 退出信号通道 |
| `maxHistorySize` | `int` | 历史记录最大条数，默认 100 |
| `serviceName` | `string` | 服务名称标识 |
| `stackBuf` | `[]byte` | 预分配 1MB 缓冲区，用于 `runtime.Stack()` 复用 |
| `lastGCCount` | `uint32` | 上次采集时的 GC 累积次数，用于计算增量 |
| `lastCgoCallCount` | `int64` | 上次采集时的 CGO 调用累积次数，用于计算增量 |
| `lastMutexContentions` | `int64` | 上次采集时的锁竞争累积次数 |
| `lastMutexDelay` | `float64` | 上次采集时的锁延迟累积值 |

---

## 二、文件结构说明

| 文件 | 职责 |
| :--- | :--- |
| `utils/monitor.go` | 核心监控逻辑：数据结构定义、采集循环、运行时指标采集、告警检查、历史数据管理 |
| `utils/monitor_prometheus.go` | Prometheus 指标定义与推送：11 个自定义指标的注册和更新 |
| `utils/monitor_module.go` | 模块协程统计：解析协程堆栈，按模块分组统计 |
| `controllers/monitor_controller.go` | HTTP 控制器：提供 API 接口查询监控数据 |
| `routes/routes.go` | 路由注册：监控 API、/metrics 端点、pprof 端点 |

---

## 三、监控运行流程

### 3.1 流程架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    监控器生命周期                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  初始化              启动             运行中            停止      │
│  NewMonitor()  →   Start()  →   collectMetrics →  Stop()       │
│       │                │              │                          │
│       ↓                ↓              ↓                          │
│  设置采样率        创建后台协程    定时采集指标                     │
│  注册Prometheus         │              │                          │
│  预分配缓冲区            │         ┌─────┴─────┐                  │
│                         │         ↓           ↓                  │
│                         │    collectOnce   检查退出信号            │
│                         │         │                               │
│                         │    ┌────┴────────────────┐             │
│                         │    ↓                     ↓             │
│                         │  采集运行时指标      计算增量值           │
│                         │    │                     │             │
│                         │    ↓                     ↓             │
│                         │  存储历史记录      推送 Prometheus       │
│                         │    │                     │             │
│                         │    ↓                     ↓             │
│                         │  检查告警条件      输出信息日志           │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 代码执行流程

```
main.go
  │
  ├── NewMonitor(1000, 10s, 100)     // 创建监控器实例
  │     ├── 设置 Mutex 采样率: SetMutexProfileFraction(5)  // 每5次竞争采样1次
  │     ├── 设置 Block 采样率: SetBlockProfileRate(100)    // 阻塞>=100ns才记录
  │     ├── 注册 Prometheus 指标 (11个)
  │     └── 预分配 1MB stackBuf 缓冲区
  │
  ├── monitor.Start()                // 启动后台采集协程
  │     └── go collectMetrics()      // 后台协程循环
  │           ├── <-ticker.C         // 每10秒触发
  │           │     └── collectOnce()
  │           │           ├── collectRuntimeStats()   // 采集5类指标
  │           │           │     ├── runtime.NumGoroutine()          → 协程数
  │           │           │     ├── runtime.ReadMemStats()          → 内存 (触发短暂STW)
  │           │           │     ├── collectModuleStats()            → 模块协程分布
  │           │           │     │     └── runtime.Stack(buf, true)  → 解析堆栈按模块分组
  │           │           │     ├── collectMutexStats()             → 锁竞争增量
  │           │           │     │     └── pprof.Lookup("mutex")     → 竞争次数+延迟
  │           │           │     └── getThreadCount()                → OS线程数
  │           │           │           └── pprof.Lookup("threadcreate").Count()
  │           │           ├── 存储历史记录 (环形缓冲区, max=100)
  │           │           ├── 计算 GC/CGO 增量 (当前累积 - 上次累积)
  │           │           ├── updatePrometheusMetrics()  → 推送到 Prometheus
  │           │           ├── checkAlerts()              → 告警检查
  │           │           │     ├── 协程数 > 1000 → Warn日志
  │           │           │     └── 堆内存 > 512MB → Warn日志
  │           │           └── Info日志 (协程数, 堆内存, 线程数)
  │           │
  │           └── <-quit             // 收到退出信号
  │                 └── return
  │
  └── routes.SetupRoutes(r, ..., monitor)
        ├── /api/monitor/stats         → GetCurrentStats()  (JWT认证)
        ├── /api/monitor/stats/history → GetHistoryStats()  (JWT认证)
        ├── /metrics                   → promhttp.Handler() (IP白名单)
        └── /debug/pprof/*             → gin-contrib/pprof  (JWT认证)
```

### 3.3 详细流程说明

#### 阶段一：监控器初始化

```go
func NewMonitor(alertThreshold int, checkInterval time.Duration, maxHistorySize int) *Monitor
```

**初始化参数说明：**

| 参数 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `alertThreshold` | `int` | 1000 | 协程数告警阈值 |
| `checkInterval` | `time.Duration` | 10s | 采集间隔 |
| `maxHistorySize` | `int` | 100 | 历史记录最大条数 |

**初始化逻辑：**
1. 参数校验：若传入参数小于等于 0，使用默认值
2. 设置 mutex/block 采样率（`SetMutexProfileFraction(5)`、`SetBlockProfileRate(100)`）
3. 注册 Prometheus 自定义指标
4. 预分配 1MB 栈缓冲区（`stackBuf`）
5. 创建 `Monitor` 实例

**采样率说明：**

| 采样设置 | 值 | 含义 |
| :--- | :--- | :--- |
| `SetMutexProfileFraction(5)` | 5 | 每 5 次 mutex 竞争事件采样 1 次 |
| `SetBlockProfileRate(100)` | 100ns | 阻塞时间 >= 100ns 才记录 |

> **为什么不使用全量采样（值=1）？** 全量采样在高并发场景下会产生显著性能开销，采样率为 5/100ns 可以在性能和精度之间取得平衡。

#### 阶段二：启动监控

```go
func (m *Monitor) Start() {
    go m.collectMetrics()
}
```

在新的后台协程中运行 `collectMetrics()`，主线程不会阻塞。

#### 阶段三：定时采集指标

```go
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
```

| 分支 | 触发条件 | 处理逻辑 |
| :--- | :--- | :--- |
| `ticker.C` | 到达采集间隔 | 调用 `collectOnce()` 采集一次指标 |
| `quit` | 收到退出信号 | 退出循环，停止采集 |

#### 阶段四：单次指标采集

```go
func (m *Monitor) collectOnce()
```

**采集步骤：**

1. **采集运行时指标**：调用 `collectRuntimeStats()` 获取完整快照
2. **存储历史记录**：加写锁，追加到切片，超过最大容量时淘汰最早记录
3. **计算 GC 增量**：`gcDelta = stats.MemoryStats.NumGC - m.lastGCCount`
4. **计算 CGO 增量**：`cgoDelta = stats.ThreadStats.CgoCallCount - m.lastCgoCallCount`
5. **推送 Prometheus**：调用 `updatePrometheusMetrics()` 更新所有指标
6. **检查告警条件**：协程数超阈值或堆内存超阈值时输出警告日志
7. **输出信息日志**：记录当前协程数、堆内存、线程数

**增量计算原理：**

Prometheus 的 Counter 类型只能单调递增，不能直接 Set 绝对值。对于 GC 次数、CGO 调用次数这类累积指标，需要计算增量后通过 `Counter.Add(delta)` 推送。

```
当前累积值: 150
上次累积值: 120
增量 delta: 30
Counter.Add(30)
```

Monitor 结构体中保存了 `lastGCCount`、`lastCgoCallCount`、`lastMutexContentions`、`lastMutexDelay` 四个字段用于增量计算。

#### 阶段五：运行时指标采集

```go
func (m *Monitor) collectRuntimeStats() RuntimeStats
```

**采集数据源：**

| 数据项 | 来源 API | 说明 |
| :--- | :--- | :--- |
| 协程总数 | `runtime.NumGoroutine()` | 所有活跃协程 |
| 内存统计 | `runtime.ReadMemStats()` | 触发短暂 STW（微秒级） |
| 模块统计 | `runtime.Stack()` + 堆栈解析 | 解析协程堆栈按模块分组 |
| 线程数量 | `pprof.Lookup("threadcreate").Count()` | OS 线程数 |
| CGO 调用 | `runtime.NumCgoCall()` | CGO 调用累积次数 |
| 锁竞争 | `pprof.Lookup("mutex")` | 竞争次数 + 延迟 |

#### 阶段六：停止监控

```go
func (m *Monitor) Stop() {
    close(m.quit)
}
```

关闭 `quit` 通道，后台协程收到信号后安全退出。

---

## 四、统计内容汇总

| 指标类别 | 统计项 | 数据来源 | Prometheus 指标名 | 类型 |
|:---|:---|:---|:---|:---|
| **协程** | 协程总数 | `runtime.NumGoroutine()` | `shop_goroutine_count` | Gauge |
| **协程** | 按模块协程数 | `runtime.Stack()` + 堆栈解析 | `shop_goroutine_module_count{module}` | GaugeVec |
| **内存** | 已分配堆对象字节 | `runtime.ReadMemStats().Alloc` | `shop_memory_alloc_bytes` | Gauge |
| **内存** | 从OS获取总内存 | `runtime.ReadMemStats().Sys` | `shop_memory_sys_bytes` | Gauge |
| **内存** | 堆分配字节 | `runtime.ReadMemStats().HeapAlloc` | `shop_memory_heap_alloc_bytes` | Gauge |
| **内存** | 栈区使用字节 | `runtime.ReadMemStats().StackInuse` | `shop_memory_stack_inuse_bytes` | Gauge |
| **内存** | GC次数(增量) | `runtime.ReadMemStats().NumGC` | `shop_gc_count_total` | Counter |
| **线程** | OS线程数 | `pprof.Lookup("threadcreate")` | `shop_thread_count` | Gauge |
| **线程** | CGO调用次数(增量) | `runtime.NumCgoCall()` | `shop_cgo_call_count_total` | Counter |
| **锁竞争** | 竞争次数(增量) | `pprof.Lookup("mutex")` | `shop_mutex_contentions_total` | Counter |
| **锁竞争** | 等待延迟(增量,秒) | `pprof.Lookup("mutex")` 文本解析 | `shop_mutex_delay_seconds_total` | Counter |

---

## 五、模块协程统计原理

### 5.1 采集流程

```
runtime.Stack(buf, true)
        │
        ↓
获取所有协程堆栈文本
        │
        ↓
按 "\n\n" 分割为单个协程块
        │
        ↓
对每个协程块遍历所有行
        │
        ↓
匹配已知模块前缀
        │
        ↓
返回模块名或 "other"
```

### 5.2 堆栈格式示例

```
goroutine 42 [running]:
shop-backend/services.(*OrderService).CreateOrder(...)
        /app/services/order_service.go:45 +0x123
shop-backend/controllers.(*OrderController).CreateOrder(...)
        /app/controllers/order_controller.go:28 +0x456
```

### 5.3 模块匹配规则

| 堆栈中的路径 | 匹配到的模块 |
| :--- | :--- |
| `shop-backend/services/...` | `services` |
| `shop-backend/controllers/...` | `controllers` |
| `shop-backend/mq/...` | `mq` |
| `shop-backend/utils/...` | `utils` |
| `shop-backend/middleware/...` | `middleware` |
| `shop-backend/cache/...` | `cache` |
| `shop-backend/config/...` | `config` |
| `shop-backend/routes/...` | `routes` |
| `shop-backend/pkg/...` | `pkg` |
| `runtime.*` | `runtime` |
| 其他 | `other` |

### 5.4 缓冲区复用

`stackBuf` 是预分配的 1MB 缓冲区，存储在 `Monitor` 结构体中。每次调用 `runtime.Stack()` 复用此缓冲区，避免每 10 秒分配 1MB 内存造成 GC 压力。

---

## 六、Prometheus 指标说明

### 6.1 自定义指标列表

| 指标名 | 类型 | 标签 | 说明 |
| :--- | :--- | :--- | :--- |
| `shop_goroutine_count` | Gauge | - | 当前协程总数 |
| `shop_goroutine_module_count` | GaugeVec | `module` | 按模块分组的协程数 |
| `shop_memory_alloc_bytes` | Gauge | - | 已分配的堆对象字节数 |
| `shop_memory_sys_bytes` | Gauge | - | 从 OS 获取的总内存字节数 |
| `shop_memory_heap_alloc_bytes` | Gauge | - | 堆分配字节数 |
| `shop_memory_stack_inuse_bytes` | Gauge | - | 栈区使用字节数 |
| `shop_gc_count_total` | Counter | - | GC 总次数（增量推送） |
| `shop_thread_count` | Gauge | - | 当前 OS 线程数 |
| `shop_cgo_call_count_total` | Counter | - | CGO 调用总次数（增量推送） |
| `shop_mutex_contentions_total` | Counter | - | 互斥锁竞争总次数（增量推送） |
| `shop_mutex_delay_seconds_total` | Counter | - | 互斥锁等待延迟总秒数（增量推送） |

### 6.2 Gauge vs Counter 使用规则

| 类型 | 语义 | 操作 | 适用场景 |
| :--- | :--- | :--- | :--- |
| **Gauge** | 当前值，可增可减 | `Set(value)` | 协程数、内存、线程数 |
| **Counter** | 累积值，只能递增 | `Add(delta)` | GC 次数、CGO 调用、锁竞争 |

### 6.3 GaugeVec 标签管理

`shop_goroutine_module_count` 使用 `Reset()` 策略：每次更新前先清除所有标签，再重新设置当前存在的模块。这避免了模块协程数降为 0 后标签仍然残留在 Prometheus 中的问题。

### 6.4 增量计算逻辑

对于 Counter 类型指标，需要计算增量后推送：

```
增量 = 当前累积值 - 上次记录的累积值
如果增量 < 0，则增量 = 0（防止进程重启后累积值重置导致负增量）
Counter.Add(增量)
```

---

## 七、HTTP 接口

### 7.1 监控 API（需要认证）

| 接口路径 | HTTP 方法 | 功能描述 | 认证 |
| :--- | :--- | :--- | :--- |
| `/api/monitor/stats` | `GET` | 获取最新运行时统计快照 | 需要 |
| `/api/monitor/stats/history` | `GET` | 获取历史统计列表 | 需要 |

### 7.2 Prometheus Metrics 端点

| 接口路径 | HTTP 方法 | 功能描述 | 认证 |
| :--- | :--- | :--- | :--- |
| `/metrics` | `GET` | Prometheus 指标抓取端点 | IP 白名单 |

**IP 白名单策略：** 仅允许以下网段访问 `/metrics`：
- `172.16.0.0/12` — Docker 默认网段
- `192.168.0.0/16` — 局域网
- `127.0.0.1` — 本机回环

### 7.3 pprof 性能分析端点（需要认证）

pprof 端点仅在 **shop-backend** 中注册，通过 `gin-contrib/pprof` 库提供标准 Go 性能分析能力：

| 接口路径 | HTTP 方法 | 功能描述 | 认证 | 用途 |
| :--- | :--- | :--- | :--- | :--- |
| `/debug/pprof/` | `GET` | pprof 概览页 | 需要 | 查看所有 profile 链接 |
| `/debug/pprof/profile` | `GET` | CPU profile | 需要 | 分析 CPU 热点函数（默认30秒采样） |
| `/debug/pprof/heap` | `GET` | 堆内存 profile | 需要 | 分析内存分配和泄漏 |
| `/debug/pprof/goroutine` | `GET` | 协程 profile | 需要 | 分析协程泄漏和阻塞 |
| `/debug/pprof/mutex` | `GET` | 互斥锁 profile | 需要 | 分析锁竞争热点 |
| `/debug/pprof/block` | `GET` | 阻塞 profile | 需要 | 分析阻塞操作（采样率100ns） |
| `/debug/pprof/threadcreate` | `GET` | 线程创建 profile | 需要 | 分析线程创建情况 |

pprof 数据同时被 Monitor 内部使用：`pprof.Lookup("threadcreate").Count()` 获取线程数，`pprof.Lookup("mutex")` 获取锁竞争数据。

### 7.4 接口响应格式

#### `/api/monitor/stats` 响应

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "goroutine_count": 42,
        "module_stats": {
            "services": 12,
            "controllers": 5,
            "runtime": 15,
            "other": 10
        },
        "memory_stats": {
            "alloc": 10485760,
            "total_alloc": 52428800,
            "sys": 67108864,
            "heap_alloc": 8388608,
            "heap_sys": 33554432,
            "stack_inuse": 1048576,
            "num_gc": 15
        },
        "thread_stats": {
            "thread_count": 8,
            "cgo_call_count": 0
        },
        "mutex_stats": {
            "contentions": 3,
            "delay": 150000
        },
        "timestamp": "2026-05-28T10:30:00Z",
        "service_name": "shop-backend"
    }
}
```

---

## 八、告警机制

### 8.1 告警触发条件

| 条件 | 阈值 | 告警级别 | 输出格式 |
| :--- | :--- | :--- | :--- |
| 协程数超阈值 | 默认 1000 | WARN | `协程数量超过阈值: 当前=X, 阈值=Y` |
| 堆内存超阈值 | 默认 512MB | WARN | `堆内存使用超过阈值: 当前=X MB, 阈值=Y MB` |

### 8.2 告警方式

当前仅输出 WARN 级别日志。可通过 Grafana 的 Alert Rules 配置更丰富的告警通知（邮件、Webhook 等）。

---

## 九、backend 简化版监控体系

### 9.1 数据结构

```
GoroutineMetrics (协程监控指标)
├── Total        int            // 总协程数 (runtime.NumGoroutine)
├── System       int            // 系统协程数 (硬编码估算=4)
├── Application  int            // 应用协程数 (Total - System)
├── ByModule     map[string]int // 按模块统计 (手动计数器)
└── Timestamp    time.Time      // 时间戳
```

### 9.2 代码流程

```
main.go
  ├── NewMonitor()              // 创建简化版监控器
  ├── monitor.Start(5s)         // 每5秒采集
  │     └── go updateMetrics()  // 后台协程
  │           ├── runtime.NumGoroutine() → Total
  │           ├── estimateSystemGoroutines() → System (固定=4)
  │           ├── Total - System → Application
  │           └── 读取 moduleCounters → ByModule
  │
  └── routes.SetupRoutes(r, ..., monitor)
        └── /api/monitor/goroutines         → 协程指标 (认证+权限)
        └── /api/monitor/goroutines/total   → 协程总数
        └── /api/monitor/goroutines/module  → 模块分布
```

### 9.3 与 shop-backend 的关键差异

| 对比项 | shop-backend | backend |
|:---|:---|:---|
| 指标种类 | 5类(协程+内存+线程+锁+模块) | 1类(仅协程) |
| 模块统计方式 | `runtime.Stack()` 自动解析堆栈 | 手动 `IncrementModuleCounter/DecrementModuleCounter` |
| Prometheus 集成 | 完整(11个自定义指标) | 无 |
| pprof 端点 | 有 (`/debug/pprof/*`) | 无 |
| 告警机制 | 有(协程数+堆内存) | 无 |
| 历史数据 | 环形缓冲区(100条) | 仅保留当前快照 |
| 增量计算 | GC/CGO/锁竞争增量 | 无 |
| IP 白名单 | `/metrics` 有 | 无 |

---

## 十、Docker 部署架构

### 10.1 服务依赖关系

```
                    ┌─────────────┐
                    │   Nginx     │ :80/:443
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ↓            ↓            ↓
      ┌──────────┐  ┌──────────┐  ┌──────────────┐
      │ frontend │  │ shop-    │  │ shop-        │
      │ :5173    │  │ frontend │  │ backend      │
      └──────────┘  │ :3001    │  │ :8081(内部)  │
                    └──────────┘  └──────┬───────┘
                                         │
                          ┌──────────────┼──────────────┐
                          ↓              ↓              ↓
                  ┌──────────┐   ┌──────────┐   ┌──────────┐
                  │ MySQL    │   │ Redis    │   │Prometheus│
                  │ :3306    │   │ :6379    │   │ :9090    │
                  └──────────┘   └──────────┘   └────┬─────┘
                                                      │
                                               ┌──────┴─────┐
                                               │  Grafana    │
                                               │  :3000      │
                                               │  admin/admin│
                                               └────────────┘
```

### 10.2 端口映射

| 服务 | 容器端口 | 宿主机映射 | 访问限制 |
| :--- | :--- | :--- | :--- |
| shop-backend | 8081 | `127.0.0.1:8081:8081` | 仅本机可访问 |
| Prometheus | 9090 | `9090:9090` | Docker 内部网络 |
| Grafana | 3000 | `3000:3000` | 需要登录（admin/admin） |

### 10.3 Grafana 配置

| 配置项 | 值 | 说明 |
| :--- | :--- | :--- |
| 匿名访问 | 关闭 | 需要登录才能查看监控面板 |
| 管理员账号 | admin / admin | 首次登录后建议修改密码 |
| 数据源 | Prometheus (uid: prometheus) | 自动配置 |
| Dashboard | Go Runtime Monitor | 自动加载 |

### 10.4 Grafana Dashboard 面板

| 面板 | 指标 | 图表类型 | 说明 |
| :--- | :--- | :--- | :--- |
| Goroutine Count | `shop_goroutine_count` | 时序图 | 协程数量趋势 |
| Goroutine Distribution | `shop_goroutine_module_count` | 饼图 | 协程模块分布 |
| Heap Memory Usage | `shop_memory_heap_alloc_bytes`, `shop_memory_sys_bytes` | 时序图 | 堆内存使用趋势 |
| GC Rate | `rate(shop_gc_count_total[5m])` | 时序图 | GC 频率（次/秒） |
| Thread Count | `shop_thread_count` | 时序图 | 线程数量趋势 |
| Mutex Contention | `rate(shop_mutex_contentions_total[5m])`, `rate(shop_mutex_delay_seconds_total[5m])` | 时序图 | 锁竞争频率和延迟 |

---

## 十一、使用示例

### 11.1 创建并启动监控器

```go
// 创建监控器：阈值2000，每5秒采集，保留200条历史
monitor := NewMonitor(2000, 5*time.Second, 200)

// 启动监控（在后台协程中运行）
monitor.Start()

// ... 业务逻辑 ...

// 停止监控（程序退出前）
monitor.Stop()
```

### 11.2 查询监控数据

```go
// 获取最新统计
stats := monitor.GetCurrentStats()
fmt.Printf("协程数: %d, 堆内存: %d bytes\n", stats.GoroutineCount, stats.MemoryStats.HeapAlloc)

// 获取历史记录
history := monitor.GetHistoryStats()
for _, s := range history {
    fmt.Printf("[%s] 协程: %d\n", s.Timestamp.Format("15:04:05"), s.GoroutineCount)
}
```

### 11.3 通过 API 查询

```bash
# 获取最新统计（需要认证）
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/monitor/stats

# 获取历史记录（需要认证）
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/monitor/stats/history

# Prometheus 抓取指标（Docker 内部网络访问）
curl http://goshopadmin-shop-backend:8081/metrics
```

---

## 十二、安全策略

### 12.1 端点安全

| 端点 | 安全策略 | 原因 |
| :--- | :--- | :--- |
| `/api/monitor/*` | JWT 认证 | 监控数据属于敏感信息 |
| `/metrics` | IP 白名单 | Prometheus 需要频繁访问，JWT 不适用 |
| `/debug/pprof/*` | JWT 认证 | pprof 暴露 CPU、内存等敏感运行时数据 |

### 12.2 网络隔离

- `shop-backend` 端口 8081 仅映射到 `127.0.0.1`，外部无法直接访问
- Prometheus 通过 Docker 内部网络 (`goshopadmin-network`) 访问 shop-backend
- Grafana 关闭匿名访问，需要登录后查看

---

## 十三、性能影响分析

### 13.1 采集开销

| 采集项 | 开销 | 说明 |
| :--- | :--- | :--- |
| `runtime.NumGoroutine()` | 极低 | 仅读取全局计数器 |
| `runtime.ReadMemStats()` | 低 | 触发短暂 STW（微秒级） |
| `runtime.Stack(buf, true)` | 中 | 遍历所有协程堆栈，1MB 缓冲区复用 |
| `pprof.Lookup("mutex")` | 低 | 读取已有的 profile 数据 |
| `pprof.Lookup("threadcreate")` | 极低 | 仅读取计数器 |

### 13.2 采样率影响

| 采样设置 | 值 | 性能影响 |
| :--- | :--- | :--- |
| `SetMutexProfileFraction(5)` | 1/5 采样 | 低开销 |
| `SetBlockProfileRate(100ns)` | >=100ns 记录 | 低开销 |

> 如果使用全量采样（值=1），高并发场景下 mutex/block 采样会产生显著性能开销。

---

## 十四、配置参数汇总

| 参数 | 默认值 | 配置方式 | 说明 |
| :--- | :--- | :--- | :--- |
| 协程告警阈值 | 1000 | `NewMonitor(alertThreshold, ...)` | 协程数超过此值输出警告 |
| 堆内存告警阈值 | 512MB | 代码中硬编码 | 堆内存超过此值输出警告 |
| 采集间隔 | 10s | `NewMonitor(..., checkInterval, ...)` | 每隔多久采集一次 |
| 历史记录条数 | 100 | `NewMonitor(..., maxHistorySize)` | 最大保留的历史记录数 |
| Mutex 采样率 | 5 | `runtime.SetMutexProfileFraction(5)` | 每 N 次竞争采样 1 次 |
| Block 采样阈值 | 100ns | `runtime.SetBlockProfileRate(100)` | 阻塞时间阈值 |
| Prometheus 采集间隔 | 15s | `prometheus.yml: scrape_interval` | Prometheus 抓取间隔 |
| Grafana 刷新间隔 | 10s | Dashboard JSON: `refresh` | 面板自动刷新间隔 |
| stackBuf 缓冲区 | 1MB | `make([]byte, 1<<20)` | 协程堆栈读取缓冲区，复用避免GC压力 |

---

*文档版本: 3.0*
*最后更新: 2026-06-11*
