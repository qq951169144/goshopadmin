# GoShopAdmin 项目结构分析：Go 代码的"Java 味"与 Go 原生风格对比

---

## 一、现状诊断：你的 Go 代码为什么一股 Java 味

### 1.1 典型 Java 症状清单

| 症状 | 项目中的表现 | Java 来源 |
|:---|:---|:---|
| Controller-Service 三层架构 | `controllers/` → `services/` → `models/`，每层严格分离 | Spring MVC |
| BaseController 继承 | 所有 Controller 嵌入 `BaseController` 获取 `ResponseSuccess/ResponseError` | Spring `BaseController` 或 `BaseAction` |
| NewXxxController 工厂函数 | `controllers.NewOrderController(db, cacheUtil)` | Spring `@Autowired` 构造器注入 |
| Service 持有 DB 连接 | `OrderService{db: *gorm.DB, cacheUtil: *cache.CacheUtil}` | Spring `@Service` 注入 `@Autowired DataSource` |
| 全局错误码枚举 | `errors/code.go` 定义 `CodeParamError = 4001`，配合 `ErrorMessage` map | Java `ErrorCode` 枚举类 |
| Dependencies 结构体 | `routes.Dependencies` 聚合所有 Controller | Spring IoC 容器的翻版 |
| Service 直接操作 ORM | `s.db.Where(...).First(&order)` 散布在 Service 方法中 | MyBatis-Plus / JPA Repository 内联 |

### 1.2 具体代码味道

#### 味道 1：Service 是 DB 的薄包装

```go
// shop-backend/services/order_service.go — 典型的"贫血模型"
func (s *OrderService) CancelOrder(orderNo string, customerID int) error {
    tx := s.db.Begin()
    var order models.Order
    result := tx.Where("order_no = ? AND customer_id = ?", orderNo, customerID).First(&order)
    // ... 一堆数据库操作 ...
    return tx.Commit().Error
}
```

**问题**：`OrderService` 不是"订单服务"，而是"订单数据库操作集合"。业务逻辑（状态校验、库存返还）和数据访问（SQL 查询）混在一起。这是 Java Service 层的典型写法——Service 只是个事务脚本。

#### 味道 2：Controller 做参数搬运工

```go
// shop-backend/controllers/order_controller.go
func (c *OrderController) CreateOrder(ctx *gin.Context) {
    var req CreateOrderRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ResponseError(ctx, errors.CodeParamInvalid, err)
        return
    }
    customerID, exists := ctx.Get("customer_id")
    // ... 把 req 字段一个个搬到 services.OrderItemRequest ...
    var orderItems []services.OrderItemRequest
    for _, item := range req.Items {
        orderItems = append(orderItems, services.OrderItemRequest{
            ProductID: item.ProductID,
            SkuID:     item.SkuID,
            Quantity:  item.Quantity,
        })
    }
    order, err := c.orderService.CreateOrder(services.CreateOrderRequest{...})
    // ...
}
```

**问题**：Controller 和 Service 各自定义了几乎一样的 Request 结构体，Controller 的唯一作用就是搬运字段。这是 Java DTO 翻转的典型场景——每层都有自己的 DTO，然后写一堆 BeanUtils.copyProperties。

#### 味道 3：全局变量充当单例

```go
// search-service/services/es_client.go
var (
    esClient  *elastic.Client
    db        *gorm.DB
    clientOnce sync.Once
    dbOnce     sync.Once
)

func GetESClient() *elastic.Client { return esClient }
func GetDB() *gorm.DB              { return db }
```

**问题**：用包级全局变量 + `sync.Once` 模拟 Java 的单例模式。Service 层通过 `services.SearchProducts(params)` 这种包级函数调用，隐式依赖全局 ES 客户端。测试时无法替换，并发时无法控制生命周期。

#### 味道 4：search-service 的 Service 是纯函数包

```go
// search-service/services/product_service.go
func SearchProducts(params ProductSearchParams) (*models.SearchResult, error) {
    client := GetESClient()  // 从全局变量拿
    // ...
}
```

**问题**：这不是面向对象的 Service，而是一个包级函数集合。它依赖全局状态（`GetESClient()`），无法注入、无法测试、无法替换。这是 Java 开发者写 Go 时最常见的妥协——既想要 Service 的概念，又不想写 struct，结果变成了四不像。

