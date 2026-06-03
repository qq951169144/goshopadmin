# GoShopAdmin DDD 示例项目架构文档

---

## 一、项目概览

本文档对比两种 Go DDD 实现方案，以"创建订单"接口为例，展示完整的代码运转流程和架构关系。

| 项目 | 路径 | 架构 | 端口 |
|:---|:---|:---|:---|
| 方案 B（单体 DDD） | `shop-backend-ddd/` | 按领域子包，共享 DB | 8090 |
| 方案 C（微服务 DDD） | `shop-micro-ddd/` | 按领域拆服务，独立 DB | 8090/8091/8092 |

---

## 二、方案 B 架构：单体 DDD（按领域子包）

### 2.1 分层架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        interfaces 层（HTTP）                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ OrderHandler │  │   Router     │  │   Middleware (Auth)   │  │
│  │ (Gin Handler)│  │ (路由映射)    │  │   (认证中间件)         │  │
│  └──────┬───────┘  └──────────────┘  └──────────────────────┘  │
│         │ 调用领域服务                                         │
└─────────┼───────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                         domain 层（纯 Go）                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐  │
│  │  order/ 聚合     │  │  product/ 聚合   │  │  customer/ 聚合│  │
│  │                 │  │                 │  │                │  │
│  │ Order 实体      │  │ Product 实体    │  │ Customer 实体  │  │
│  │ OrderItem 值对象│  │ SKU 值对象      │  │ Address 值对象 │  │
│  │ OrderService    │  │ ProductRepo接口 │  │ CustomerRepo  │  │
│  │ OrderRepo接口   │  │                 │  │   接口         │  │
│  │ ProductQuerier  │──│ (跨聚合查询接口) │──│ CustomerQuer. │  │
│  │ CustomerQuerier │  │                 │  │   接口         │  │
│  └─────────────────┘  └─────────────────┘  └────────────────┘  │
│         │ 依赖接口，不依赖具体实现                                 │
└─────────┼───────────────────────────────────────────────────────┘
          │ 接口由 infrastructure 层实现
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     infrastructure 层（技术实现）                 │
│  ┌──────────────────────┐  ┌────────────────────────────────┐  │
│  │  persistence/        │  │  cache/                        │  │
│  │                      │  │                                │  │
│  │ GormOrderRepository  │  │ CacheService                   │  │
│  │ GormProductRepository│  │ (Redis 缓存实现)               │  │
│  │ GormCustomerRepository│  │                                │  │
│  │ GormTransactionMgr   │  │                                │  │
│  │                      │  │                                │  │
│  │ models/ (PO 对象)    │  │                                │  │
│  │  ├ OrderPO           │  │                                │  │
│  │  ├ OrderItemPO       │  │                                │  │
│  │  ├ ProductPO         │  │                                │  │
│  │  └ CustomerPO        │  │                                │  │
│  └──────────────────────┘  └────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
          │
          ▼
   ┌──────────────┐
   │   MySQL DB   │
   └──────────────┘
```

### 2.2 依赖方向规则

```
interfaces ──→ domain ←── infrastructure

✅ 允许的依赖方向：
   interfaces  → domain（调用领域服务）
   infrastructure → domain（实现领域接口）

❌ 禁止的依赖方向：
   domain → interfaces（领域层不知道 HTTP 的存在）
   domain → infrastructure（领域层不知道 GORM 的存在）
   interfaces → infrastructure（通过接口解耦，不直接依赖）
