# GoShopAdmin 待修 Bug 与优化列表

> 最后更新：2026-06-10
> 状态标记：🔴 待处理 | 🟡 进行中 | 🟢 已完成 | ⏸️ 暂缓

---

## 一、架构与基础设施

### 1.1 MQ 连接协程优化

| 属性 | 值 |
|:---|:---|
| 状态 | 🟡 进行中 |
| 优先级 | 高 |
| 影响范围 | shop-backend |

**问题描述**：协程开启在大并发下可能会暴涨，缺乏协程池控制。

**参考方案**：
- `D:\code\goshopadmin\.trae\documents\goroutine_optimization_enhanced_plan.md`
- `D:\code\goshopadmin\.trae\documents\goroutine_optimization_implementation_plan.md`

**备注**：首次使用 superpowers-zh 完成该功能，待功能测试。

---

### 1.2 WebSocket 引入

| 属性 | 值 |
|:---|:---|
| 状态 | ⏸️ 暂缓 |
| 优先级 | 最低 |
| 影响范围 | 全项目 |

**问题描述**：需要引入 WebSocket 功能实现实时通知。

**参考方案**：`d:\code\goshopadmin\.trae\documents\websocket-notification-plan.md`

---

### 1.3 多商户订单处理方案

| 属性 | 值 |
|:---|:---|
| 状态 | ⏸️ 暂缓 |
| 优先级 | 低 |
| 影响范围 | shop-backend |

**问题描述**：需要实现多商户的订单处理，目前默认为 1。

**参考方案**：`D:\code\goshopadmin\docs\多商户订单处理方案.md`

**备注**：该项目着重关注 Redis、MQ、WebSocket、Nginx 抗压架构上，多商户暂缓。

---

### 1.4 保持三个服务 JWT 包版本一致

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 高 |
| 影响范围 | backend / shop-backend / search-service |

**问题描述**：保持三个服务 `auth.go` 的验签和生成 token 用的包相同，避免以后出现别的问题（JWT 版本保持一致）。

**涉及文件**：
- `D:\code\goshopadmin\backend\middleware\auth.go`
- `D:\code\goshopadmin\shop-backend\middleware\auth.go`
- `D:\code\goshopadmin\search-service\middleware\auth.go`

---

## 二、监控与日志

### 2.1 日志文件记录长记录问题

| 属性 | 值 |
|:---|:---|
| 状态 | 🟡 进行中 |
| 优先级 | 高 |
| 影响范围 | backend / shop-backend |

**问题描述**：日志对于一些大文件请求的写入非常消耗资源，需要过滤掉图片、文件等大文件的 response/request 日志记录。

**备注**：首次使用 superpowers-zh 完成该功能，待功能测试。

---

### 2.2 引入更详细 monitor.go 监控

| 属性 | 值 |
|:---|:---|
| 状态 | 🟡 进行中 |
| 优先级 | 高 |
| 影响范围 | backend / shop-backend |

**问题描述**：引入更详细的运行时监控指标。

**备注**：首次使用 superpowers-zh 完成该功能，待功能测试。

---

### 2.3 引入 ELK

| 属性 | 值 |
|:---|:---|
| 状态 | 🟡 进行中 |
| 优先级 | 高 |
| 影响范围 | 全项目 |

**问题描述**：引入 ELK 日志收集与分析系统。

**备注**：功能测试初步完成，改日完整测一下。

---

## 三、缓存与数据一致性

### 3.1 SKU 新增/更新库存时缓存更新检查

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 低 |
| 影响范围 | shop-backend |

**问题描述**：检查 SKU 新增/更新库存的时候有没有同步更新 Redis 缓存，可能存在数据不一致。

---

## 四、搜索服务（search-service）代码审查问题

> 以下问题来自 2026-06-10 代码审查，修复搜索功能 Bug 后发现。