### 1.3 Java 味的根源

| 根源 | 说明 |
|:---|:---|
| 思维惯性 | 从 Java/Spring 转来，习惯性地按 Controller → Service → DAO 分层 |
| 缺少接口抽象 | Go 的接口是隐式的、小而精的，Java 的接口是显式的、大而全的 |
| 过度封装 | 为了"统一"而抽象出 BaseController、统一错误码，但牺牲了灵活性 |
| 忽视组合 | Go 倡导组合优于继承，但项目里大量使用嵌入 BaseController 的"伪继承" |

---

## 二、Go 崇尚的领域驱动设计（DDD）到底是什么

### 2.1 DDD 不是 Java 专属

DDD（Domain-Driven Design）是 Eric Evans 在 2003 年提出的**思想方法论**，不是框架，不绑定任何语言。它的核心是：

> **让代码结构反映业务领域，而不是技术架构。**

Java 社区把 DDD 做成了框架（Spring + DDD 脚手架），导致很多人以为 DDD = 分层 + 注解 + 领域事件。这是误解。

### 2.2 DDD 的核心概念

```
┌─────────────────────────────────────────────────┐
│                  领域层 (Domain)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │ 实体      │  │ 值对象    │  │ 领域服务      │  │
│  │ Entity   │  │ ValueObj │  │ DomainSvc    │  │
│  └──────────┘  └──────────┘  └──────────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │ 聚合根    │  │ 领域事件  │  │ 仓库接口      │  │
│  │ Aggregate│  │ Event    │  │ Repository   │  │
│  └──────────┘  └──────────┘  └──────────────┘  │
└─────────────────────────────────────────────────┘
```

| 概念 | 含义 | GoShopAdmin 对应 |
|:---|:---|:---|
| **实体 (Entity)** | 有唯一标识的对象，生命周期跨越多个操作 | `models.Order`（但现在是贫血的） |
| **值对象 (Value Object)** | 无唯一标识，通过属性判等 | `models.Address`（应该是值对象） |
| **聚合根 (Aggregate Root)** | 一致性边界，外部只能通过聚合根操作内部实体 | 不存在——Order 和 OrderItem 是分开操作的 |
| **领域服务 (Domain Service)** | 不属于任何实体的业务逻辑 | 不存在——所有逻辑在 Service 层 |
| **仓库 (Repository)** | 领域层与基础设施的桥梁，提供类似集合的接口 | 不存在——Service 直接用 `gorm.DB` |
| **领域事件 (Domain Event)** | 领域内发生的有业务意义的事件 | 不存在——MQ 消息是技术实现，不是领域事件 |

### 2.3 当前项目的"伪 DDD"

```
当前结构（贫血模型 + 事务脚本）：

Controller ──→ Service（事务脚本）──→ gorm.DB ──→ MySQL
    │               │
    │               └── 直接写 SQL / ORM 调用
    │               └── 直接操作 Redis
    │               └── 直接发 MQ 消息
    └── 参数校验 + 字段搬运
    └── 统一响应包装
```

```
DDD 理想结构（充血模型）：

Controller ──→ Application Service ──→ Domain Model
                    │                      │
                    │                      ├── Entity（自带业务方法）
                    │                      ├── Value Object
                    │                      └── Domain Event
                    │
                    └──→ Repository（接口）──→ Infrastructure（实现）
                    └──→ Domain Service（跨实体逻辑）
```

### 2.4 用订单举例：贫血 vs 充血

#### 贫血模型（当前写法）

```go
// models/order.go — 只是个数据结构，没有行为
type Order struct {
    ID     int    `json:"id"`
    Status string `json:"status"`
    // ... 纯数据字段 ...
}

// services/order_service.go — 行为全在 Service 里
func (s *OrderService) CancelOrder(orderNo string, customerID int) error {
    var order models.Order
    s.db.Where("order_no = ?", orderNo).First(&order)

    // 业务规则写在 Service 里
    if order.Status != "pending" && order.Status != "paid" {
        return errors.New("当前订单状态不允许取消")
    }

    order.Status = "cancelled"
    s.db.Save(&order)
    return nil
}
```

