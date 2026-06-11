# GoShopAdmin 待修 Bug 与优化列表

> 最后更新：2026-06-11
> 状态标记：🔴 待处理 | 🟡 进行中 | 🟢 已完成 | ⏸️ 暂缓

***

## 一、进行中任务

### 1.1 MQ 连接协程优化

| 属性   | 值            |
| :--- | :----------- |
| 状态   | 🟢 已完成       |
| 优先级  | 高            |
| 影响范围 | shop-backend |

**问题描述**：协程开启在大并发下可能会暴涨，缺乏协程池控制。

**备注**：使用 superpowers-zh 完成该功能，暂时未发现bug。

***

### 1.2 引入更详细 monitor.go 监控

| 属性   | 值                      |
| :--- | :--------------------- |
| 状态   | 🟢 已完成                 |
| 优先级  | 高                      |
| 影响范围 | backend / shop-backend |

**问题描述**：引入更详细的运行时监控指标。

**备注**：使用 superpowers-zh 完成该功能，暂时未发现bug。

***

## 二、暂缓任务

### 2.1 WebSocket 引入

| 属性   | 值     |
| :--- | :---- |
| 状态   | ⏸️ 暂缓 |
| 优先级  | 最低    |
| 影响范围 | 全项目   |

**问题描述**：需要引入 WebSocket 功能实现实时通知。

**参考方案**：`d:\code\goshopadmin\.trae\documents\websocket-notification-plan.md`

***

### 2.2 多商户订单处理方案

| 属性   | 值            |
| :--- | :----------- |
| 状态   | ⏸️ 暂缓        |
| 优先级  | 低            |
| 影响范围 | shop-backend |

**问题描述**：需要实现多商户的订单处理，目前默认为 1。

**参考方案**：`D:\code\goshopadmin\docs\多商户订单处理方案.md`

**备注**：该项目着重关注 Redis、MQ、WebSocket、Nginx 抗压架构上，多商户暂缓。

***

## 三、已完成任务

### 3.1 WildcardQuery 前缀通配符性能风险

| 属性     | 值                                             |
| :----- | :-------------------------------------------- |
| 状态     | 🟢 已完成                                        |
| 优先级    | 高                                             |
| 影响范围   | search-service                                |
| 文件     | `search-service/services/customer_service.go` |
| 位置     | L67-80                                        |
| 修复日期   | 2026-06-11                                    |
| Commit | `e23a2e9`                                     |

**问题描述**：对 `username`、`phone`、`email`、`nickname.keyword` 四个字段都使用了 `*keyword*` 前缀通配符查询。ES 中前缀通配符（以 `*` 开头）无法利用倒排索引，必须全索引扫描，在数据量大时性能极差。

**修复方案**：

- **短期**：对 `username`/`phone` 等短字段，考虑使用 `prefix` 查询替代（用户通常输入前缀而非中间片段），或者对 WildcardQuery 的关键词长度做限制（如最少 2-3 个字符），避免单字符全索引扫描
- **中期**：如果必须支持中间子串搜索，在 ES mapping 中为这些字段添加 n-gram 子字段，用 MatchQuery 替代 WildcardQuery

**实际修复**：

- `username`/`phone` 改用 `PrefixQuery`（前缀匹配，性能远优于通配符子串匹配）
- `email`/`nickname.keyword` 保留 `WildcardQuery`（需要子串搜索）
- 新增关键词长度限制：至少 2 个字符（rune）才触发 prefix/wildcard 查询，1 字符仅走 IK 分词搜索
- 新增 `shouldUsePrefixWildcard()` 辅助函数，使用 `utf8.RuneCountInString` 正确处理中文字符

**功能测试方法**：

1. **前缀匹配验证**：在客户搜索中输入用户名前缀（如 "ad"），应匹配 "admin" 等用户
2. **子串匹配验证**：搜索邮箱中间片段（如 "qq"），应匹配 "<xxx@qq.com>"
3. **短关键词验证**：输入 1 个字符搜索，应仅返回 IK 分词匹配结果，不触发全索引扫描
4. **中文关键词验证**：输入 1 个中文字（如"明"），应走 IK 分词；输入 2 个中文字（如"小明"），应同时走 prefix 和 IK 分词

***

### 3.2 escapeWildcardChars 缺少反斜杠转义