```

### 2.3 创建订单流程（方案 B）

```
客户端                    OrderHandler              OrderService               GormProductRepo          GormOrderRepo
  │                          │                         │                          │                       │
  │  POST /api/orders        │                         │                          │                       │
  │  {address_id, items}     │                         │                          │                       │
  │─────────────────────────→│                         │                          │                       │
  │                          │                         │                          │                       │
  │                          │  绑定 DTO               │                          │                       │
  │                          │  转换为 ItemInput        │                          │                       │
  │                          │                         │                          │                       │
  │                          │  CreateOrder(ctx,1,1,items,"")                     │                       │
  │                          │────────────────────────→│                          │                       │
  │                          │                         │                          │                       │
  │                          │                         │  1. VerifyAddress(ctx,1,1)│                       │
  │                          │                         │──────────────────────────→│                       │
  │                          │                         │  ← merchantID=1          │                       │
  │                          │                         │                          │                       │
  │                          │                         │  2. BeginTx()            │                       │
  │                          │                         │──────┐                   │                       │
  │                          │                         │←─────┘ tx               │                       │
  │                          │                         │                          │                       │
  │                          │                         │  3. FindProductAndSKU()  │                       │
  │                          │                         │──────────────────────────→│                       │
  │                          │                         │  ← name,attrs,price,stock│                       │
  │                          │                         │                          │                       │
  │                          │                         │  4. DeductStockTx(tx,1,2)│                       │
  │                          │                         │──────────────────────────→│                       │
  │                          │                         │  ← ok                    │                       │
  │                          │                         │                          │                       │
  │                          │                         │  5. NewOrder(1,1,1,items,"")                      │
  │                          │                         │──────┐ (工厂方法，业务规则) │                       │
  │                          │                         │←─────┘ order实体         │                       │
  │                          │                         │                          │                       │
  │                          │                         │  6. Save(ctx, tx, order) │                       │
  │                          │                         │──────────────────────────────────────────────────→│
  │                          │                         │  ← ok                    │                       │
  │                          │                         │                          │                       │
  │                          │                         │  7. tx.Commit()          │                       │
  │                          │                         │──────┐                   │                       │
  │                          │                         │←─────┘                   │                       │
  │                          │                         │                          │                       │
  │                          │  ← order实体            │                          │                       │
  │                          │                         │                          │                       │
  │                          │  领域模型 → DTO         │                          │                       │
  │                          │                         │                          │                       │
  │  {code:0, data:{...}}    │                         │                          │                       │
  │←─────────────────────────│                         │                          │                       │
```

### 2.4 关键代码流转

#### Step 1: HTTP 入口 → DTO 绑定

```go
// interfaces/http/order_handler.go
func (h *OrderHandler) CreateOrder(ctx *gin.Context) {
    var req dto.CreateOrderRequest          // DTO
    ctx.ShouldBindJSON(&req)                // 绑定请求参数
    // DTO → 领域输入
    items := make([]order.ItemInput, 0)
    for _, item := range req.Items {
        items = append(items, order.ItemInput{...})
    }
    // 调用领域服务
    orderEntity, err := h.orderService.CreateOrder(ctx, customerID, req.AddressID, items, req.Remark)
}
```

#### Step 2: 领域服务编排

```go
// domain/order/service.go
func (s *OrderService) CreateOrder(...) (*Order, error) {
    // 1. 跨聚合查询：验证地址（通过接口，不知道具体实现）
    merchantID, err := s.customers.VerifyAddress(ctx, customerID, addressID)
    // 2. 开启事务
    tx, _ := s.txManager.BeginTx(ctx)
    // 3. 跨聚合查询：商品信息 + 扣库存
    name, attrs, price, stock, _ := s.products.FindProductAndSKU(ctx, productID, skuID)
    s.products.DeductStockTx(ctx, tx, skuID, quantity)
    // 4. 创建实体（业务规则在工厂方法中）
    order, _ := NewOrder(customerID, merchantID, addressID, orderItems, remark)
    // 5. 持久化
    s.orders.Save(ctx, tx, order)
    tx.Commit()
}
```

#### Step 3: 实体工厂方法（业务规则封装）

```go
// domain/order/order.go
func NewOrder(customerID, merchantID, addressID int, items []OrderItem, remark string) (*Order, error) {
    if customerID <= 0 { return nil, ErrInvalidCustomerID }  // 规则1：必须有客户
    if len(items) == 0 { return nil, ErrEmptyOrderItems }     // 规则2：必须有订单项
    // 计算总金额
    var total decimal.Decimal
    for _, item := range items { total = total.Add(item.TotalAmount()) }
    return &Order{status: StatusPending, totalAmount: total, ...}, nil
}
```

#### Step 4: 基础设施层实现接口

```go
// infrastructure/persistence/gorm_product_repo.go
// GormProductRepository 同时实现两个接口：
// - product.ProductRepository（商品聚合自己的接口）
// - order.ProductQuerier（订单聚合定义的跨聚合查询接口）
type GormProductRepository struct { db *gorm.DB }