#### 充血模型（DDD 写法）

```go
// domain/order.go — 实体自带业务方法
type Order struct {
    id         int
    orderNo    string
    status     OrderStatus
    items      []OrderItem
    // 私有字段，外部不能随意修改
}

// Cancel 取消订单 — 业务规则在实体内部
func (o *Order) Cancel() error {
    if o.status != StatusPending && o.status != StatusPaid {
        return ErrInvalidStatusForCancel
    }
    o.status = StatusCancelled
    o.cancelledAt = time.Now()
    // 返还库存是领域事件，不是直接操作 DB
    o.events = append(o.events, OrderCancelledEvent{OrderID: o.id, Items: o.items})
    return nil
}

// CanCancel 判断是否可取消 — 规则也属于实体
func (o *Order) CanCancel() bool {
    return o.status == StatusPending || o.status == StatusPaid
}
```

**关键区别**：
- 贫血模型：`Order` 是数据袋，规则散落在 Service
- 充血模型：`Order` 是活的，规则封装在实体内部，Service 只做编排

### 2.5 Go 的 DDD 不需要框架

Go 的 DDD 实现比 Java 简单得多，因为：

1. **不需要 Spring**：Go 的接口是隐式满足的，不需要 `@Autowired`
2. **不需要 Hibernate**：Go 的 `database/sql` 已经很轻量，Repository 模式用接口包一层即可
3. **不需要事件总线**：Go 的 channel 就是天然的事件机制
4. **不需要 DTO 库**：Go 的结构体天然就是值对象

---

## 三、Go 的依赖注入该怎么写

### 3.1 当前项目的"伪注入"

```go
// shop-backend/routes/routes.go
func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cfg *config.Config, monitor *utils.Monitor) {
    deps := &Dependencies{
        OrderController: controllers.NewOrderController(db, cacheUtil),
        // ...
    }
}
```

**问题**：
1. `SetupRoutes` 承担了组装所有依赖的职责，函数签名越来越长
2. Controller 内部直接 `NewOrderService(db, cacheUtil)`，依赖是硬编码的
3. 没有接口，无法替换实现，无法测试

### 3.2 Go 的依赖注入：接口 + 构造器

Go 的依赖注入不需要 wire/dig 等框架，核心就两步：

1. **定义小接口**（消费者定义接口，不是提供者）
2. **构造器注入**（通过参数传入依赖）

#### Step 1：定义接口

```go
// domain/repository.go — 领域层定义它需要的接口
type OrderRepository interface {
    FindByOrderNo(orderNo string) (*Order, error)
    Save(order *Order) error
    BeginTx() (*gorm.DB, error)
}

type StockRepository interface {
    Deduct(skuID int, quantity int) error
    Restore(skuID int, quantity int) error
}

type CacheService interface {
    DeleteOrderCache(orderNo string) error
}
```

**关键原则**：接口由消费者（领域层）定义，不是由提供者（基础设施层）定义。接口要小，一个接口只做一件事。

#### Step 2：构造器注入

```go
// domain/order_service.go — 领域服务
type OrderService struct {
    orders OrderRepository
    stock  StockRepository
    cache  CacheService
}

func NewOrderService(orders OrderRepository, stock StockRepository, cache CacheService) *OrderService {
    return &OrderService{
        orders: orders,
        stock:  stock,
        cache:  cache,
    }
}

func (s *OrderService) CancelOrder(orderNo string, customerID int) error {
    order, err := s.orders.FindByOrderNo(orderNo)
    if err != nil {
        return err
    }

    // 业务规则在实体内部
    if err := order.Cancel(); err != nil {
        return err
    }

    // 持久化
    if err := s.orders.Save(order); err != nil {
        return err
    }

    // 清缓存
    s.cache.DeleteOrderCache(orderNo)
    return nil
}
```

#### Step 3：基础设施层实现接口

```go
// infrastructure/order_repository.go
type GormOrderRepository struct {
    db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) domain.OrderRepository {
    return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) FindByOrderNo(orderNo string) (*domain.Order, error) {
    var po OrderPO // 持久化对象
    if err := r.db.Where("order_no = ?", orderNo).First(&po).Error; err != nil {
        return nil, err
    }
    return po.ToDomain(), nil // PO → Domain 实体
}

func (r *GormOrderRepository) Save(order *domain.Order) error {
    po := OrderPOFromDomain(order) // Domain 实体 → PO
    return r.db.Save(&po).Error
}
```

