# Shop-Backend MQ 工作流程分析文档

本文档详细梳理 `shop-backend` 项目中 RabbitMQ 消息队列的工作池运作流程。

---

## 一、架构概览

### 1.1 核心组件

| 组件 | 文件位置 | 功能描述 |
| :--- | :--- | :--- |
| **Connection** | `pkg/mq/connection.go` | RabbitMQ 连接管理，支持重连机制 |
| **ConnectionPool** | `pkg/mq/connection_pool.go` | MQ 连接池，管理连接的获取和归还 |
| **Producer** | `pkg/mq/producer.go` | 消息生产者，支持普通消息和 TTL 延迟消息 |
| **Consumer** | `pkg/mq/consumer.go` | 消息消费者基类，支持重试机制 |
| **WorkerPool** | `utils/worker_pool.go` | 工作池，异步处理任务 |
| **PoolAccess** | `pkg/pool/pool_access.go` | 池访问接口，提供全局访问函数 |

### 1.2 初始化流程

**位置**: `main.go`

```
启动顺序:
1. 创建 MQ 连接池 (minConns=5, maxConns=50)
2. 创建工作池 (minWorkers=2, maxWorkers=CPU*4, queueSize=5000)
3. 注册池获取函数到 pool_access
4. 启动协程监控
5. 初始化 MQ 消费者 (异步)
```

---

## 二、交换机配置

### 2.1 交换机列表

| 交换机名称 | 类型 | 持久化 | 用途 |
| :--- | :--- | :--- | :--- |
| `activity_exchange` | `direct` | ✓ | 活动订单消息路由 |
| `order_status_exchange` | `fanout` | ✓ | 订单状态变更广播 |
| `dead_letter_exchange` | `direct` | ✓ | 死信消息处理 |

### 2.2 交换机声明位置

| 交换机 | 声明位置 | 代码行 |
| :--- | :--- | :--- |
| `activity_exchange` | `consumer_init.go` | L32 |
| `order_status_exchange` | `consumer_init.go` | L46 |
| `dead_letter_exchange` | `delay_queue.go` | L14 |

---

## 三、队列配置

### 3.1 队列列表

| 队列名称 | 持久化 | 死信配置 | 用途 |
| :--- | :--- | :--- | :--- |
| `activity_order_queue` | ✓ | - | 处理活动订单创建请求 |
| `activity_order_delay_queue` | ✓ | → `dead_letter_exchange` | 活动订单 30 分钟延迟 |
| `activity_order_dead_letter_queue` | ✓ | - | 活动订单超时处理 |
| `activity_order_alert_queue` | ✓ | - | 重试超限告警 |
| `order_delay_queue` | ✓ | → `dead_letter_exchange` | 普通订单 30 分钟延迟 |
| `order_dead_letter_queue` | ✓ | - | 普通订单超时处理 |
| `order_status_queue` | ✓ | - | 订单状态变更记录 |

### 3.2 队列绑定关系

| 队列 | 交换机 | 路由键 | 绑定位置 |
| :--- | :--- | :--- | :--- |
| `activity_order_queue` | `activity_exchange` | `activity_order` | `consumer_init.go:40` |
| `order_status_queue` | `order_status_exchange` | `""` (fanout) | `consumer_init.go:54` |
| `order_dead_letter_queue` | `dead_letter_exchange` | `dead_letter` | `delay_queue.go:26` |
| `activity_order_dead_letter_queue` | `dead_letter_exchange` | `dead_letter` | `delay_queue.go:26` |

### 3.3 延迟队列配置

**实现方式**: 死信队列 + TTL

```go
// 延迟队列声明参数 (delay_queue.go:32-42)
amqp091.Table{
    "x-dead-letter-exchange":    constants.MQExchangeDeadLetter,
    "x-dead-letter-routing-key": constants.MQRoutingKeyDeadLetter,
}
```

**TTL 超时时间**: `30 * 60 * 1000` = 30 分钟 = 1,800,000 毫秒

---

## 四、路由键配置

| 路由键 | 关联交换机 | 用途 |
| :--- | :--- | :--- |
| `activity_order` | `activity_exchange` | 活动订单消息路由 |
| `order_status` | `order_status_exchange` | 订单状态变更消息 |
| `dead_letter` | `dead_letter_exchange` | 死信消息路由 |

---

## 五、生产者逻辑分析

### 5.1 生产者方法