func (r *GormProductRepository) FindProductAndSKU(...) (string, string, decimal.Decimal, int, error) {
    // 查询数据库，返回商品信息
}
func (r *GormProductRepository) DeductStockTx(ctx, tx, skuID, qty) error {
    // 在事务中扣减库存
}
```

#### Step 5: Composition Root（main.go 组装依赖）

```go
// cmd/server/main.go — 唯一知道所有具体实现的地方
func main() {
    db, _ := gorm.Open(mysql.Open(dsn))
    // 基础设施层
    orderRepo := persistence.NewGormOrderRepository(db)
    productRepo := persistence.NewGormProductRepository(db)  // 同时实现两个接口
    customerRepo := persistence.NewGormCustomerRepository(db)
    txManager := persistence.NewGormTransactionManager(db)
    // 领域层（注入接口，不知道具体实现）
    orderService := order.NewOrderService(orderRepo, productRepo, customerRepo, txManager)
    // 接口层
    handler := http.NewOrderHandler(orderService)
}
```

### 2.5 跨聚合调用关系

```
┌─────────────────────────────────────────────────────────────┐
│                    OrderService（编排者）                     │
│                                                             │
│  依赖接口：                                                  │
│  ├── OrderRepository      ──→ GormOrderRepository           │
│  ├── ProductQuerier       ──→ GormProductRepository         │
│  ├── CustomerQuerier      ──→ GormCustomerRepository        │
│  └── TransactionManager   ──→ GormTransactionManager        │
│                                                             │
│  注意：ProductQuerier 和 ProductRepository                  │
│       由同一个 GormProductRepository 实现                    │
│       但接口由不同的聚合定义                                   │
└─────────────────────────────────────────────────────────────┘
```

### 2.6 单元测试策略

```
┌──────────────────────────────────────────────────────┐
│              service_test.go（纯内存测试）              │
│                                                      │
│  OrderService                                        │
│  ├── MockOrderRepository     (实现 OrderRepository)   │
│  ├── MockProductQuerier      (实现 ProductQuerier)    │
│  ├── MockCustomerQuerier     (实现 CustomerQuerier)   │
│  └── MockTransactionManager  (实现 TransactionManager)│
│                                                      │
│  不需要 MySQL、Redis、HTTP 服务器                      │
│  测试速度：毫秒级                                     │
│  5 个测试用例全部通过 ✅                               │
└──────────────────────────────────────────────────────┘
```

---

## 三、方案 C 架构：微服务 DDD

### 3.1 服务拓扑图

```
                    ┌─────────────────────┐
                    │     客户端           │
                    └─────────┬───────────┘
                              │
                    ┌─────────▼───────────┐
                    │   API Gateway       │
                    │   (Nginx/Traefik)   │
                    └──┬──────┬──────┬────┘
                       │      │      │
            ┌──────────▼┐  ┌──▼────┐ ┌▼──────────┐
            │ order-svc │  │product│ │customer   │
            │  :8090    │  │-svc   │ │-svc       │
            │           │  │:8091  │ │:8092      │
            │           │  │       │ │           │
            │  创建订单  │──│ 查商品 │ │ 验证地址   │
            │  扣减库存  │──│ 扣库存 │ │           │
            │  恢复库存  │──│ 恢复   │ │           │
            └─────┬─────┘  └───┬───┘ └─────┬─────┘
                  │            │           │
            ┌─────▼─────┐ ┌───▼────┐ ┌────▼─────┐
            │ order DB  │ │product │ │customer  │
            │           │ │  DB    │ │   DB     │
            └───────────┘ └────────┘ └──────────┘