| 属性     | 值                                          |
| :----- | :----------------------------------------- |
| 状态     | 🟢 已完成                                     |
| 优先级    | 中                                          |
| 影响范围   | search-service                             |
| 文件     | `search-service/services/order_service.go` |
| 位置     | L254-257                                   |
| 修复日期   | 2026-06-11                                 |
| Commit | `3987e16`                                  |

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

**功能测试方法**：

1. 在订单搜索中输入包含反斜杠的关键词（如 `test\file`），确认不报错
2. 输入包含 `*` 或 `?` 的关键词，确认被当作普通字符而非通配符
3. 单元测试已覆盖 10 个用例（普通字符串、`*`、`?`、`\`、混合场景等）

***

### 3.3 json.Unmarshal 错误被静默忽略

| 属性     | 值                                         |
| :----- | :---------------------------------------- |
| 状态     | 🟢 已完成                                    |
| 优先级    | 中                                         |
| 影响范围   | search-service                            |
| 文件     | `search-service/services/sync_service.go` |
| 位置     | L189, L409                                |
| 修复日期   | 2026-06-11                                |
| Commit | `19ce5f5`                                 |

**问题描述**：`json.Unmarshal(sku.Attributes, &attributes)` 如果包含非法 JSON，反序列化会失败但错误被忽略，`attributes` 保持零值（nil map），最终写入 ES 的是 `null`。

**修复方案**：至少记录一条警告日志。

**实际修复**：在 `syncProductSkus()` 和 `syncProductSkusFull()` 中的 `json.Unmarshal` 调用添加错误检查和 `utils.Warn` 日志。

**功能测试方法**：

1. 在数据库中插入一条 attributes 字段为非法 JSON 的 SKU 记录（如 `{invalid}`）
2. 触发增量同步或全量同步
3. 查看 search-service 日志，确认出现 `SKU attributes 反序列化失败` 警告

***

### 3.4 Search.vue handleClear 未重置 currentPage

| 属性     | 值                                      |
| :----- | :------------------------------------- |
| 状态     | 🟢 已完成                                 |
| 优先级    | 中                                      |
| 影响范围   | frontend                               |
| 文件     | `frontend/src/views/search/Search.vue` |
| 位置     | L182-187                               |
| 修复日期   | 2026-06-11                             |
| Commit | `6b14b4a`                              |

**问题描述**：`handleClear` 清空后 `hasSearched` 置为 false，但未重置 `currentPage`。如果用户清除后再次搜索，`currentPage` 可能还保留上次的值，导致首次搜索就从非第一页开始。

**修复方案**：在 `handleClear` 中添加 `currentPage.value = 1`。

**功能测试方法**：

1. 在搜索页搜索关键词，翻到第 2 页
2. 点击清除按钮
3. 再次搜索，确认从第 1 页开始显示结果

***

### 3.5 Orders.vue 使用 console.error 而非 ElMessage.error

| 属性     | 值                                      |
| :----- | :------------------------------------- |
| 状态     | 🟢 已完成                                 |
| 优先级    | 中                                      |
| 影响范围   | frontend                               |
| 文件     | `frontend/src/views/orders/Orders.vue` |
| 位置     | L125                                   |
| 修复日期   | 2026-06-11                             |
| Commit | `e1295ac`                              |

**问题描述**：对比 `Customers.vue` 使用了 `ElMessage.error('搜索客户失败，请稍后重试')`，Orders.vue 只在控制台打印错误，用户看不到任何提示。应统一使用 `ElMessage.error` 显示错误信息。

**实际修复**：添加 `import { ElMessage } from 'element-plus'`，将 `console.error` 替换为 `ElMessage.error('搜索订单失败，请稍后重试')`。

**功能测试方法**：

1. 在订单管理页面搜索关键词，触发搜索失败（如临时停止 search-service）
2. 确认页面显示红色错误提示"搜索订单失败，请稍后重试"
3. 确认浏览器控制台不再只有 console.error

***

## 变更记录

| 日期         | 变更内容                                                                       |
| :--------- | :------------------------------------------------------------------------- |
| 2026-06-11 | 修复搜索服务全部代码审查问题（4.1-4.13），状态更新为 🟢 已完成；整理文件结构，按状态分组展示 |
| 2026-06-10 | 整理格式，新增搜索服务代码审查问题（4.1-4.13）                                                |
| 2026-06-10 | 移除已完成/已合并的旧条目（原 #1 日志分割、#10 订单管理搜索、#11 搜索接口认证、#12 侧边栏删除、#13 Auth token 提取） |