| 方法 | 文件位置 | 功能 |
| :--- | :--- | :--- |
| `Publish()` | `producer.go:24-51` | 发布普通消息 |
| `PublishWithTTL()` | `producer.go:54-82` | 发布带 TTL 的延迟消息 |

### 5.2 生产者调用位置

#### 5.2.1 活动订单创建

**位置**: `controllers/activity_order_controller.go:56-78`

```go
// 消息结构
msg := map[string]interface{}{
    "customer_id": customerID,
    "activity_id": req.ActivityID,
    "address_id":  req.AddressID,
    "items":       req.Items,
}

// 发布到活动交换机
producer.Publish(constants.MQExchangeActivity, constants.MQRoutingKeyActivityOrder, msg)
```

**流程**:
```
用户请求 → Controller → WorkerPool.SubmitTask → 
获取 MQ 连接 → Producer.Publish → activity_exchange → 
activity_order_queue → ActivityConsumer.HandleActivityOrder
```

#### 5.2.2 普通订单超时延迟

**位置**: `controllers/order_controller.go:82-103`

```go
// 消息结构
msg := map[string]interface{}{
    "order_id":   order.OrderID,
    "created_at": time.Now(),
}

// 发布到延迟队列（30分钟 TTL）
producer.PublishWithTTL("", constants.MQQueueOrderDelay, msg, constants.MQOrderTimeoutTTL)
```

**流程**:
```
创建订单 → WorkerPool.SubmitTask → 
Producer.PublishWithTTL → order_delay_queue (30分钟) → 
order_dead_letter_queue → OrderConsumer.HandleTimeoutOrder
```

#### 5.2.3 活动订单超时延迟

**位置**: `pkg/mq/activity_consumer.go:72-98`

```go
// 消息结构
delayMsg := map[string]interface{}{
    "order_id":    order.OrderID,
    "created_at":  order.CreatedAt,
    "retry_count": 0,
}

// 发布到延迟队列（30分钟 TTL）
producer.PublishWithTTL("", constants.MQQueueActivityOrderDelay, delayMsg, constants.MQOrderTimeoutTTL)
```

**流程**:
```
HandleActivityOrder 创建订单成功 → WorkerPool.SubmitTask → 
Producer.PublishWithTTL → activity_order_delay_queue (30分钟) → 
activity_order_dead_letter_queue → ActivityConsumer.HandleTimeoutActivityOrder
```

#### 5.2.4 订单状态变更

**位置**: `controllers/payment_controller.go:68-77` 和 `payment_controller.go:133-142`

```go
// 消息结构
msg := map[string]interface{}{
    "order_id":   order.ID,
    "status":     constants.OrderStatusPaid,
    "updated_at": time.Now(),
}

// 发布到状态交换机
producer.Publish(constants.MQExchangeOrderStatus, constants.MQRoutingKeyOrderStatus, msg)
```

**流程**:
```
支付成功 → WorkerPool.SubmitTask → 
Producer.Publish → order_status_exchange → 
order_status_queue → StatusConsumer.HandleOrderStatus
```

#### 5.2.5 重试消息发送

**位置**: `pkg/mq/consumer.go:80-98`

```go
// 消息结构（增加 retry_count）
msg["retry_count"] = currentRetry + 1

// 重新发送到延迟队列
producer.PublishWithTTL("", delayQueue, msg, ttl)
```

#### 5.2.6 告警消息发送

**位置**: `pkg/mq/consumer.go:100-117`

```go
// 告警消息结构
body := map[string]interface{}{
    "original_body":  string(msg.Body),
    "retry_count":    retryCount,
    "original_queue": queue,
    "arrival_time":   msg.Timestamp,
}

// 发送到告警队列
producer.Publish("", constants.MQQueueActivityOrderAlert, body)
```

---

## 六、消费者逻辑分析

### 6.1 消费者注册

**位置**: `pkg/mq/consumer_init.go`

| 消费者 | 队列 | 处理方法 | 重试配置 |
| :--- | :--- | :--- | :--- |
| OrderConsumer | `order_dead_letter_queue` | `HandleTimeoutOrder` | 无 |
| ActivityConsumer | `activity_order_queue` | `HandleActivityOrder` | 无 |
| ActivityConsumer | `activity_order_dead_letter_queue` | `HandleTimeoutActivityOrder` | 有 (DelayQueue + TTL) |
| ActivityConsumer | `activity_order_alert_queue` | `HandleAlertMessage` | 无 |
| StatusConsumer | `order_status_queue` | `HandleOrderStatus` | 无 |