```

### 3.2 跨服务调用方式

```
方案 B（单体）：跨聚合通过接口直接调用
┌──────────┐    ProductQuerier 接口    ┌──────────────────┐
│  Order   │──────────────────────────→│ GormProductRepo  │
│  Service │    (进程内方法调用)         │ (同一个进程)      │
└──────────┘                           └──────────────────┘

方案 C（微服务）：跨服务通过 HTTP 调用
┌──────────┐  ProductServiceProvider  ┌──────────────────┐    HTTP     ┌──────────────────┐
│  Order   │────────────────────────→│ ProductHTTPClient │───────────→│  Product Service │
│  Service │   (接口定义在 domain)     │ (infrastructure)  │  REST API  │  (另一个进程)     │
└──────────┘                          └──────────────────┘            └──────────────────┘
```

### 3.3 创建订单流程（方案 C）

```
客户端     OrderHandler   OrderService   CustomerHTTP   ProductHTTP   GormOrderRepo
  │            │              │            Client         Client          │
  │ POST       │              │              │              │             │
  │───────────→│              │              │              │             │
  │            │ CreateOrder  │              │              │             │
  │            │─────────────→│              │              │             │
  │            │              │              │              │             │
  │            │              │ 1. VerifyAddress             │             │
  │            │              │─────────────→│              │             │
  │            │              │              │──HTTP GET──→│             │
  │            │              │              │  customer-svc│             │
  │            │              │←─────────────│              │             │
  │            │              │ merchantID=1 │              │             │
  │            │              │              │              │             │
  │            │              │ 2. GetProductAndSKU          │             │
  │            │              │─────────────────────────────→│             │
  │            │              │                              │──HTTP GET→│
  │            │              │                              │product-svc │
  │            │              │←─────────────────────────────│             │
  │            │              │ name,price,stock             │             │
  │            │              │              │              │             │
  │            │              │ 3. DeductStock              │             │
  │            │              │─────────────────────────────→│             │
  │            │              │                              │─HTTP POST→│
  │            │              │                              │product-svc │
  │            │              │←─────────────────────────────│             │
  │            │              │ ok           │              │             │
  │            │              │              │              │             │
  │            │              │ 4. NewOrder (本地)           │             │
  │            │              │──────┐       │              │             │
  │            │              │←─────┘       │              │             │
  │            │              │              │              │             │
  │            │              │ 5. BeginTx + Save (本地事务) │             │
  │            │              │──────────────────────────────────────────→│
  │            │              │              │              │             │
  │            │              │ 6. Commit    │              │             │
  │            │              │──────┐       │              │             │
  │            │              │←─────┘       │              │             │
  │            │              │              │              │             │
  │            │ ← order      │              │              │             │
  │ ← response │              │              │              │             │
```

### 3.4 Saga 补偿模式（方案 C 独有）

微服务中，跨服务事务无法用数据库事务保证。方案 C 使用 Saga 补偿模式：

```
创建订单流程（正常路径）：
  1. 验证地址 ✅
  2. 查询商品 ✅
  3. 扣减库存 ✅  ← HTTP 调用 product-service
  4. 创建订单 ✅  ← 本地事务
  5. 返回成功

创建订单流程（异常路径 - 步骤4失败）：
  1. 验证地址 ✅
  2. 查询商品 ✅
  3. 扣减库存 ✅  ← 已执行
  4. 创建订单 ❌  ← 失败！
  5. 补偿：恢复库存 ← HTTP 调用 product-service RestoreStock
  6. 返回失败

关键代码：
  // domain/order/service.go
  deductedSKUs := []struct{ skuID, qty int }{}  // 记录已扣减的 SKU
  for _, input := range items {
      s.products.DeductStock(ctx, skuID, qty)   // 扣减
      deductedSKUs = append(deductedSKUs, ...)   // 记录
  }
  order, err := NewOrder(...)
  if err != nil {
      s.compensateDeduction(ctx, deductedSKUs)   // 补偿
      return nil, err
  }
