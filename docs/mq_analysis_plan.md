# GoShopAdmin 商城后端 MQ 代码分析文档

## 1. 整体架构

### 1.1 MQ 组件结构
- **Connection**: RabbitMQ 连接管理，负责建立和维护与 RabbitMQ 的连接
- **Producer**: 消息生产者，负责发送消息到队列
- **Consumer**: 消息消费者，负责从队列接收和处理消息
- **具体消费者实现**: 
  - `ActivityConsumer`: 处理活动订单
  - `OrderConsumer`: 处理订单超时
  - `StatusConsumer`: 处理订单状态变更

### 1.2 队列和交换机配置
| 交换机 | 队列 | 类型 | 用途 |
| :--- | :--- | :--- | :--- |
| `activity_exchange` | `activity_order_queue` | direct | 活动订单处理 |
| `order_status_exchange` | `order_status_queue` | fanout | 订单状态变更通知 |
| `dead_letter_exchange` | `order_dead_letter_queue` | direct | 普通订单超时处理（死信队列） |
| `dead_letter_exchange` | `activity_order_dead_letter_queue` | direct | 活动订单超时处理（死信队列） |
| - | `order_delay_queue` | - | 普通订单延迟队列（用于实现30分钟超时） |
| - | `activity_order_delay_queue` | - | 活动订单延迟队列（用于实现30分钟超时） |

## 2. 消息流程分析

### 2.1 活动订单流程

**生产者1**：`activity_order_controller.go`
- **触发时机**: 用户创建活动订单时
- **消息数据**: 
  ```json
  {
    "customer_id": 1,
    "activity_id": 1,
    "items": [
      {
        "product_id": 1,
        "sku_id": 1,
        "quantity": 2
      }
    ]
  }
  ```
- **发送方式**: `producer.Publish("activity_exchange", "activity_order", msg)`

**队列1**：`activity_order_queue`
- **绑定关系**: 绑定到 `activity_exchange` 交换机，路由键为 `activity_order`

**消费者1**：`ActivityConsumer.HandleActivityOrder`
- **处理逻辑**:
  1. 解析消息，提取 customer_id, activity_id 和 items
  2. 转换 items 为 `ActivityOrderItem` 结构
  3. 调用 `activityOrderService.CreateActivityOrder()` 创建活动订单
  4. 发送延迟消息到活动订单延迟队列
  5. 记录创建结果日志

**生产者2**：`ActivityConsumer.HandleActivityOrder`
- **触发时机**: 活动订单创建成功后
- **消息数据**: 
  ```json
  {
    "order_id": 1,
    "created_at": "2026-04-03T12:00:00Z"
  }
  ```
- **发送方式**: `producer.PublishWithTTL("", "activity_order_delay_queue", msg, 30*60*1000)` （30分钟超时）

**队列2**：
1. **延迟队列**: `activity_order_delay_queue`
   - 消息在此队列中等待30分钟
   - 配置了死信参数，超时后消息会被转发到 `dead_letter_exchange`

2. **死信队列**: `activity_order_dead_letter_queue`
   - 绑定到 `dead_letter_exchange` 交换机，路由键为 `dead_letter`
   - 接收从延迟队列超时的消息

**消费者2**：`ActivityConsumer.HandleTimeoutActivityOrder`
- **处理逻辑**:
  1. 解析消息，提取 order_id
  2. 调用 `activityOrderService.GetActivityOrderByID()` 获取活动订单
  3. 调用 `activityOrderService.CancelActivityOrder()` 取消活动订单（内部包含状态检查）
  4. 记录取消结果日志

### 2.2 订单超时流程

**生产者**：`order_controller.go`
- **触发时机**: 用户创建普通订单时
- **消息数据**: 
  ```json
  {
    "order_id": 1,
    "created_at": "2026-04-03T12:00:00Z"
  }
  ```
- **发送方式**: `producer.PublishWithTTL("", "order_delay_queue", msg, 30*60*1000)` （30分钟超时）

**队列**：
1. **延迟队列**: `order_delay_queue`
   - 消息在此队列中等待30分钟
   - 配置了死信参数，超时后消息会被转发到 `dead_letter_exchange`

2. **死信队列**: `order_dead_letter_queue`
   - 绑定到 `dead_letter_exchange` 交换机，路由键为 `dead_letter`
   - 接收从延迟队列超时的消息

**消费者**：`OrderConsumer.HandleTimeoutOrder`
- **处理逻辑**:
  1. 解析消息，提取 order_id
  2. 调用 `orderService.GetOrderByID()` 获取订单
  3. 调用 `orderService.CancelOrder()` 取消订单（内部包含状态检查和缓存清理）
  4. 记录取消结果日志

### 2.3 订单状态变更流程

**生产者**：`payment_controller.go`
- **触发时机**:
  1. 支付成功时（更新状态为 `paid`）
  2. 支付状态回调时（更新状态为回调中的状态）