### 6.2 消费者处理逻辑详解

#### 6.2.1 ActivityConsumer.HandleActivityOrder

**位置**: `pkg/mq/activity_consumer.go:32-102`

**处理流程**:
```
1. 解析消息 (customer_id, activity_id, address_id, items)
2. 调用 ActivityOrderService.CreateActivityOrder 创建订单
3. 创建成功后，通过 WorkerPool 发送 30 分钟延迟消息
4. 返回 nil 表示处理成功
```

**错误处理**:
- 解析失败 → 返回错误，消息 Nack
- 创建订单失败 → 返回错误，消息 Nack

#### 6.2.2 ActivityConsumer.HandleTimeoutActivityOrder

**位置**: `pkg/mq/activity_consumer.go:104-148`

**处理流程**:
```
1. 解析消息 (order_id, created_at)
2. 查询订单信息 (getOrderForTimeout)
3. 检查订单状态是否为终态 (cancelled/completed/shipped)
   - 是终态 → 返回 nil，跳过处理
   - 不是终态 → 继续处理
4. 调用 ActivityOrderService.CancelActivityOrder 取消订单
5. 返回 nil 表示处理成功
```

**重试机制**:
- 配置了 RetryConfig: `{DelayQueue: activity_order_delay_queue, TTL: 30分钟}`
- 最大重试次数: 3 次 (`MaxRetryCount`)
- 重试超限后发送到告警队列

**错误处理**:
- 解析失败 → 返回错误，触发重试
- 订单不存在 → 返回错误，触发重试
- 取消失败 → 返回错误，触发重试
- 重试超限 → 发送到告警队列，消息 Ack

#### 6.2.3 OrderConsumer.HandleTimeoutOrder

**位置**: `pkg/mq/order_consumer.go:23-57`

**处理流程**:
```
1. 解析消息 (order_id, created_at)
2. 调用 OrderService.GetOrderByOrderNo 查询订单
3. 调用 OrderService.CancelOrder 取消订单
4. 返回 nil 表示处理成功
```

**错误处理**:
- 解析失败 → 返回错误，消息 Nack (无重试配置)
- 订单不存在 → 返回错误，消息 Nack
- 取消失败 → 返回错误，消息 Nack

#### 6.2.4 StatusConsumer.HandleOrderStatus

**位置**: `pkg/mq/status_consumer.go:19-40`

**处理流程**:
```
1. 解析消息 (customer_id, order_id, order_no, status, updated_at)
2. 记录状态变更日志
3. 返回 nil 表示处理成功
```

**注意**: 当前仅记录日志，未实现具体业务逻辑

#### 6.2.5 ActivityConsumer.HandleAlertMessage

**位置**: `pkg/mq/activity_consumer.go:175-180`

**处理流程**:
```
1. 记录告警日志
2. TODO: 实现邮件通知运维功能
3. 返回 nil
```

---

## 七、消息结构汇总

### 7.1 生产者消息结构

#### 7.1.1 活动订单创建消息

**队列**: `activity_order_queue`

```json
{
    "customer_id": 1,
    "activity_id": 100,
    "address_id": 5,
    "items": [
        {
            "product_id": 10,
            "sku_id": 20,
            "quantity": 2
        }
    ]
}
```

**字段说明**:

| 字段 | 类型 | 来源 | 说明 |
| :--- | :--- | :--- | :--- |
| `customer_id` | int | `ctx.Get("customer_id")` | 消费者 ID |
| `activity_id` | int | `req.ActivityID` | 活动 ID |
| `address_id` | int | `req.AddressID` | 收货地址 ID |
| `items` | array | `req.Items` | 商品列表 |
| `product_id` | int | `item.ProductID` | 商品 ID |
| `sku_id` | int | `item.SkuID` | SKU ID |
| `quantity` | int | `item.Quantity` | 数量 |

#### 7.1.2 普通订单超时延迟消息

**队列**: `order_delay_queue`

```json
{
    "order_no": "ORD20260101123456",
    "created_at": "2026-01-01T12:34:56Z"
}
```

**字段说明**:

| 字段 | 类型 | 来源 | 说明 |
| :--- | :--- | :--- | :--- |
| `order_no` | string | `order.OrderNo` | 订单号 |
| `created_at` | time.Time | `time.Now()` | 创建时间 |

#### 7.1.3 活动订单超时延迟消息

**队列**: `activity_order_delay_queue`