#### Step 4：组装（main.go 或专门的 wire 函数）

```go
// main.go — 手动组装，清晰可控
func main() {
    db := config.InitDB(cfg)
    redisClient := config.InitRedis(cfg)

    // 基础设施层
    orderRepo := infrastructure.NewGormOrderRepository(db)
    stockRepo := infrastructure.NewGormStockRepository(db)
    cacheSvc := infrastructure.NewRedisCacheService(redisClient)

    // 领域层
    orderService := domain.NewOrderService(orderRepo, stockRepo, cacheSvc)

    // 表现层
    orderController := controllers.NewOrderController(orderService)

    // 路由
    r := gin.New()
    r.POST("/api/orders", orderController.CreateOrder)
    r.Run(":8080")
}
```

### 3.3 要不要用 wire/dig？

| 方案 | 优点 | 缺点 | 推荐 |
|:---|:---|:---|:---|
| 手动组装 | 零依赖、完全可控、编译期检查 | 项目大了 main.go 会长 | 小项目首选 |
| google/wire | 编译期代码生成、类型安全 | 学习成本、生成代码需维护 | 中大型项目 |
| uber/dig | 运行时注入、灵活 | 运行时错误、调试困难 | 不推荐 |

**建议**：当前项目规模，手动组装完全够用。当 main.go 超过 200 行时再考虑 wire。

### 3.4 对比总结

```
当前写法（Java 味）：
  Controller ──→ NewXxxService(db, cache) ──→ gorm.DB 直接操作
  依赖是具体的 struct，无法替换，无法测试

Go 原生写法：
  Controller ──→ Service(接口) ──→ Repository(接口) ──→ gorm.DB
  依赖是接口，可以替换实现，可以 mock 测试
```

---

## 四、单元测试该怎么写

### 4.1 当前项目为什么写不了单元测试

```go
// 当前 OrderService 的依赖链
OrderService {
    db: *gorm.DB           // 具体实现，无法 mock
    cacheUtil: *cache.CacheUtil  // 具体实现，无法 mock
}

// 调用方式
order, err := c.orderService.CreateOrder(services.CreateOrderRequest{...})
```

要测试 `CreateOrder`，你必须：
1. 启动一个真实的 MySQL
2. 启动一个真实的 Redis
3. 准备测试数据

这不是单元测试，这是集成测试。根本原因：**依赖是具体的，不是接口**。

### 4.2 正确的测试姿势：以订单创建为例

#### 定义接口

```go
// domain/repository.go
type OrderRepository interface {
    FindByOrderNo(orderNo string) (*Order, error)
    FindByID(id int) (*Order, error)
    Save(order *Order) error
    BeginTx() (Transaction, error)
}

type ProductRepository interface {
    FindByID(id int) (*Product, error)
    LockForUpdate(tx Transaction, id int) (*Product, error)
}

type SKURepository interface {
    FindByID(id int) (*SKU, error)
    DeductStock(tx Transaction, skuID int, quantity int) error
    RestoreStock(tx Transaction, skuID int, quantity int) error
}

type AddressRepository interface {
    FindByIDAndCustomer(id int, customerID int) (*Address, error)
}

type CacheService interface {
    DeleteOrderCache(orderNo string) error
    AddOrderToBloomFilter(orderNo string) error
}

type EventPublisher interface {
    PublishOrderCreated(order *Order) error
}
```

#### 领域服务实现

