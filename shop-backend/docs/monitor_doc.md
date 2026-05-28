
# 协程监控器技术文档

## 概述

`monitor.go` 实现了一个基于 Go 语言的协程（Goroutine）监控器，用于实时采集和监控应用程序中的协程数量，支持阈值告警和历史数据查询。

---

## 一、核心数据结构

### 1.1 GoroutineMetrics - 协程指标快照

```go
type GoroutineMetrics struct {
    TotalGoroutines int              `json:"total_goroutines"` // 当前总协程数
    ModuleStats     map[string]int   `json:"module_stats"`     // 分模块协程统计
    Timestamp       time.Time        `json:"timestamp"`        // 采集时间
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `TotalGoroutines` | `int` | 当前系统中运行的协程总数 |
| `ModuleStats` | `map[string]int` | 按模块分组的协程统计（预留字段） |
| `Timestamp` | `time.Time` | 指标采集的时间戳 |

### 1.2 Monitor - 协程监控器

```go
type Monitor struct {
    metrics        []GoroutineMetrics // 历史指标数据
    mu             sync.RWMutex       // 互斥锁
    alertThreshold int                // 告警阈值
    checkInterval  time.Duration      // 检查间隔
    quit           chan struct{}      // 退出信号
    maxHistorySize int                // 最大历史记录数
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `metrics` | `[]GoroutineMetrics` | 存储历史指标的切片，用于趋势分析 |
| `mu` | `sync.RWMutex` | 读写互斥锁，保证并发安全 |
| `alertThreshold` | `int` | 协程数量告警阈值，超过此值触发警告日志 |
| `checkInterval` | `time.Duration` | 定时采集的时间间隔 |
| `quit` | `chan struct{}` | 退出信号通道，用于优雅停止监控 |
| `maxHistorySize` | `int` | 历史记录的最大条数，超过后自动淘汰最早记录 |

---

## 二、监控运行流程

### 2.1 流程架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    监控器生命周期                            │
├─────────────────────────────────────────────────────────────┤
│  初始化          启动           运行中           停止       │
│  NewMonitor →   Start() →   collectMetrics →  Stop()       │
│                   │              │                          │
│                   ↓              ↓                          │
│             创建后台协程    定时采集指标                      │
│                   │              │                          │
│                   │         ┌─────┴─────┐                    │
│                   │         ↓           ↓                    │
│                   │    collectOnce    检查退出信号            │
│                   │         │                               │
│                   │    ┌────┴────┐                          │
│                   │    ↓         ↓                          │
│                   │  采集指标  检查阈值                      │
│                   │    │         │                          │
│                   │    ↓         ↓                          │
│                   │  存储历史  输出告警日志                    │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 详细流程说明

#### 阶段一：监控器初始化

```go
func NewMonitor(alertThreshold int, checkInterval time.Duration, maxHistorySize int) *Monitor
```

**初始化参数说明：**

| 参数 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `alertThreshold` | `int` | 1000 | 告警阈值，协程数超过此值触发警告 |
| `checkInterval` | `time.Duration` | 10s | 采集间隔，定时采集协程指标 |
| `maxHistorySize` | `int` | 100 | 历史记录最大条数 |

**初始化逻辑：**
1. 参数校验：若传入参数小于等于0，使用默认值
2. 创建 `Monitor` 实例，初始化切片容量为 `maxHistorySize`
3. 创建退出信号通道 `quit`

#### 阶段二：启动监控

```go
func (m *Monitor) Start() {
    go m.collectMetrics()
}
```

**启动逻辑：**
- 在新的后台协程中运行 `collectMetrics()` 方法
- 主线程不会阻塞，可以继续执行其他任务

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

**采集循环逻辑：**

| 分支 | 触发条件 | 处理逻辑 |
| :--- | :--- | :--- |
| `ticker.C` | 到达采集间隔 | 调用 `collectOnce()` 采集一次指标 |
| `quit` | 收到退出信号 | 退出循环，停止采集 |

#### 阶段四：单次指标采集

```go
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

    if metrics.TotalGoroutines > m.alertThreshold {
        Warn("协程数量超过阈值: 当前=%d, 阈值=%d", metrics.TotalGoroutines, m.alertThreshold)
    }

    Info("当前协程数量: %d", metrics.TotalGoroutines)
}
```

**采集步骤：**

1. **获取当前协程数**：调用 `runtime.NumGoroutine()` 获取系统当前运行的协程总数
2. **创建指标快照**：初始化 `ModuleStats` 为空 map（预留字段），记录当前时间戳
3. **存储历史数据**：加写锁，将新指标追加到切片末尾，超过最大容量时淘汰最早记录
4. **阈值检查**：若当前协程数超过告警阈值，输出警告日志
5. **常规日志**：输出当前协程数量信息日志

#### 阶段五：停止监控

```go
func (m *Monitor) Stop() {
    close(m.quit)
}
```

**停止逻辑：**
- 关闭 `quit` 通道，通知 `collectMetrics()` 退出循环
- 后台协程收到信号后自动退出

---

## 三、监控数据说明

### 3.1 采集的数据项

| 数据项 | 来源 | 类型 | 说明 |
| :--- | :--- | :--- | :--- |
| `TotalGoroutines` | `runtime.NumGoroutine()` | `int` | 当前进程中所有活跃协程的总数 |
| `ModuleStats` | 预留字段 | `map[string]int` | 计划按模块分组统计（尚未实现） |
| `Timestamp` | `time.Now()` | `time.Time` | 指标采集的精确时间 |

### 3.2 `runtime.NumGoroutine()` 返回值说明

该函数返回的协程总数包含：
- 用户创建的业务协程
- Go 运行时内部协程（如垃圾回收协程、调度器协程等）
- 当前正在执行的主协程

### 3.3 数据存储机制

```go
// 存储结构：环形缓冲区（逻辑上）
m.metrics = append(m.metrics, metrics)
if len(m.metrics) > m.maxHistorySize {
    m.metrics = m.metrics[1:]  // 移除最早的记录
}
```

**存储特点：**
- 使用切片模拟环形缓冲区
- 先进先出（FIFO）策略
- 固定最大容量，避免内存无限增长
- 读写操作通过互斥锁保证并发安全

---

## 四、HTTP 接口

### 4.1 接口注册

```go
func (m *Monitor) RegisterHTTPHandlers()
```

该方法注册两个 HTTP 接口到标准库的 `http.DefaultServeMux`。

### 4.2 接口列表

| 接口路径 | HTTP方法 | 功能描述 |
| :--- | :--- | :--- |
| `/metrics/goroutines` | `GET` | 获取当前协程指标快照 |
| `/metrics/goroutines/history` | `GET` | 获取历史指标数据列表 |

### 4.3 接口响应格式

#### `/metrics/goroutines` 响应

```json
{
    "total_goroutines": 42,
    "module_stats": {},
    "timestamp": "2024-01-15T10:30:00Z"
}
```

#### `/metrics/goroutines/history` 响应

```json
{
    "metrics": [
        {
            "total_goroutines": 38,
            "module_stats": {},
            "timestamp": "2024-01-15T10:29:50Z"
        },
        {
            "total_goroutines": 42,
            "module_stats": {},
            "timestamp": "2024-01-15T10:30:00Z"
        }
    ]
}
```

---

## 五、告警机制

### 5.1 告警触发条件

```go
if metrics.TotalGoroutines > m.alertThreshold {
    Warn("协程数量超过阈值: 当前=%d, 阈值=%d", metrics.TotalGoroutines, m.alertThreshold)
}
```

### 5.2 告警级别

当前实现仅输出 **WARN** 级别的日志，不包含：
- 邮件告警
- 短信告警
- 外部监控系统集成（如 Prometheus、Grafana）

---

## 六、使用示例

```go
// 创建监控器：阈值2000，每5秒采集，保留200条历史
monitor := NewMonitor(2000, 5*time.Second, 200)

// 启动监控
monitor.Start()

// 注册HTTP接口
monitor.RegisterHTTPHandlers()

// ... 业务逻辑 ...

// 停止监控（程序退出前）
monitor.Stop()
```

---

## 七、代码优化建议

### 7.1 当前局限性

| 问题 | 说明 | 影响 |
| :--- | :--- | :--- |
| `ModuleStats` 未实现 | 预留字段但未填充数据 | 无法按模块分析协程分布 |
| 告警方式单一 | 仅输出日志 | 无法及时通知运维人员 |
| 接口未使用 Gin | 使用标准库 HTTP | 与项目整体框架不一致 |
| 无指标重置功能 | 无法手动清除历史数据 | 测试和维护不便 |

### 7.2 优化建议

1. **实现模块级协程统计**：通过 runtime.Stack() 解析协程堆栈，按包路径分组统计

2. **扩展告警方式**：增加邮件、Webhook 等告警渠道

3. **接入 Gin 框架**：将接口注册到项目的 Gin router 中

4. **添加指标管理接口**：增加重置历史、动态调整阈值等功能

5. **接入 Prometheus**：暴露标准 Prometheus metrics 格式接口

---

## 八、总结

| 项目 | 说明 |
| :--- | :--- |
| **监控目标** | Go 应用程序协程数量 |
| **采集频率** | 可配置（默认10秒） |
| **数据保留** | 可配置（默认100条） |
| **告警方式** | WARN 级别日志 |
| **访问方式** | HTTP API |
| **并发安全** | 读写互斥锁保护 |

---

*文档版本: 1.0*
*最后更新: 2026-05-28*