```

### 3.5 内部 API 契约

```
product-service 提供的内部 API：
┌──────────────────────────────────────────────────────────────┐
│ GET  /api/internal/products/:productID/skus/:skuID           │
│      → 返回 {name, sku_attrs, price, stock}                 │
│                                                              │
│ POST /api/internal/skus/:skuID/deduct   body: {quantity: N} │
│      → 扣减库存                                              │
│                                                              │
│ POST /api/internal/skus/:skuID/restore   body: {quantity: N} │
│      → 恢复库存                                              │
└──────────────────────────────────────────────────────────────┘

customer-service 提供的内部 API：
┌──────────────────────────────────────────────────────────────┐
│ GET /api/internal/customers/:customerID/addresses/:addressID/verify │
│     → 返回 {merchant_id: N}                                  │
└──────────────────────────────────────────────────────────────┘
```

---

## 四、两种方案对比

### 4.1 架构对比

```
方案 B（单体 DDD）：
┌─────────────────────────────────────────┐
│              单个进程 (:8090)             │
│                                         │
│  interfaces ──→ domain ──→ infrastructure│
│                                         │
│  跨聚合：接口直接调用（进程内）            │
│  事务：一个 DB 事务搞定                   │
│  部署：一个 Docker 容器                   │
└─────────────────────────────────────────┘

方案 C（微服务 DDD）：
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ order-svc    │ │ product-svc  │ │ customer-svc │
│   :8090      │ │   :8091      │ │   :8092      │
│              │ │              │ │              │
│ 跨服务：HTTP │→│ HTTP API     │ │ HTTP API     │
│ 事务：Saga   │ │              │ │              │
│ 部署：3 容器 │ │              │ │              │
└──────────────┘ └──────────────┘ └──────────────┘
```

### 4.2 代码对比

| 维度 | 方案 B（单体） | 方案 C（微服务） |
|:---|:---|:---|
| 跨聚合调用 | `ProductQuerier` 接口，进程内方法调用 | `ProductServiceProvider` 接口，HTTP 客户端实现 |
| 事务 | `TransactionManager` + 单个 DB 事务 | Saga 补偿模式，本地事务 + 补偿操作 |
| 依赖注入 | `GormProductRepository` 同时实现两个接口 | `ProductHTTPClient` 实现服务提供者接口 |
| 测试 Mock | Mock Repository 接口 | Mock HTTP 客户端接口 |
| 部署 | 1 个容器 | 3 个容器 + docker-compose |
| 代码量 | ~18 个文件 | ~30+ 个文件 |
| 复杂度 | 中 | 高 |

### 4.3 选择建议

| 场景 | 推荐方案 |
|:---|:---|
| 团队 1-5 人 | 方案 B |
| 团队 5+ 人，各模块独立迭代 | 方案 C |
| 日活 < 10 万 | 方案 B |
| 日活 > 10 万，需要独立扩缩容 | 方案 C |
| 快速验证业务 | 方案 B |
| 业务成熟，需要技术解耦 | 方案 C |

---

## 五、DDD 核心概念映射

### 5.1 概念 → 代码映射表

| DDD 概念 | 方案 B 代码位置 | 方案 C 代码位置 |
|:---|:---|:---|
| 实体 (Entity) | `domain/order/order.go` — Order | 同左 |
| 值对象 (Value Object) | `domain/order/order_item.go` — OrderItem | 同左 |
| 聚合根 (Aggregate Root) | `domain/order/` 包 | order-service |
| 仓库接口 (Repository) | `domain/order/repository.go` | 同左 |
| 仓库实现 (Repository Impl) | `infrastructure/persistence/gorm_order_repo.go` | 同左 |
| 领域服务 (Domain Service) | `domain/order/service.go` — OrderService | 同左 |
| 领域事件 (Domain Event) | 未实现（可扩展） | Saga 补偿替代 |
| 限界上下文 (Bounded Context) | `domain/order/`、`domain/product/` 子包 | 独立微服务 |
| 应用服务 (Application Service) | `interfaces/http/order_handler.go` | 同左 |
| DTO | `interfaces/dto/order_dto.go` | 同左 |
| PO | `infrastructure/persistence/models/` | 同左 |
| Composition Root | `cmd/server/main.go` | 各服务 main.go |

### 5.2 充血模型 vs 贫血模型

```
贫血模型（原项目 shop-backend）：
  models.Order → 纯数据结构，没有行为
  services.OrderService.CancelOrder() → 业务规则在 Service 里
  if order.Status != "pending" { ... } → 规则散落