### 4.1 🔴 Critical：全量同步无分页，大数据量下可能 OOM

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 高 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |
| 位置 | `syncProductSkusFull()` / `syncOrderItemsFull()` |

**问题描述**：`syncProductSkusFull()` 一次性 `SELECT * FROM product_skus` 加载所有 SKU 到内存，`syncOrderItemsFull()` 同理。如果数据量增长到数万/数十万条，会导致：
- Go 进程内存暴涨甚至 OOM
- 单次 Bulk 请求体过大，ES 可能拒绝（默认 `http.max_content_length=100MB`）
- 30 秒超时可能不够

**修复方案**：增加分页查询逻辑，每批处理 500~1000 条，分批 Bulk 写入 ES。

```go
const batchSize = 500
offset := 0
for {
    var batch []orderItemRow
    result := database.Raw(sql + " LIMIT ? OFFSET ?", batchSize, offset).Scan(&batch)
    if len(batch) == 0 { break }
    // 处理本批次...
    offset += batchSize
}
```

---

### 4.2 🔴 Critical：增量 SKU 同步缺少 product_image 字段

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 高 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |
| 位置 | `syncProductSkus()` |

**问题描述**：增量同步的 SQL 只查了 `product_skus` 表，没有像 `syncOrderItems()` 那样关联 `product_images` 表获取 `product_image`。而 `skuRow` 结构体也没有 `ProductImage` 字段，构建的 `skuDocs` 也没有 `product_image` 字段。

**修复方案**：如果 SKU 级别确实需要 `product_image`，需要在增量 SKU 同步的 SQL 中也关联 `product_images` 表，并在 `skuDocs` 中包含该字段。如果 SKU 不需要图片（图片在 product 顶层），则确认当前逻辑是否正确。

---

### 4.3 🔴 Critical：WildcardQuery 前缀通配符性能风险

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 高 |
| 影响范围 | search-service |
| 文件 | `search-service/services/customer_service.go` |
| 位置 | L67-80 |

**问题描述**：对 `username`、`phone`、`email`、`nickname.keyword` 四个字段都使用了 `*keyword*` 前缀通配符查询。ES 中前缀通配符（以 `*` 开头）无法利用倒排索引，必须全索引扫描，在数据量大时性能极差。

**修复方案**：
- **短期**：对 `username`/`phone` 等短字段，考虑使用 `prefix` 查询替代（用户通常输入前缀而非中间片段），或者对 WildcardQuery 的关键词长度做限制（如最少 2-3 个字符），避免单字符全索引扫描
- **中期**：如果必须支持中间子串搜索，在 ES mapping 中为这些字段添加 n-gram 子字段，用 MatchQuery 替代 WildcardQuery

---

### 4.4 🟡 Important：escapeWildcardChars 缺少反斜杠转义

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | search-service |
| 文件 | `search-service/services/order_service.go` |
| 位置 | L254-257 |

