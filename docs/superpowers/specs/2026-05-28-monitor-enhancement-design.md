# 系统监控增强设计规格

## 概述

增强 shop-backend 的运行时监控能力，通过 Prometheus + Grafana 实现可视化监控，在 backend 管理后台通过侧边栏链接跳转 Grafana 查看监控数据。

---

## 一、架构设计

### 1.1 整体架构

```
┌────────────────────┐    Scrape /metrics    ┌───────────┐
│  shop-backend      │◄─────────────────────│ Prometheus │
│  (C端商城 8081)     │    每 15 秒           │  :9090    │
│                    │                       └─────┬─────┘
│  /metrics          │                             │
│  /debug/pprof/*    │                       ┌─────▼─────┐
│                    │                       │  Grafana   │
└────────────────────┘                       │  :3000    │
                                             │           │
                                             │ Go Runtime│
                                             │ Dashboard │
                                             └─────┬─────┘
                                                   │
┌────────────────────┐                       ┌─────▼─────┐
│  backend           │                       │  frontend │
│  (管理后台 8080)     │◄────────────────────│  管理后台   │
│                    │   侧边栏链接跳转       │  前端      │
└────────────────────┘                       └───────────┘
```

### 1.2 数据流

| 数据类型 | 传输方式 | 方向 | 频率 |
|:---|:---|:---|:---|
| 运行时统计（协程/内存/线程/锁/模块） | Prometheus 拉取 `/metrics` | shop-backend → Prometheus | 每 15 秒 |
| CPU / Heap / Mutex / Goroutine Profile | 标准 pprof 端点 | Grafana / go tool pprof → shop-backend | 按需 |
| 可视化展示 | Grafana Dashboard | Prometheus → Grafana | 实时 |

### 1.3 管理后台集成方式

在 frontend 管理后台侧边栏添加"系统监控"菜单项，点击后在新标签页中打开 Grafana 页面。

---

## 二、shop-backend 改造

### 2.1 Monitor 增强

重构现有 `utils/monitor.go`，扩展采集范围：

**采集项**：

| 采集项 | 数据来源 | 采集方式 |
|:---|:---|:---|
| 协程数量 | `runtime.NumGoroutine()` | 定时采集 |
| 模块协程分布 | `runtime.Stack()` 解析堆栈 | 定时采集 |
| 内存统计 | `runtime.ReadMemStats()` | 定时采集 |
| 线程数量 | `runtime.NumCgoCall()` | 定时采集 |
| 锁竞争统计 | `runtime.MutexProfile` | 定时采集 |
| CPU Profile | `net/http/pprof` | 按需采集 |
| Heap Profile | `net/http/pprof` | 按需采集 |
| Mutex Profile | `net/http/pprof` | 按需采集 |
| Goroutine Profile | `net/http/pprof` | 按需采集 |

**数据结构**：

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
```

### 2.2 ModuleStats 实现方案

通过 `runtime.Stack()` 获取所有协程堆栈，解析包路径后按模块分组：

**分组规则**：

| 堆栈包路径前缀 | 归属模块 | 示例 |
|:---|:---|:---|
| `shop-backend/services` | services | 订单服务、商品服务等 |
| `shop-backend/controllers` | controllers | HTTP 请求处理 |
| `shop-backend/mq` | mq | 消息队列消费者 |
| `shop-backend/utils` | utils | 工具类 |
| `shop-backend/middleware` | middleware | 中间件 |
| `runtime` | runtime | Go 运行时（GC、调度器等） |
| 其他/未知 | other | 第三方库、标准库等 |

### 2.3 Prometheus 指标暴露

使用 `prometheus/client_golang` 暴露标准 `/metrics` 端点：

**自定义指标**：

| 指标名 | 类型 | 标签 | 说明 |
|:---|:---|:---|:---|
| `shop_goroutine_count` | Gauge | - | 当前协程总数 |
| `shop_goroutine_module_count` | Gauge | module | 分模块协程数 |
| `shop_memory_alloc_bytes` | Gauge | - | 已分配内存 |
| `shop_memory_sys_bytes` | Gauge | - | 系统分配内存 |
| `shop_memory_heap_alloc_bytes` | Gauge | - | 堆分配内存 |
| `shop_memory_stack_inuse_bytes` | Gauge | - | 栈使用内存 |
| `shop_gc_count_total` | Counter | - | GC 总次数 |
| `shop_thread_count` | Gauge | - | 线程数 |
| `shop_cgo_call_count` | Gauge | - | CGO 调用次数 |
| `shop_mutex_contentions_total` | Counter | - | 锁竞争总次数 |
| `shop_mutex_delay_seconds_total` | Counter | - | 锁等待总时间 |

同时启用 `prometheus/client_golang/prometheus/collectors` 中的默认 Go 收集器，提供标准 Go runtime 指标。

### 2.4 Gin 路由设计

将现有标准库 HTTP 接口迁移到 Gin Router：

**无需认证的路由**（Docker 内部网络访问）：

```
/metrics                → Prometheus 指标端点（仅 Docker 网络内可访问）
```

**需要认证的路由**（使用 shop-backend 的 Auth 中间件）：

```
/api/monitor
├── GET  /stats          → 获取最新运行时统计
├── GET  /stats/history  → 获取历史统计列表
└── (其他监控 API)