充血模型（DDD 项目）：
  domain.Order → 实体，自带业务方法
  order.Cancel() → 业务规则在实体内部
  if o.status != StatusPending { return ErrInvalidStatusForCancel }
  → 规则封装，外部只需调用方法
```

---

## 六、文件索引

### 方案 B：shop-backend-ddd/

```
shop-backend-ddd/
├── cmd/server/main.go                          # 入口（Composition Root）
├── config/config.go                            # 配置
├── domain/
│   ├── order/
│   │   ├── errors.go                           # 领域错误
│   │   ├── order.go                            # 订单实体（充血模型）
│   │   ├── order_item.go                       # 订单项值对象
│   │   ├── repository.go                       # 仓库接口 + 事务接口
│   │   ├── service.go                          # 领域服务
│   │   └── service_test.go                     # 单元测试（5个用例）
│   ├── product/
│   │   ├── product.go                          # 商品实体 + SKU 值对象
│   │   └── repository.go                       # 商品仓库接口
│   └── customer/
│       ├── customer.go                         # 客户实体 + 地址值对象
│       └── repository.go                       # 客户仓库接口
├── infrastructure/
│   ├── persistence/
│   │   ├── models/
│   │   │   ├── order_po.go                     # 订单 PO
│   │   │   ├── product_po.go                   # 商品 PO
│   │   │   └── customer_po.go                  # 客户 PO
│   │   ├── gorm_order_repo.go                  # GORM 订单仓库
│   │   ├── gorm_product_repo.go                # GORM 商品仓库（含 ProductQuerier 适配）
│   │   ├── gorm_customer_repo.go               # GORM 客户仓库（含 CustomerQuerier 适配）
│   │   └── gorm_transaction.go                 # GORM 事务管理器
│   └── cache/
│       └── cache_service.go                    # 缓存服务
├── interfaces/
│   ├── dto/
│   │   └── order_dto.go                        # 请求/响应 DTO
│   └── http/
│       ├── order_handler.go                    # HTTP Handler
│       ├── router.go                           # 路由配置
│       └── middleware.go                       # 认证中间件
└── go.mod
```

### 方案 C：shop-micro-ddd/

```
shop-micro-ddd/
├── services/
│   ├── order-service/                          # 订单微服务 (:8090)
│   │   ├── cmd/server/main.go
│   │   ├── domain/order/                       # 充血模型 + 领域服务 + 测试
│   │   ├── infrastructure/
│   │   │   ├── persistence/                    # GORM 实现
│   │   │   └── client/                         # HTTP 客户端（跨服务调用）
│   │   ├── interfaces/http/                    # HTTP Handler
│   │   └── config/
│   ├── product-service/                        # 商品微服务 (:8091)
│   │   ├── cmd/server/main.go
│   │   ├── domain/product/                     # 商品实体 + SKU
│   │   ├── infrastructure/persistence/         # GORM 实现
│   │   └── interfaces/http/                    # 外部 API + 内部 API
│   └── customer-service/                       # 客户微服务 (:8092)
│       ├── cmd/server/main.go
│       ├── domain/customer/                    # 客户实体 + 地址
│       ├── infrastructure/persistence/         # GORM 实现
│       └── interfaces/http/                    # 外部 API + 内部 API
└── docker-compose.yml                          # 本地开发编排
```