**问题描述**：当前只转义了 `*` 和 `?`，但 ES WildcardQuery 还会将 `\` 本身视为转义字符。如果用户输入包含 `\`，可能导致查询异常。

**修复方案**：先转义反斜杠 `\` → `\\`，再转义 `*` 和 `?`：

```go
func escapeWildcardChars(s string) string {
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, "*", "\\*")
    s = strings.ReplaceAll(s, "?", "\\?")
    return s
}
```

---

### 4.5 🟡 Important：SKU Attributes 反序列化后以 map 写入 ES，可能与 mapping 不一致

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |
| 位置 | `syncProductSkus()` / `syncProductSkusFull()` |

**问题描述**：在 `syncOrderItems` 中，`sku_attributes` 被正确地转为 JSON 字符串（匹配 ES keyword 类型）。但在 `syncProductSkus` 中，`sku.Attributes` 被反序列化为 `map[string]string` 后直接写入 ES 的 `attributes` 字段。如果 ES mapping 中 `attributes` 是 keyword 类型，写入 map 会导致类型不匹配错误。

**修复方案**：确认 ES mapping 中 `attributes` 的实际类型，确保写入格式与 mapping 一致。

---

### 4.6 🟡 Important：json.Unmarshal 错误被静默忽略

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |
| 位置 | L189, L409 |

**问题描述**：`json.Unmarshal(sku.Attributes, &attributes)` 如果包含非法 JSON，反序列化会失败但错误被忽略，`attributes` 保持零值（nil map），最终写入 ES 的是 `null`。

**修复方案**：至少记录一条警告日志。

---

### 4.7 🟡 Important：syncOrderItems() 空检查缺少日志

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |
| 位置 | L254-256 |

**问题描述**：对比 `syncProductSkus()` 的空检查有 `utils.Warn` 日志，但 `syncOrderItems()` 的同样检查没有日志。应保持一致。

---

### 4.8 🟡 Important：Search.vue handleClear 未重置 currentPage

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | frontend |
| 文件 | `frontend/src/views/search/Search.vue` |
| 位置 | L182-187 |

**问题描述**：`handleClear` 清空后 `hasSearched` 置为 false，但未重置 `currentPage`。如果用户清除后再次搜索，`currentPage` 可能还保留上次的值，导致首次搜索就从非第一页开始。

**修复方案**：在 `handleClear` 中添加 `currentPage.value = 1`。

---

### 4.9 🟡 Important：Orders.vue 使用 console.error 而非 ElMessage.error

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | frontend |
| 文件 | `frontend/src/views/orders/Orders.vue` |
| 位置 | L125 |

**问题描述**：对比 `Customers.vue` 使用了 `ElMessage.error('搜索客户失败，请稍后重试')`，Orders.vue 只在控制台打印错误，用户看不到任何提示。应统一使用 `ElMessage.error` 显示错误信息。

---

### 4.10 🟡 Important：customer_service.go 注释与实际行为不一致

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 中 |
| 影响范围 | search-service |
| 文件 | `search-service/services/customer_service.go` |
| 位置 | L23-24, L39-44 |

**问题描述**：注释写的是"用户名/手机号精确匹配 + 邮箱模糊匹配"，但实际代码已全部改为 WildcardQuery 通配符匹配。注释应同步更新。

---

### 4.11 🟢 Minor：全量同步和增量同步的 Bulk 超时时间不同但无注释说明

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 低 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |

**问题描述**：增量同步 Bulk 超时 10 秒，全量同步 Bulk 超时 30 秒。虽然逻辑上合理（全量数据更多），但建议加注释说明超时时间的选择依据。

---

### 4.12 🟢 Minor：runFullSync 在启动协程中同步执行，会阻塞定时器

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 低 |
| 影响范围 | search-service |
| 文件 | `search-service/services/sync_service.go` |
| 位置 | L78-92 |

**问题描述**：如果全量同步耗时数分钟，期间 `syncTicker` 的 tick 会被丢弃（channel 缓冲为 0），但不会丢失数据（因为增量同步使用 5 分钟窗口）。这不是 bug，但值得注意：全量同步完成前，增量同步不会执行。

---

### 4.13 🟢 Minor：Logstash orders.conf 中 MySQL 密码明文

| 属性 | 值 |
|:---|:---|
| 状态 | 🔴 待处理 |
| 优先级 | 低 |
| 影响范围 | docker/elk |

**问题描述**：`jdbc_password => "password"` 是开发环境默认密码，与项目其他配置文件一致。但建议生产环境使用环境变量注入，并在注释中提醒。

---

## 变更记录

| 日期 | 变更内容 |
|:---|:---|
| 2026-06-10 | 整理格式，新增搜索服务代码审查问题（4.1-4.13） |
| 2026-06-10 | 移除已完成/已合并的旧条目（原 #1 日志分割、#10 订单管理搜索、#11 搜索接口认证、#12 侧边栏删除、#13 Auth token 提取） |