```go
// domain/order_service.go
type OrderService struct {
    orders   OrderRepository
    products ProductRepository
    skus     SKURepository
    addresses AddressRepository
    cache    CacheService
    events   EventPublisher
}

func NewOrderService(
    orders OrderRepository,
    products ProductRepository,
    skus SKURepository,
    addresses AddressRepository,
    cache CacheService,
    events EventPublisher,
) *OrderService {
    return &OrderService{
        orders:   orders,
        products: products,
        skus:     skus,
        addresses: addresses,
        cache:    cache,
        events:   events,
    }
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(customerID int, addressID int, items []OrderItemInput, remark string) (*Order, error) {
    // 1. 验证地址
    addr, err := s.addresses.FindByIDAndCustomer(addressID, customerID)
    if err != nil {
        return nil, ErrAddressNotFound
    }

    // 2. 开启事务
    tx, err := s.orders.BeginTx()
    if err != nil {
        return nil, err
    }

    // 3. 验证商品和 SKU，计算金额
    var totalAmount decimal.Decimal
    orderItems := make([]OrderItem, 0, len(items))

    for _, item := range items {
        product, err := s.products.FindByID(item.ProductID)
        if err != nil {
            tx.Rollback()
            return nil, ErrProductNotFound
        }

        sku, err := s.skus.FindByID(item.SkuID)
        if err != nil {
            tx.Rollback()
            return nil, ErrSKUNotFound
        }

        if sku.Stock < item.Quantity {
            tx.Rollback()
            return nil, ErrStockInsufficient
        }

        if err := s.skus.DeductStock(tx, item.SkuID, item.Quantity); err != nil {
            tx.Rollback()
            return nil, err
        }

        itemAmount := sku.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
        totalAmount = totalAmount.Add(itemAmount)

        orderItems = append(orderItems, OrderItem{
            ProductID:   product.ID,
            SkuID:       sku.ID,
            ProductName: product.Name,
            Price:       sku.Price,
            Quantity:    item.Quantity,
            TotalAmount: itemAmount,
        })
    }

    // 4. 创建订单实体（充血模型）
    order, err := NewOrder(customerID, addr.MerchantID, totalAmount, addressID, orderItems, remark)
    if err != nil {
        tx.Rollback()
        return nil, err
    }

    // 5. 持久化
    if err := s.orders.Save(order); err != nil {
        tx.Rollback()
        return nil, err
    }

    tx.Commit()

    // 6. 后置操作（异步）
    s.cache.AddOrderToBloomFilter(order.OrderNo)
    s.events.PublishOrderCreated(order)

    return order, nil
}
```

#### 编写 Mock（手写，不需要 mockgen）

```go
// domain/mocks/mock_repositories.go
package mocks

type MockOrderRepository struct {
    SaveFunc      func(order *domain.Order) error
    FindByOrderNoFunc func(orderNo string) (*domain.Order, error)
    BeginTxFunc   func() (domain.Transaction, error)
}

func (m *MockOrderRepository) Save(order *domain.Order) error {
    if m.SaveFunc != nil {
        return m.SaveFunc(order)
    }
    return nil
}

func (m *MockOrderRepository) FindByOrderNo(orderNo string) (*domain.Order, error) {
    if m.FindByOrderNoFunc != nil {
        return m.FindByOrderNoFunc(orderNo)
    }
    return nil, nil
}

func (m *MockOrderRepository) BeginTx() (domain.Transaction, error) {
    if m.BeginTxFunc != nil {
        return m.BeginTxFunc()
    }
    return &MockTransaction{}, nil
}

type MockTransaction struct {
    committed bool
    rolledBack bool
}

func (t *MockTransaction) Commit() error   { t.committed = true; return nil }
func (t *MockTransaction) Rollback() error { t.rolledBack = true; return nil }

type MockSKURepository struct {
    FindByIDFunc     func(id int) (*domain.SKU, error)
    DeductStockFunc  func(tx domain.Transaction, skuID int, quantity int) error
    RestoreStockFunc func(tx domain.Transaction, skuID int, quantity int) error
}

func (m *MockSKURepository) FindByID(id int) (*domain.SKU, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(id)
    }
    return &domain.SKU{ID: id, Stock: 100, Price: decimal.NewFromFloat(99.9)}, nil
}

func (m *MockSKURepository) DeductStock(tx domain.Transaction, skuID int, quantity int) error {
    if m.DeductStockFunc != nil {
        return m.DeductStockFunc(tx, skuID, quantity)
    }
    return nil
}

func (m *MockSKURepository) RestoreStock(tx domain.Transaction, skuID int, quantity int) error {
    if m.RestoreStockFunc != nil {
        return m.RestoreStockFunc(tx, skuID, quantity)
    }
    return nil
}

// ... 其他 Mock 类似 ...
```