```json
{
    "order_no": "ORD20260101123456",
    "created_at": "2026-01-01T12:34:56Z",
    "retry_count": 0
}
```

**字段说明**:

| 字段 | 类型 | 来源 | 说明 |
| :--- | :--- | :--- | :--- |
| `order_no` | string | `order.OrderID` | 订单号 |
| `created_at` | string | `order.CreatedAt` | 创建时间 |
| `retry_count` | int | 初始为 0 | 重试计数 |

#### 7.1.4 订单状态变更消息

**队列**: `order_status_queue`

```json
{
    "order_id": 123,
    "status": "paid",
    "updated_at": "2026-01-01T12:34:56Z"
}
```

**字段说明**:

| 字段 | 类型 | 来源 | 说明 |
| :--- | :--- | :--- | :--- |
| `order_id` | int | `order.ID` | 订单 ID (数据库主键) |
| `status` | string | `constants.OrderStatusPaid` | 新状态 |
| `updated_at` | time.Time | `time.Now()` | 更新时间 |

#### 7.1.5 告警消息

**队列**: `activity_order_alert_queue`

```json
{
    "original_body": "{\"order_id\":\"ORD...\",\"retry_count\":3}",
    "retry_count": 3,
    "original_queue": "activity_order_dead_letter_queue",
    "arrival_time": "2026-01-01T12:34:56Z"
}
```

**字段说明**:

| 字段 | 类型 | 来源 | 说明 |
| :--- | :--- | :--- | :--- |
| `original_body` | string | `string(msg.Body)` | 原始消息体 |
| `retry_count` | int | 重试计数 | 重试次数 |
| `original_queue` | string | `queue` | 原始队列名 |
| `arrival_time` | time.Time | `msg.Timestamp` | 消息到达时间 |

### 7.2 消费者解析结构

#### 7.2.1 ActivityConsumer.HandleActivityOrder 解析结构

```go
var req struct {
    CustomerID int `json:"customer_id"`
    ActivityID int `json:"activity_id"`
    AddressID  int `json:"address_id"`
    Items      []struct {
        ProductID int `json:"product_id"`
        SkuID     int `json:"sku_id"`
        Quantity  int `json:"quantity"`
    }
}
```

#### 7.2.2 ActivityConsumer.HandleTimeoutActivityOrder 解析结构

```go
var message struct {
    OrderNo   string `json:"order_no"`
    CreatedAt string `json:"created_at"`
}
```

#### 7.2.3 OrderConsumer.HandleTimeoutOrder 解析结构

```go
var message struct {
    OrderNo   string `json:"order_no"`
    CreatedAt string `json:"created_at"`
}
```

#### 7.2.4 StatusConsumer.HandleOrderStatus 解析结构

```go
var message struct {
    CustomerID int    `json:"customer_id"`
    OrderID    int    `json:"order_id"`
    OrderNo    string `json:"order_no"`
    Status     string `json:"status"`
    UpdatedAt  string `json:"updated_at"`
}
```

---

## 八、变量命名规范检查

### 8.1 检查结果

**所有消息字段均使用 snake_case 命名，符合规范。**

| 消息类型 | 字段命名 | 规范状态 |
| :--- | :--- | :--- |
| 活动订单创建 | `customer_id`, `activity_id`, `address_id`, `product_id`, `sku_id`, `quantity` | ✓ 符合 |
| 普通订单延迟 | `order_no`, `created_at` | ✓ 符合 |
| 活动订单延迟 | `order_no`, `created_at`, `retry_count` | ✓ 符合 |
| 订单状态变更 | `order_id`, `status`, `updated_at` | ✓ 符合 |
| 告警消息 | `original_body`, `retry_count`, `original_queue`, `arrival_time` | ✓ 符合 |
| 状态消费者 | `customer_id`, `order_id`, `order_no`, `status`, `updated_at` | ✓ 符合 |

### 8.2 无命名问题

经过全面检查，所有 MQ 消息结构中的变量命名均符合项目规范（snake_case），无需修改。

---

## 九、工作流程图

### 9.1 普通订单超时流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           普通订单超时处理流程                                  │
└─────────────────────────────────────────────────────────────────────────────┘

用户创建订单
    │
    ▼
OrderController.CreateOrder
    │
    ▼
WorkerPool.SubmitTask ──────────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
订单创建成功                                                               │
    │                                                                        │
    ▼                                                                        │
Producer.PublishWithTTL                                                    │
    │                                                                        │
    │  消息: {order_id, created_at}                                          │
    │  TTL: 30分钟                                                           │
    ▼                                                                        │