- **消息数据**:
  ```json
  {
    "order_id": 1,
    "status": "paid",
    "updated_at": "2026-04-03T12:00:00Z"
  }
  ```
- **发送方式**: `producer.Publish("order_status_exchange", "order_status", msg)`

**队列**：`order_status_queue`
- **绑定关系**: 绑定到 `order_status_exchange` 交换机（fanout类型，路由键为空）

**消费者**：`StatusConsumer.HandleOrderStatus`
- **处理逻辑**:
  1. 解析消息，提取 order_id 和 status
  2. 记录状态变更日志
  3. 预留扩展点（如发送通知、更新缓存等）

## 3. 代码实现细节

### 3.1 连接管理
- **文件**: `pkg/mq/connection.go`
- **功能**: 管理与 RabbitMQ 的连接，支持重连机制
- **关键方法**: `NewConnection()`, `Reconnect()`

### 3.2 消息生产
- **文件**: `pkg/mq/producer.go`
- **功能**: 提供消息发布接口，支持普通消息和带TTL的延迟消息
- **关键方法**: `Publish()`, `PublishWithTTL()`

### 3.3 消息消费
- **文件**: `pkg/mq/consumer.go`
- **功能**: 提供消息消费接口，支持队列声明、绑定和消费
- **关键方法**: `Consume()`, `BindQueue()`, `DeclareQueue()`, `DeclareExchange()`

### 3.4 消费者初始化
- **文件**: `pkg/mq/consumer_init.go`
- **功能**: 初始化所有消费者，设置队列和交换机
- **关键步骤**:
  1. 创建连接和消费者
  2. 设置普通订单延迟队列
  3. 设置活动订单延迟队列
  4. 声明活动订单交换机和队列
  5. 声明订单状态交换机和队列
  6. 启动订单超时消费者
  7. 启动活动订单消费者
  8. 启动活动订单超时消费者
  9. 启动状态变更消费者

### 3.5 延迟队列实现
- **文件**: `pkg/mq/delay_queue.go`
- **功能**: 使用死信队列 + TTL 实现延迟功能
- **实现原理**:
  1. 声明死信交换机和死信队列
  2. 声明延迟队列，并配置死信参数
  3. 消息在延迟队列中等待TTL时间后，会被转发到死信队列
  4. 消费者从死信队列中获取消息进行处理

### 3.6 MQ常量定义
- **文件**: `constants/constants.go`
- **功能**: 定义MQ相关常量，避免硬编码
- **主要常量**:
  - 交换机名称：`MQExchangeActivity`, `MQExchangeOrderStatus`, `MQExchangeDeadLetter`
  - 队列名称：`MQQueueActivityOrder`, `MQQueueOrderStatus`, `MQQueueOrderDelay`, `MQQueueOrderDeadLetter`, `MQQueueActivityOrderDelay`, `MQQueueActivityOrderDeadLetter`
  - 路由键：`MQRoutingKeyActivityOrder`, `MQRoutingKeyOrderStatus`, `MQRoutingKeyDeadLetter`
  - 超时时间：`MQOrderTimeoutTTL` (30分钟)

## 4. 代码优化建议

### 4.1 错误处理优化
- **问题**: 消费者处理消息失败时，只是记录日志并重新入队，可能导致消息不断重试
- **建议**: 实现重试次数限制，超过次数后将消息放入死信队列或告警

### 4.2 消息幂等性
- **问题**: 消息可能会被重复消费，导致业务逻辑重复执行
- **建议**: 在消费者中实现幂等性检查，例如使用数据库唯一约束或缓存标记

### 4.3 监控和告警
- **问题**: 缺少对MQ的监控和告警机制
- **建议**: 添加MQ连接状态监控、消息积压监控和消费失败告警

### 4.4 代码结构优化
- **问题**: 消费者初始化代码与业务逻辑耦合
- **建议**: 将队列配置和消费者初始化分离，使用配置文件管理队列参数

## 5. 总结

GoShopAdmin 商城后端使用 RabbitMQ 实现了以下功能：

1. **活动订单异步处理**: 通过消息队列异步创建活动订单，提高系统响应速度
2. **订单超时自动取消**: 使用延迟队列实现30分钟订单超时自动取消功能，包括普通订单和活动订单
3. **订单状态变更通知**: 通过 fanout 交换机广播订单状态变更，支持多消费者处理
4. **统一的订单取消逻辑**: 所有订单取消操作使用相同的 `CancelOrder` 和 `CancelActivityOrder` 方法，确保逻辑一致
5. **避免硬编码**: 使用常量定义MQ相关配置，提高代码可维护性

整体设计合理，使用了标准的 MQ 模式，代码结构清晰，易于维护和扩展。通过添加活动订单延迟队列，确保了所有类型的订单都能自动处理超时情况，提高了系统的可靠性。