/debug/pprof             → 需要 Auth 中间件保护
├── GET  /               → pprof 概览页
├── GET  /profile        → CPU Profile（参数：seconds）
├── GET  /heap           → 堆内存 Profile
├── GET  /mutex          → 锁竞争 Profile
├── GET  /goroutine      → 协程 Profile
└── GET  /threadcreate   → 线程创建 Profile
```

**pprof 认证方案**：pprof 端点使用 shop-backend 现有的 JWT Auth 中间件保护。Grafana 访问 pprof 数据时，通过配置的监控服务 Token 进行认证。

### 2.5 告警机制

保持现有日志告警方式，超阈值输出 Warn 日志：

- 协程数超过阈值 → `Warn("协程数量超过阈值: 当前=%d, 阈值=%d")`
- 内存使用超过阈值 → `Warn("内存使用超过阈值: 当前=%d MB, 阈值=%d MB")`

Grafana 中可额外配置告警规则，通过 Grafana 告警通知渠道发送通知。

### 2.6 新增依赖

```
github.com/prometheus/client_golang
```

---

## 三、Docker Compose 新增服务

### 3.1 Prometheus

```yaml
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
```

**prometheus.yml 配置**：

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'shop-backend'
    static_configs:
      - targets: ['goshopadmin-shop-backend:8081']
```

### 3.2 Grafana

```yaml
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
```

### 3.3 Grafana 自动配置

通过 provisioning 配置自动添加 Prometheus 数据源和导入 Go Runtime Dashboard：

**datasources.yml**：

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://goshopadmin-prometheus:9090
    isDefault: true
```

**dashboards.yml**：

```yaml
apiVersion: 1
providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    options:
      path: /etc/grafana/provisioning/dashboards
      foldersFromFilesStructure: false
```

使用 Grafana 官方 Go Runtime Dashboard JSON（Dashboard ID: 14058 或类似模板）。

---

## 四、frontend 管理后台集成

### 4.1 侧边栏菜单

在 frontend 管理后台侧边栏添加"系统监控"菜单项，点击后在新标签页打开 Grafana：

```javascript
{
  path: '/monitor',
  name: 'SystemMonitor',
  meta: { title: '系统监控', icon: 'Monitor' },
  children: [
    {
      path: 'grafana',
      name: 'GrafanaMonitor',
      meta: { title: '运行时监控', icon: 'DataLine' },
    }
  ]
}
```

### 4.2 跳转实现

点击菜单项时，通过 `window.open` 在新标签页打开 Grafana 地址：

```javascript
const openGrafana = () => {
  const grafanaUrl = import.meta.env.VITE_GRAFANA_URL || 'http://localhost:3000'
  window.open(grafanaUrl, '_blank')
}
```

### 4.3 环境变量

在 `.env.example` 中添加：

```
VITE_GRAFANA_URL=http://localhost:3000
```

---

## 五、文件变更清单

### 5.1 shop-backend 变更

| 文件 | 变更类型 | 说明 |
|:---|:---|:---|
| `utils/monitor.go` | 重构 | 扩展采集范围，新增 ModuleStats/MemoryStats/ThreadStats/MutexStats，迁移到 Prometheus 指标 |
| `utils/monitor_prometheus.go` | 新增 | Prometheus 指标定义和采集逻辑 |
| `utils/monitor_module.go` | 新增 | ModuleStats 堆栈解析逻辑 |
| `routes/routes.go` | 修改 | 注册监控和 pprof 路由到 Gin |
| `controllers/monitor_controller.go` | 新增 | 监控 API 控制器 |
| `main.go` | 修改 | 调整 Monitor 初始化逻辑 |
| `go.mod` | 修改 | 添加 prometheus/client_golang 依赖 |

### 5.2 Docker 变更

| 文件 | 变更类型 | 说明 |
|:---|:---|:---|
| `docker/docker-compose.yml` | 修改 | 添加 Prometheus 和 Grafana 服务 |
| `docker/prometheus/prometheus.yml` | 新增 | Prometheus 配置文件 |
| `docker/grafana/provisioning/datasources/datasources.yml` | 新增 | Grafana 数据源自动配置 |
| `docker/grafana/provisioning/dashboards/dashboards.yml` | 新增 | Grafana Dashboard 自动配置 |
| `docker/grafana/provisioning/dashboards/go_runtime.json` | 新增 | Go Runtime Dashboard 模板 |

### 5.3 frontend 变更

| 文件 | 变更类型 | 说明 |
|:---|:---|:---|
| `src/router/index.js` | 修改 | 添加系统监控菜单路由 |
| `src/views/MonitorView.vue` | 新增 | 监控页面（跳转 Grafana） |
| `.env.example` | 修改 | 添加 VITE_GRAFANA_URL |

### 5.4 backend 变更

| 文件 | 变更类型 | 说明 |
|:---|:---|:---|
| `utils/monitor.go` | 保留 | backend 自身的协程监控保持不变 |
| `routes/routes.go` | 保留 | 现有 /api/monitor 路由保持不变 |

---

## 六、Grafana Dashboard 内容

### 6.1 运行时概览面板

| 面板 | 图表类型 | 指标 | 说明 |
|:---|:---|:---|:---|
| 协程数量 | 折线图 | `shop_goroutine_count` | 实时协程数趋势 |
| 模块协程分布 | 饼图 | `shop_goroutine_module_count` | 按模块分组统计 |
| 内存使用 | 面积图 | `shop_memory_alloc_bytes`, `shop_memory_sys_bytes`, `shop_memory_heap_alloc_bytes` | 内存使用趋势 |
| GC 次数 | 折线图 | `shop_gc_count_total` | GC 频率趋势 |
| 线程数 | 折线图 | `shop_thread_count` | 线程数趋势 |
| 锁竞争 | 折线图 | `shop_mutex_contentions_total`, `shop_mutex_delay_seconds_total` | 锁竞争趋势 |

### 6.2 pprof 分析面板

通过 Grafana 的 Explore 页面，使用 pprof 数据源查看：
- CPU Flame Graph
- Memory Flame Graph
- Goroutine 分析

---

*设计版本: 1.0*
*最后更新: 2026-05-28*