order_delay_queue ─────────────────────────────────────────────────────────►
    │                                                                        │
    │  等待 30 分钟                                                           │
    │  (消息过期)                                                             │
    ▼                                                                        │
dead_letter_exchange (路由键: dead_letter)                                  │
    │                                                                        │
    ▼                                                                        │
order_dead_letter_queue ───────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
OrderConsumer.HandleTimeoutOrder                                           │
    │                                                                        │
    │  1. 查询订单                                                            │
    │  2. 取消订单                                                            │
    ▼                                                                        │
处理完成                                                                    │
```

### 9.2 活动订单处理流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           活动订单处理流程                                     │
└─────────────────────────────────────────────────────────────────────────────┘

用户创建活动订单
    │
    ▼
ActivityOrderController.CreateActivityOrder
    │
    ▼
WorkerPool.SubmitTask ──────────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
Producer.Publish                                                           │
    │                                                                        │
    │  交换机: activity_exchange                                              │
    │  路由键: activity_order                                                 │
    │  消息: {customer_id, activity_id, address_id, items}                   │
    ▼                                                                        │
activity_exchange ─────────────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
activity_order_queue ──────────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
ActivityConsumer.HandleActivityOrder                                       │
    │                                                                        │
    │  1. 解析消息                                                            │
    │  2. 创建活动订单                                                         │
    │  3. 发送延迟消息                                                         │
    ▼                                                                        │
WorkerPool.SubmitTask (发送延迟消息)                                         │
    │                                                                        │
    ▼                                                                        │
Producer.PublishWithTTL                                                    │
    │                                                                        │
    │  消息: {order_id, created_at, retry_count: 0}                          │
    │  TTL: 30分钟                                                           │
    ▼                                                                        │
activity_order_delay_queue ───────────────────────────────────────────────►
    │                                                                        │
    │  等待 30 分钟                                                           │
    │  (消息过期)                                                             │
    ▼                                                                        │
dead_letter_exchange (路由键: dead_letter)                                  │
    │                                                                        │
    ▼                                                                        │
activity_order_dead_letter_queue ─────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
ActivityConsumer.HandleTimeoutActivityOrder                                │
    │                                                                        │
    │  1. 查询订单                                                            │
    │  2. 检查状态是否终态                                                     │
    │     ├─ 是终态 → 跳过                                                    │
    │     └─ 不是终态 → 取消订单                                              │
    │                                                                        │
    │  失败时触发重试机制:                                                     │
    │     ├─ retry_count < 3 → 重新发送到 delay_queue                         │
    │     └─ retry_count >= 3 → 发送到 alert_queue                           │
    ▼                                                                        │
处理完成                                                                    │
```

### 9.3 订单状态变更流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           订单状态变更流程                                     │
└─────────────────────────────────────────────────────────────────────────────┘

用户支付成功
    │
    ▼
PaymentController.FakePay / PaymentCallback
    │
    ▼
WorkerPool.SubmitTask ──────────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
更新订单状态                                                                │
    │                                                                        │
    ▼                                                                        │
Producer.Publish                                                           │
    │                                                                        │
    │  交换机: order_status_exchange                                          │
    │  路由键: order_status                                                   │
    │  消息: {order_id, status, updated_at}                                  │
    ▼                                                                        │
order_status_exchange (fanout) ────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
order_status_queue ────────────────────────────────────────────────────────►
    │                                                                        │
    ▼                                                                        │
StatusConsumer.HandleOrderStatus                                           │
    │                                                                        │
    │  1. 解析消息                                                            │
    │  2. 记录状态变更日志                                                     │
    ▼                                                                        │
处理完成                                                                    │
```

---

## 十、重试机制详解

### 10.1 重试配置

**最大重试次数**: `MaxRetryCount = 3`

**重试间隔**: 30 分钟 (与订单超时时间相同)

**适用队列**: 仅 `activity_order_dead_letter_queue` 配置了重试机制

### 10.2 重试流程

```
消息处理失败
    │
    ▼
Consumer.Consume 检查 retry_count
    │
    ├─ retry_count < 3
    │   │
    │   ▼
    │   incrementRetryCountAndResend
    │   │
    │   │  retry_count++
    │   │
    │   ▼
    │   Producer.PublishWithTTL → delay_queue
    │   │
    │   ▼
    │   30 分钟后重新进入 dead_letter_queue
    │
    └─ retry_count >= 3
        │
        ▼
        sendToAlertQueue
        │
        ▼
        activity_order_alert_queue
        │
        ▼
        HandleAlertMessage (记录告警)