#### 编写单元测试

```go
// domain/order_service_test.go
package domain

import (
    "testing"

    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "goshopadmin/domain/mocks"
)

func TestOrderService_CreateOrder_Success(t *testing.T) {
    // Arrange — 准备依赖
    mockOrders := &mocks.MockOrderRepository{
        SaveFunc: func(order *Order) error {
            assert.Equal(t, StatusPending, order.Status)
            return nil
        },
    }

    mockSKUs := &mocks.MockSKURepository{
        FindByIDFunc: func(id int) (*SKU, error) {
            return &SKU{
                ID:    1,
                Stock: 100,
                Price: decimal.NewFromFloat(99.90),
            }, nil
        },
        DeductStockFunc: func(tx Transaction, skuID int, quantity int) error {
            assert.Equal(t, 1, skuID)
            assert.Equal(t, 2, quantity)
            return nil
        },
    }

    mockProducts := &mocks.MockProductRepository{
        FindByIDFunc: func(id int) (*Product, error) {
            return &Product{ID: 1, Name: "测试商品"}, nil
        },
    }

    mockAddresses := &mocks.MockAddressRepository{
        FindByIDAndCustomerFunc: func(id int, customerID int) (*Address, error) {
            return &Address{ID: 1, MerchantID: 1}, nil
        },
    }

    mockCache := &mocks.MockCacheService{}
    mockEvents := &mocks.MockEventPublisher{}

    svc := NewOrderService(mockOrders, mockProducts, mockSKUs, mockAddresses, mockCache, mockEvents)

    // Act — 执行测试
    items := []OrderItemInput{
        {ProductID: 1, SkuID: 1, Quantity: 2},
    }
    order, err := svc.CreateOrder(1, 1, items, "测试备注")

    // Assert — 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, order)
    assert.Equal(t, decimal.NewFromFloat(199.80), order.TotalAmount)
    assert.Equal(t, StatusPending, order.Status)
}

func TestOrderService_CreateOrder_StockInsufficient(t *testing.T) {
    // Arrange
    mockSKUs := &mocks.MockSKURepository{
        FindByIDFunc: func(id int) (*SKU, error) {
            return &SKU{ID: 1, Stock: 1, Price: decimal.NewFromFloat(99.90)}, nil
        },
    }

    mockProducts := &mocks.MockProductRepository{
        FindByIDFunc: func(id int) (*Product, error) {
            return &Product{ID: 1, Name: "测试商品"}, nil
        },
    }

    mockAddresses := &mocks.MockAddressRepository{
        FindByIDAndCustomerFunc: func(id int, customerID int) (*Address, error) {
            return &Address{ID: 1, MerchantID: 1}, nil
        },
    }

    mockOrders := &mocks.MockOrderRepository{}
    mockCache := &mocks.MockCacheService{}
    mockEvents := &mocks.MockEventPublisher{}

    svc := NewOrderService(mockOrders, mockProducts, mockSKUs, mockAddresses, mockCache, mockEvents)

    // Act — 请求 2 件，库存只有 1 件
    items := []OrderItemInput{
        {ProductID: 1, SkuID: 1, Quantity: 2},
    }
    order, err := svc.CreateOrder(1, 1, items, "")

    // Assert
    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Equal(t, ErrStockInsufficient, err)
}

func TestOrderService_CreateOrder_AddressNotFound(t *testing.T) {
    // Arrange
    mockAddresses := &mocks.MockAddressRepository{
        FindByIDAndCustomerFunc: func(id int, customerID int) (*Address, error) {
            return nil, ErrAddressNotFound
        },
    }

    mockOrders := &mocks.MockOrderRepository{}
    mockProducts := &mocks.MockProductRepository{}
    mockSKUs := &mocks.MockSKURepository{}
    mockCache := &mocks.MockCacheService{}
    mockEvents := &mocks.MockEventPublisher{}

    svc := NewOrderService(mockOrders, mockProducts, mockSKUs, mockAddresses, mockCache, mockEvents)

    // Act
    order, err := svc.CreateOrder(1, 999, nil, "")

    // Assert
    assert.Error(t, err)
    assert.Equal(t, ErrAddressNotFound, err)
}
```