```

### 10.3 重试计数来源

**两种来源**:

1. **消息体中的 retry_count 字段** (主要)
   - 通过 `getRetryCountFromBody()` 获取
   - 每次重试时递增

2. **x-death 头信息** (备用)
   - 通过 `getRetryCountFromXDeath()` 获取
   - RabbitMQ 自动记录的死信信息

---

## 十一、连接池管理

### 11.1 MQ 连接池配置

**位置**: `pkg/mq/connection_pool.go`

| 参数 | 值 | 说明 |
| :--- | :--- | :--- |
| `minConns` | 5 | 最小连接数 |
| `maxConns` | 50 | 最大连接数 |
| `idleTimeout` | 5 分钟 | 连接空闲超时 |
| `healthCheckInt` | 30 秒 | 健康检查间隔 |
| `maxUseCount` | 1000 | 连接最大使用次数 |

### 11.2 工作池配置

**位置**: `utils/worker_pool.go`

| 参数 | 值 | 说明 |
| :--- | :--- | :--- |
| `minWorkers` | 2 | 最小工作协程数 |
| `maxWorkers` | CPU * 4 | 最大工作协程数 |
| `queueSize` | 5000 | 任务队列容量 |
| `scaleCheckInt` | 10 秒 | 伸缩检查间隔 |

### 11.3 池访问接口

**位置**: `pkg/pool/pool_access.go`

提供全局访问函数:
- `GetMQConn()` - 获取 MQ 连接
- `PutMQConn(conn)` - 归还 MQ 连接
- `SubmitTask(fn)` - 提交异步任务

---

## 十二、待优化项

### 12.1 功能待实现

| 项目 | 位置 | 说明 |
| :--- | :--- | :--- |
| 告警邮件通知 | `activity_consumer.go:178` | `HandleAlertMessage` 仅记录日志，需实现邮件通知 |
| 状态变更业务 | `status_consumer.go` | `HandleOrderStatus` 仅记录日志，可扩展业务逻辑 |

### 12.2 潜在优化点

1. **普通订单超时无重试机制**
   - `order_dead_letter_queue` 未配置 RetryConfig
   - 处理失败直接 Nack，可能导致订单无法取消

2. **消息结构不一致**
   - 普通订单延迟消息的 `created_at` 是 `time.Time` 类型
   - 活动订单延迟消息的 `created_at` 是 `string` 类型
   - 建议统一为 `string` 类型

3. **状态消费者消息字段不完整**
   - 生产者发送 `{order_id, status, updated_at}`
   - 消费者解析包含 `{customer_id, order_id, order_no, status, updated_at}`
   - 生产者缺少 `customer_id` 和 `order_no` 字段

---

## 十三、文件索引

| 文件 | 路径 | 主要功能 |
| :--- | :--- | :--- |
| connection.go | `pkg/mq/connection.go` | MQ 连接管理 |
| connection_pool.go | `pkg/mq/connection_pool.go` | 连接池实现 |
| producer.go | `pkg/mq/producer.go` | 消息生产者 |
| consumer.go | `pkg/mq/consumer.go` | 消息消费者基类 |
| consumer_init.go | `pkg/mq/consumer_init.go` | 消费者初始化 |
| delay_queue.go | `pkg/mq/delay_queue.go` | 延迟队列设置 |
| activity_consumer.go | `pkg/mq/activity_consumer.go` | 活动订单消费者 |
| order_consumer.go | `pkg/mq/order_consumer.go` | 订单超时消费者 |
| status_consumer.go | `pkg/mq/status_consumer.go` | 状态变更消费者 |
| pool_access.go | `pkg/pool/pool_access.go` | 池访问接口 |
| worker_pool.go | `utils/worker_pool.go` | 工作池实现 |
| constants.go | `constants/constants.go` | MQ 常量定义 |
| mq_config.go | `config/mq_config.go` | MQ 配置 |
| main.go | `main.go` | 程序入口，初始化 |
| order_controller.go | `controllers/order_controller.go` | 订单控制器 |
| payment_controller.go | `controllers/payment_controller.go` | 支付控制器 |
| activity_order_controller.go | `controllers/activity_order_controller.go` | 活动订单控制器 |

---

*文档版本: 1.0*
*生成时间: 2026-05-28*
*适用项目: shop-backend*