### 4.3 测试对比

| 维度 | 当前写法 | DDD + 接口注入 |
|:---|:---|:---|
| 需要 MySQL | 是 | 否 |
| 需要 Redis | 是 | 否 |
| 测试速度 | 慢（秒级） | 快（毫秒级） |
| 能测边界条件 | 难（要构造 DB 数据） | 易（直接设 Mock 返回值） |
| 能测错误路径 | 难（DB 不好模拟错误） | 易（Mock 返回 error） |
| 测试稳定性 | 低（依赖外部服务） | 高（纯内存） |

---

## 五、推荐的 Go 项目结构

### 5.1 推荐目录结构

```
shop-backend/
├── cmd/
│   └── server/
│       └── main.go              # 入口，只做组装
├── domain/                       # 领域层（纯 Go，零外部依赖）
│   ├── order/
│   │   ├── order.go             # 实体 + 值对象
│   │   ├── repository.go        # 仓库接口（领域层定义）
│   │   ├── service.go           # 领域服务
│   │   ├── service_test.go      # 领域服务测试
│   │   └── errors.go            # 领域错误
│   ├── product/
│   │   ├── product.go
│   │   ├── repository.go
│   │   └── service.go
│   └── customer/
│       ├── customer.go
│       └── repository.go
├── infrastructure/               # 基础设施层（实现领域接口）
│   ├── persistence/
│   │   ├── gorm_order_repo.go   # OrderRepository 的 GORM 实现
│   │   ├── gorm_product_repo.go
│   │   └── po/                  # 持久化对象（PO）
│   │       └── order_po.go      # DB 模型 ↔ 领域模型转换
│   ├── cache/
│   │   └── redis_cache.go       # CacheService 的 Redis 实现
│   └── mq/
│       └── rabbitmq_publisher.go # EventPublisher 的 RabbitMQ 实现
├── interfaces/                   # 接口层（HTTP/gRPC）
│   ├── http/
│   │   ├── order_controller.go  # HTTP Handler
│   │   ├── middleware/
│   │   └── router.go
│   └── dto/
│       └── order_dto.go         # 请求/响应 DTO
├── config/
│   └── config.go
├── go.mod
└── go.sum
```

### 5.2 依赖方向规则

```
interfaces（HTTP层）──→ domain（领域层）←── infrastructure（基础设施层）

规则：
1. domain 不依赖任何外部包（零 import 外部库）
2. infrastructure 依赖 domain（实现 domain 的接口）
3. interfaces 依赖 domain（调用 domain 的服务）
4. interfaces 不依赖 infrastructure（通过接口解耦）
5. main.go 负责把 infrastructure 实现注入到 domain 和 interfaces
```

### 5.3 渐进式改造建议

不要一次性重构，按以下顺序逐步改造：

| 阶段 | 改造内容 | 风险 |
|:---|:---|:---|
| Phase 1 | 为 Service 层提取接口，Controller 依赖接口而非具体实现 | 低 |
| Phase 2 | 将 Service 中的 DB 操作提取到 Repository 接口 | 中 |
| Phase 3 | 将业务规则从 Service 移入实体（充血模型） | 中 |
| Phase 4 | 按领域划分子包（domain/order, domain/product） | 高 |
| Phase 5 | 分离 PO 和领域模型，基础设施层做转换 | 高 |

---

## 六、核心结论

| 问题 | 回答 |
|:---|:---|
| Java 味重吗？ | 重。三层架构 + 贫血模型 + 全局变量 + 无接口抽象 |
| DDD 是什么？ | 让代码结构反映业务领域，不是框架，Go 实现比 Java 简单 |
| 依赖注入怎么写？ | 接口 + 构造器注入，不需要框架，手动组装即可 |
| 单元测试怎么写？ | 先有接口，再写 Mock，最后写测试。没有接口就没有单元测试 |
| 要不要全部重构？ | 不要。渐进式改造，先加接口，再抽 Repository，最后充血化 |

**一句话总结**：Go 的哲学是"少即是多"——少一层抽象、少一个框架、少一个全局变量，用接口表达意图，用组合代替继承，用简单代替复杂。你的项目功能没问题，但结构上可以更 Go。
