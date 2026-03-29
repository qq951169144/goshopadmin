# GoShopAdmin 项目规则

本文件定义了AI在执行计划和编写代码时必须强制遵守的规范。

---

## 规则1：统一响应处理

**强制要求**：所有API响应必须通过 BaseController 的统一方法处理。

### 成功响应
- **位置**: `d:\code\goshopadmin\shop-backend\controllers\base_controller.go#L14-20`
- **方法**: `ResponseSuccess(ctx *gin.Context, data interface{})`
- **必须**: 使用此方法返回成功响应，HTTP状态码始终为200

### 错误响应
- **位置**: `d:\code\goshopadmin\shop-backend\controllers\base_controller.go#L27-60`
- **方法**: `ResponseError(ctx *gin.Context, bizCode int, err error)`
- **必须**: 使用此方法返回错误响应，HTTP状态码始终为200，错误信息放在body中
- **注意**: 前端根据 body.code 判断成功/失败

### 例外情况
以下情况可以选择例外，不强制使用统一响应：
- 返回验证码图片（如 captcha 接口）
- 返回文件流（如图片下载）
- 健康检查接口（如 /health）
- 其他特殊二进制响应

---

## 规则2：数据类型规范

### 2.1 结构体整数类型
**强制要求**: 结构体中的整数类型一律使用 `int`，禁止使用 `uint`。

```go
// ✅ 正确
func (s *Service) GetCount() int {
    return count
}

type User struct {
    ID   int
    Age  int
    Count int
}

// ❌ 错误
func (s *Service) GetCount() uint {
    return count
}

type User struct {
    ID   uint
    Age  uint
}
```

### 2.2 数据库表设计
**强制要求**: 
1. 涉及整数类型时，**不能**设置"无符号"（UNSIGNED）
2. 排序规则一律使用：`utf8mb4_unicode_ci`
3. **绝对不能使用外键约束（FOREIGN KEY）**

```sql
-- ✅ 正确
CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT,
    age INT DEFAULT 0,
    count INT DEFAULT 0,
    name VARCHAR(255),
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 关联表使用普通索引，不使用外键约束
CREATE TABLE orders (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_orders_user_id (user_id)  -- 普通索引，无外键约束
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ❌ 错误 - 使用了UNSIGNED
CREATE TABLE users (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    age INT UNSIGNED DEFAULT 0,
    PRIMARY KEY (id)
);

-- ❌ 错误 - 排序规则不正确
CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ❌ 错误 - 使用了外键约束
CREATE TABLE orders (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_orders_user_id (user_id),
    CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id)  -- 禁止使用
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**不使用外键约束的原因**:
1. **性能考虑**: 外键约束会增加写入操作的开销，影响数据库性能
2. **灵活性**: 无外键约束便于数据迁移、分库分表和分布式部署
3. **维护便利**: 避免级联删除/更新带来的意外数据丢失
4. **应用层控制**: 通过代码逻辑维护数据一致性，更加可控
5. **高并发场景**: 外键约束可能导致锁竞争，影响并发性能

### 2.3 GORM模型定义
**强制要求**: 使用 `int` 类型，GORM标签中不使用 `unsigned`

```go
// ✅ 正确
type User struct {
    ID    int    `gorm:"primaryKey;type:int;autoIncrement"`
    Age   int    `gorm:"type:int;default:0"`
    Count int    `gorm:"type:int;default:0"`
}

// ❌ 错误 - 使用了uint
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Age   uint   `gorm:"default:0"`
}
```

---

## 规则3：路由风格统一

**强制要求**: 路由写法必须保持与现有项目风格一致。

### shop-backend 路由风格
**参考文件**: `d:\code\goshopadmin\shop-backend\routes\routes.go`

```go
// 风格特点：
// 1. 使用 Dependencies 结构体注入控制器
// 2. 路由分组清晰，使用中间件
// 3. 路径简洁，使用复数名词

api := router.Group("/api")
{
    // 验证码路由
    api.GET("/captcha", deps.CaptchaController.GenerateCaptcha)
    api.POST("/captcha/verify", deps.CaptchaController.VerifyCaptcha)

    // 认证路由
    auth := api.Group("/auth")
    {
        auth.POST("/register", deps.AuthController.Register)
        auth.POST("/login", deps.AuthController.Login)
        auth.POST("/logout", middleware.Auth(), deps.AuthController.Logout)
    }

    // 需要认证的路由组
    user := api.Group("/user", middleware.Auth())
    {
        user.GET("/profile", deps.CustomerController.GetProfile)
    }
}
```

### backend 路由风格
**参考文件**: `d:\code\goshopadmin\backend\routes\routes.go`

```go
// 风格特点：
// 1. 使用 Dependencies 结构体注入控制器
// 2. 嵌套路由组，中间件分层
// 3. RESTful 风格路径

api := r.Group("/api")
{
    auth := api.Group("/auth")
    {
        auth.POST("/login", deps.AuthController.Login)
        
        // 需要认证的子路由
        authProtected := auth.Group("/")
        authProtected.Use(middleware.AuthMiddleware())
        {
            authProtected.POST("/logout", deps.AuthController.Logout)
        }
    }

    // 受保护的业务路由
    protected := api.Group("/")
    protected.Use(middleware.AuthMiddleware())
    {
        users := protected.Group("/users")
        {
            users.GET("", deps.UserController.GetUsers)
            users.GET("/:id", deps.UserController.GetUser)
            users.POST("", deps.UserController.CreateUser)
            users.PUT("/:id", deps.UserController.UpdateUser)
            users.DELETE("/:id", deps.UserController.DeleteUser)
        }
    }
}
```

### 路由命名规范
1. **使用复数名词**: `/users`, `/products`, `/orders`
2. **资源操作**: GET获取, POST创建, PUT更新, DELETE删除
3. **嵌套资源**: `/products/:id/skus`, `/orders/:id/items`
4. **动作路由**: `/orders/:id/cancel`, `/products/:id/publish`

---

## 规则4：Docker容器重启

**强制要求**: 修改代码后必须重启对应的Docker容器，以保证新代码生效。

### Docker Compose文件位置
- **文件**: `d:\code\goshopadmin\docker\docker-compose.yml`

### 项目目录与容器对应关系

| 修改的目录 | 对应容器名称 | 重启命令 |
| :--- | :--- | :--- |
| `d:\code\goshopadmin\backend` | `goshopadmin-backend` | `docker restart goshopadmin-backend` |
| `d:\code\goshopadmin\shop-backend` | `goshopadmin-shop-backend` | `docker restart goshopadmin-shop-backend` |
| `d:\code\goshopadmin\frontend` | `goshopadmin-frontend` | `docker restart goshopadmin-frontend` |
| `d:\code\goshopadmin\shop-frontend` | `goshopadmin-shop-frontend` | `docker restart goshopadmin-shop-frontend` |

### 重启步骤

```bash
# 1. 进入docker目录
cd d:\code\goshopadmin\docker

# 2. 根据修改的项目重启对应容器
# 例如：修改了 backend 代码
docker restart goshopadmin-backend

# 例如：修改了 shop-backend 代码
docker restart goshopadmin-shop-backend

# 例如：修改了 frontend 代码
docker restart goshopadmin-frontend

# 例如：修改了 shop-frontend 代码
docker restart goshopadmin-shop-frontend
```

### 注意事项
1. 容器使用 volume 挂载本地代码（如 `../backend:/app`），所以只需重启容器即可加载新代码
2. 重启后等待容器健康检查通过（约10-30秒）
3. 可通过 `docker logs <容器名>` 查看启动日志确认是否成功

---

## 规则5：API接口文档更新

**强制要求**: 增加或修改API代码后，必须更新接口文档。

### 需要更新的情况

当修改以下目录中的控制器代码时：
- `d:\code\goshopadmin\shop-backend\controllers`
- `d:\code\goshopadmin\backend\controllers`

### 文档更新流程

1. **重启Docker容器**（按规则4执行）
2. **测试API接口** - 确保新接口正常工作
3. **更新文档** - 修改 `d:\code\goshopadmin\API接口和前端对接情况.md`

### 文档格式规范

新增API接口时，按照以下格式添加到文档中：

```markdown
| 接口路径 | 方法 | 功能描述 | 请求参数 | 成功响应 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/xxx` | `GET/POST/PUT/DELETE` | 功能说明 | `{"key": "value"}` | `{"code": 200, "message": "...", "data": {...}}` |
```

### 更新位置

- **后台管理API**: 在文档的 "一、后端API接口实现" 章节添加
- **C端商城API**: 在文档的 "6. C端商城API" 章节添加

---

## 规则6：数据库文档同步

**强制要求**: 修改数据库结构后，必须同步更新数据库文档。

### 需要更新的情况

1. 修改了 `d:\code\goshopadmin\docker\mysql\init.sql`
2. 进入Docker容器手动修改了数据库表结构

### 文档更新流程

1. **执行数据库变更** - 修改SQL文件或在容器内执行DDL
2. **验证变更** - 确认表结构变更正确
3. **更新文档** - 修改 `d:\code\goshopadmin\数据库表结构说明.md`

### 文档格式规范

#### 新增表时

按照以下格式添加新表说明：

```markdown
### 2.X 表名 表

**描述**: 表功能描述

| 字段名 | 数据类型 | 约束 | 默认值 | 描述 |
| :--- | :--- | :--- | :--- | :--- |
| `id` | `int(0)` | `NOT NULL AUTO_INCREMENT` | - | 主键ID |
| `xxx` | `varchar(255)` | `NOT NULL` | - | 字段说明 |

**索引**: 
- 主键: `id`
- 普通索引: `idx_xxx` (`字段名`)

**引擎**: InnoDB
```

#### 修改表时

- 更新字段列表
- 更新索引信息
- 更新表关系图（如影响关联关系）

### 注意事项

1. **数据类型**: 确保与规则2一致，使用 `int` 而非 `uint`，不使用 `UNSIGNED`
2. **排序规则**: 确保使用 `utf8mb4_unicode_ci`
3. **表关系**: 如果新增/修改了外键关系，需要更新 "3. 表关系图" 章节
4. **初始数据**: 如果有初始数据，更新 "4. 初始数据" 章节

---

## 规则7：枚举类型规范

### 7.1 数据库状态字段设计

**强制要求**: 数据库设计涉及状态（status）字段时，**必须**采用枚举类型（ENUM），禁止使用字符串或整数随意存储。

#### 正确示例

```sql
-- ✅ 正确 - 使用ENUM类型定义状态
CREATE TABLE products (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    status ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    audit_status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE orders (
    id INT NOT NULL AUTO_INCREMENT,
    order_no VARCHAR(64) NOT NULL,
    status ENUM('pending', 'paid', 'shipped', 'completed', 'cancelled') NOT NULL DEFAULT 'pending',
    payment_status ENUM('pending', 'success', 'failed') NOT NULL DEFAULT 'pending',
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 错误示例

```sql
-- ❌ 错误 - 使用VARCHAR存储状态
CREATE TABLE products (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 错误！应该用ENUM
    PRIMARY KEY (id)
);

-- ❌ 错误 - 使用INT存储状态（无业务含义）
CREATE TABLE orders (
    id INT NOT NULL AUTO_INCREMENT,
    status INT NOT NULL DEFAULT 0,  -- 错误！应该用ENUM，0/1/2无明确含义
    PRIMARY KEY (id)
);
```

#### 使用ENUM的优势

1. **数据完整性**: 数据库层面限制只能存储预定义的值
2. **可读性**: 状态值具有明确业务含义（如 'pending', 'completed'）
3. **性能**: ENUM在MySQL内部以整数存储，查询效率高
4. **维护性**: 状态值集中管理，便于理解和维护

### 7.2 常量定义同步

**强制要求**: 设计完数据库ENUM类型后，**必须**同步更新Go代码中的常量定义。

#### 同步位置

| 项目 | 常量文件路径 |
| :--- | :--- |
| 后台管理系统 | `d:\code\goshopadmin\backend\constants\constants.go` |
| C端商城 | `d:\code\goshopadmin\shop-backend\constants\constants.go` |

#### 常量命名规范

```go
// ✅ 正确 - 命名规范：{功能}Status = "枚举值"
const (
    // ProductStatusActive 商品激活状态
    ProductStatusActive = "active"
    
    // ProductStatusInactive 商品禁用状态
    ProductStatusInactive = "inactive"
)

// ✅ 正确 - 订单状态
const (
    // OrderStatusPending 待支付状态
    OrderStatusPending = "pending"
    
    // OrderStatusPaid 已支付状态
    OrderStatusPaid = "paid"
    
    // OrderStatusShipped 已发货状态
    OrderStatusShipped = "shipped"
    
    // OrderStatusCompleted 已完成状态
    OrderStatusCompleted = "completed"
    
    // OrderStatusCancelled 已取消状态
    OrderStatusCancelled = "cancelled"
)
```

### 7.3 避免硬编码

**强制要求**: 代码中**禁止**直接使用字符串或数字硬编码状态值，必须使用常量。

#### 正确示例

```go
// ✅ 正确 - 使用常量
import "goshopadmin/backend/constants"

// 查询激活状态的商品
products, err := s.productRepo.GetByStatus(constants.ProductStatusActive)

// 更新订单状态为已完成
order.Status = constants.OrderStatusCompleted

// 条件判断使用常量
if product.Status == constants.ProductStatusActive {
    // 处理激活状态逻辑
}

// SQL查询中使用常量
rows, err := db.Query("SELECT * FROM products WHERE status = ?", constants.ProductStatusActive)
```

#### 错误示例

```go
// ❌ 错误 - 硬编码状态值
products, err := s.productRepo.GetByStatus("active")  // 错误！应该用常量

// ❌ 错误 - 硬编码数字状态
order.Status = 1  // 错误！1代表什么？可读性差

// ❌ 错误 - 条件判断硬编码
if product.Status == "inactive" {  // 错误！应该用常量
    // 处理逻辑
}
```

### 7.4 现有常量参考

当前项目中已定义的常量（位于 `constants/constants.go`）：

```go
// Status 通用状态
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
)

// AuditStatus 审核状态
const (
    AuditStatusPending  = "pending"
    AuditStatusApproved = "approved"
    AuditStatusRejected = "rejected"
)

// OrderStatus 订单状态
const (
    OrderStatusPending   = "pending"
    OrderStatusPaid      = "paid"
    OrderStatusShipped   = "shipped"
    OrderStatusCompleted = "completed"
    OrderStatusCancelled = "cancelled"
)

// PaymentStatus 支付状态
const (
    PaymentStatusPending = "pending"
    PaymentStatusSuccess = "success"
    PaymentStatusFailed  = "failed"
)

// ShippingStatus 物流状态
const (
    ShippingStatusPending   = "pending"
    ShippingStatusShipped   = "shipped"
    ShippingStatusDelivered = "delivered"
    ShippingStatusReturned  = "returned"
)
```

### 7.5 新增状态流程

当需要新增状态类型时，按以下流程操作：

1. **设计数据库表**: 使用ENUM类型定义状态字段
2. **更新常量文件**: 在 `constants/constants.go` 中添加对应的常量定义
3. **更新模型定义**: 在GORM模型中使用常量作为默认值
4. **代码中使用**: 所有地方使用常量，禁止硬编码

---

## 规则8：用户命名规范

**强制要求**: 不同项目的用户相关命名必须严格区分，保持与现有项目一致。

### 8.1 后台管理系统（backend/frontend）

**项目路径**:
- 后端: `d:\code\goshopadmin\backend`
- 前端: `d:\code\goshopadmin\frontend`

**命名规范**:
- **变量/字段名**: 使用 `user`
- **数据库表**: `users` 表
- **含义**: 表示商城后台管理系统的用户（管理员、运营人员等）

```go
// ✅ 正确 - backend 项目
// 模型定义
models/user.go

type User struct {
    ID       int    `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"size:50;unique;not null"`
    Password string `json:"password" gorm:"size:100;not null"`
}

func (User) TableName() string {
    return "users"
}

// 服务层
services/user_service.go

func (s *UserService) GetUserByID(userID int) (*models.User, error)
func (s *UserService) CreateUser(user *models.User) error

// 控制器返回
// { "user_id": 1, "username": "admin" }
```

```javascript
// ✅ 正确 - frontend 项目
// 存储用户信息
localStorage.setItem('user_id', response.user_id)
localStorage.setItem('user', JSON.stringify(response.user))

// API 调用
api.get('/users/profile')
```

### 8.2 C端商城（shop-backend/shop-frontend）

**项目路径**:
- 后端: `d:\code\goshopadmin\shop-backend`
- 前端: `d:\code\goshopadmin\shop-frontend`

**命名规范**:
- **变量/字段名**: 使用 `customer`
- **数据库表**: `customers` 表
- **含义**: 表示C端商城的消费者/客户

```go
// ✅ 正确 - shop-backend 项目
// 模型定义
models/customer.go

type Customer struct {
    ID       int    `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"size:50;unique;not null"`
    Password string `json:"password" gorm:"size:100;not null"`
}

func (Customer) TableName() string {
    return "customers"
}

// 服务层
services/customer_service.go

func (s *CustomerService) GetCustomerByID(customerID int) (*models.Customer, error)
func (s *CustomerService) CreateCustomer(customer *models.Customer) error

// 控制器返回
// { "customer_id": 1, "username": "customer001" }
```

```javascript
// ✅ 正确 - shop-frontend 项目
// 存储用户信息
localStorage.setItem('customer_id', response.customer_id)
localStorage.setItem('customer', JSON.stringify(response.customer))

// API 调用
api.get('/customer/profile')
```

### 8.3 命名对照表

| 项目 | 后端目录 | 前端目录 | 命名 | 数据库表 |
| :--- | :--- | :--- | :--- | :--- |
| 后台管理系统 | `backend` | `frontend` | `user` | `users` |
| C端商城 | `shop-backend` | `shop-frontend` | `customer` | `customers` |

### 8.4 违规示例

```go
// ❌ 错误 - backend 项目使用了 customer
// backend 项目应该使用 user
type Customer struct {  // 错误！backend 项目应该用 User
    ID int
}

// ❌ 错误 - shop-backend 项目使用了 user
// shop-backend 项目应该使用 customer
type User struct {  // 错误！shop-backend 项目应该用 Customer
    ID int
}

// ❌ 错误 - 返回字段名不一致
// shop-backend 应该返回 customer_id
c.JSON(200, gin.H{"user_id": 1})  // 错误！应该用 customer_id

// ❌ 错误 - frontend 存储字段名错误
// frontend 应该存储 user_id
localStorage.setItem('customer_id', response.user_id)  // 错误！backend 项目用 user_id
```

---

## 规则9：错误码规范

### 9.1 错误码定义文件

**强制要求**: 所有错误码必须定义在统一的错误码文件中，禁止在代码中硬编码错误码。

#### 错误码文件位置

| 项目 | 错误码文件路径 |
| :--- | :--- |
| 后台管理系统 | `d:\code\goshopadmin\backend\errors\code.go` |
| C端商城 | `d:\code\goshopadmin\shop-backend\errors\code.go` |

### 9.2 错误码格式规范

错误码格式：**HTTP状态码(1位) + 模块(1位) + 序号(2位)**

```go
// 错误码定义示例
const (
    // 成功
    CodeSuccess = 0

    // 4000 - 参数错误 (4xx 客户端错误)
    CodeParamError      = 4001 // 通用参数错误
    CodeParamMissing    = 4002 // 参数缺失
    CodeParamInvalid    = 4003 // 参数格式无效
    CodeParamOutOfRange = 4004 // 参数超出范围

    // 4010 - 认证错误
    CodeUnauthorized = 4010 // 未认证
    CodeTokenExpired = 4011 // Token 过期
    CodeTokenInvalid = 4012 // Token 无效
    CodeLoginFailed  = 4013 // 登录失败

    // 4030 - 权限错误
    CodeForbidden      = 4030 // 权限不足
    CodeResourceDenied = 4031 // 资源访问被拒绝

    // 4040 - 资源错误
    CodeNotFound        = 4040 // 资源不存在
    CodeUserNotFound    = 4041 // 用户不存在
    CodeProductNotFound = 4042 // 商品不存在

    // 4090 - 业务冲突
    CodeConflict  = 4090 // 资源冲突
    CodeDuplicate = 4091 // 数据重复

    // 5000 - 服务器错误 (5xx 服务端错误)
    CodeInternalError = 5000 // 内部错误
    CodeDBError       = 5001 // 数据库错误
    CodeCacheError    = 5002 // 缓存错误
    CodeExternalError = 5003 // 外部服务错误
)
```

### 9.3 错误码使用规范

**强制要求**: 代码中必须使用错误码常量，禁止直接使用数字。

#### 正确示例

```go
// ✅ 正确 - 使用错误码常量
import "goshopadmin/backend/errors"

// 返回参数错误
return errors.New(errors.CodeParamError, "用户名不能为空")

// 返回资源不存在
return errors.New(errors.CodeUserNotFound, "用户不存在")

// 返回数据库错误
return errors.New(errors.CodeDBError, "查询失败")

// 控制器中使用
s.ResponseError(ctx, errors.CodeParamError, err)
```

#### 错误示例

```go
// ❌ 错误 - 硬编码错误码
return errors.New(4001, "用户名不能为空")  // 应该用常量 errors.CodeParamError

// ❌ 错误 - 直接返回数字
c.JSON(200, gin.H{"code": 5000, "message": "系统错误"})  // 应该用常量
```

### 9.4 新增错误码流程

当需要新增错误码时，按以下流程操作：

1. **确定错误类型**: 根据错误性质确定所属模块（4xx客户端错误/5xx服务端错误）
2. **分配错误码**: 按照格式分配新的错误码，确保不重复
3. **添加错误消息**: 在 `ErrorMessage` map 中添加对应的友好提示
4. **代码中使用**: 使用新增的错误码常量

```go
// 新增错误码示例
const (
    // ... 现有错误码 ...
    
    // 新增：4050 - 订单相关错误
    CodeOrderNotFound   = 4050 // 订单不存在
    CodeOrderCancelled  = 4051 // 订单已取消
    CodeOrderCompleted  = 4052 // 订单已完成
)

// 在 ErrorMessage 中添加对应消息
var ErrorMessage = map[int]string{
    // ... 现有消息 ...
    CodeOrderNotFound:  "订单不存在",
    CodeOrderCancelled: "订单已取消，无法操作",
    CodeOrderCompleted: "订单已完成",
}
```

### 9.5 现有错误码参考

#### 后台管理系统错误码（backend/errors/code.go）

| 错误码 | 常量名 | 说明 |
| :--- | :--- | :--- |
| 0 | `CodeSuccess` | 成功 |
| 4001 | `CodeParamError` | 通用参数错误 |
| 4002 | `CodeParamMissing` | 参数缺失 |
| 4003 | `CodeParamInvalid` | 参数格式无效 |
| 4004 | `CodeParamOutOfRange` | 参数超出范围 |
| 4010 | `CodeUnauthorized` | 未认证 |
| 4011 | `CodeTokenExpired` | Token 过期 |
| 4012 | `CodeTokenInvalid` | Token 无效 |
| 4013 | `CodeLoginFailed` | 登录失败 |
| 4030 | `CodeForbidden` | 权限不足 |
| 4031 | `CodeResourceDenied` | 资源访问被拒绝 |
| 4040 | `CodeNotFound` | 资源不存在 |
| 4041 | `CodeUserNotFound` | 用户不存在 |
| 4042 | `CodeProductNotFound` | 商品不存在 |
| 4090 | `CodeConflict` | 资源冲突 |
| 4091 | `CodeDuplicate` | 数据重复 |
| 5000 | `CodeInternalError` | 内部错误 |
| 5001 | `CodeDBError` | 数据库错误 |
| 5002 | `CodeCacheError` | 缓存错误 |
| 5003 | `CodeExternalError` | 外部服务错误 |

#### C端商城错误码（shop-backend/errors/code.go）

| 错误码 | 常量名 | 说明 |
| :--- | :--- | :--- |
| 0 | `CodeSuccess` | 成功 |
| 4001 | `CodeParamError` | 通用参数错误 |
| 4002 | `CodeParamMissing` | 参数缺失 |
| 4003 | `CodeParamInvalid` | 参数格式无效 |
| 4004 | `CodeParamOutOfRange` | 参数超出范围 |
| 4010 | `CodeUnauthorized` | 未认证 |
| 4011 | `CodeTokenExpired` | Token 过期 |
| 4012 | `CodeTokenInvalid` | Token 无效 |
| 4013 | `CodeLoginFailed` | 登录失败 |
| 4030 | `CodeForbidden` | 权限不足 |
| 4031 | `CodeResourceDenied` | 资源访问被拒绝 |
| 4040 | `CodeNotFound` | 资源不存在 |
| 4041 | `CodeUserNotFound` | 用户不存在 |
| 4042 | `CodeProductNotFound` | 商品不存在 |
| 4043 | `CodeCartNotFound` | 购物车不存在 |
| 4044 | `CodeOrderNotFound` | 订单不存在 |
| 4090 | `CodeConflict` | 资源冲突 |
| 4091 | `CodeDuplicate` | 数据重复 |
| 4092 | `CodeStockInsufficient` | 库存不足 |
| 4093 | `CodeCaptchaError` | 验证码错误 |
| 4094 | `CodeUserExists` | 用户名已存在 |
| 5000 | `CodeInternalError` | 内部错误 |
| 5001 | `CodeDBError` | 数据库错误 |
| 5002 | `CodeCacheError` | 缓存错误 |
| 5003 | `CodeExternalError` | 外部服务错误 |

---

## 规则10：日志工具使用规范

**强制要求**: 当需要写入日志时，必须统一使用项目提供的日志工具类。

### 日志工具位置

| 项目 | 日志工具文件路径 |
| :--- | :--- |
| 后台管理系统 | `d:\code\goshopadmin\backend\utils\logger.go` |
| C端商城 | `d:\code\goshopadmin\shop-backend\utils\logger.go` |

### 使用方法

**强制要求**: 所有日志记录必须使用以下方法：

```go
// 导入日志工具包
import "goshopadmin/backend/utils"  // 后台管理系统
// 或
import "goshopadmin/shop-backend/utils"  // C端商城

// 记录信息日志
utils.Info("用户登录成功: %s", username)

// 记录警告日志
utils.Warn("库存不足: %s", productName)

// 记录错误日志
utils.Error("数据库查询失败: %v", err)
```

### 禁止使用

**禁止**直接使用 `fmt.Println`、`log.Println` 等其他方式记录日志。

```go
// ❌ 错误 - 直接使用 fmt 打印
fmt.Println("用户登录成功")

// ❌ 错误 - 直接使用 log 包
log.Println("库存不足")

// ✅ 正确 - 使用统一日志工具
utils.Info("用户登录成功")
```

---

## 规则11：禁止使用TODO占位

**强制要求**: AI在生成代码时，禁止使用 `TODO` 注释代替功能实现，除非用户明确命令某个地方用 `TODO` 占位。

### 正确示例

```go
// ✅ 正确 - 完整实现功能
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// ✅ 正确 - 实现简单逻辑
func (s *ProductService) CalculatePrice(price float64, discount float64) float64 {
    return price * (1 - discount)
}
```

### 错误示例

```go
// ❌ 错误 - 使用TODO代替功能实现
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    // TODO: 实现查询逻辑
    return nil, nil
}

// ❌ 错误 - 使用TODO代替完整实现
func (s *OrderService) CreateOrder(order *models.Order) error {
    // TODO: 实现订单创建逻辑
    // TODO: 处理库存
    // TODO: 生成订单号
    return nil
}
```

### 例外情况

只有在用户明确要求使用 `TODO` 占位时，才可以使用 `TODO` 注释：

```go
// ✅ 正确 - 用户明确要求使用TODO占位
func (s *ProductService) SyncInventory() error {
    // TODO: 后续实现库存同步逻辑（用户要求暂时占位）
    return nil
}
```

---

## 规则12：代码风格规范

### 12.1 避免if语句过长

**强制要求**: 编写代码时需要避免if语句过长造成阅读障碍，应将复杂条件提取为有意义的变量或方法。

#### 正确示例

```go
// ✅ 正确 - 提取条件为有意义的变量
func (s *OrderService) ProcessOrder(order *models.Order) error {
    isPaid := order.PaymentStatus == constants.PaymentStatusSuccess
    hasStock := order.Stock > 0
    canShip := isPaid && hasStock
    
    if canShip {
        return s.shipOrder(order)
    }
    return nil
}

// ✅ 正确 - 提取为独立方法
func (s *OrderService) ProcessOrder(order *models.Order) error {
    if s.canProcessOrder(order) {
        return s.shipOrder(order)
    }
    return nil
}

func (s *OrderService) canProcessOrder(order *models.Order) bool {
    return order.PaymentStatus == constants.PaymentStatusSuccess && 
           order.Stock > 0 &&
           order.Status == constants.OrderStatusPending
}

// ✅ 正确 - 使用早返回减少嵌套
func (s *OrderService) ProcessOrder(order *models.Order) error {
    if order.Status != constants.OrderStatusPending {
        return errors.New("订单状态不正确")
    }
    if order.PaymentStatus != constants.PaymentStatusSuccess {
        return errors.New("订单未支付")
    }
    if order.Stock <= 0 {
        return errors.New("库存不足")
    }
    
    return s.shipOrder(order)
}
```

#### 错误示例

```go
// ❌ 错误 - if条件过长，难以阅读
func (s *OrderService) ProcessOrder(order *models.Order) error {
    if order.PaymentStatus == constants.PaymentStatusSuccess && order.Stock > 0 && order.Status == constants.OrderStatusPending && order.User != nil {
        return s.shipOrder(order)
    }
    return nil
}

// ❌ 错误 - 嵌套过深
func (s *OrderService) ProcessOrder(order *models.Order) error {
    if order.Status == constants.OrderStatusPending {
        if order.PaymentStatus == constants.PaymentStatusSuccess {
            if order.Stock > 0 {
                if order.User != nil {
                    return s.shipOrder(order)
                }
            }
        }
    }
    return nil
}
```

### 12.2 请求结构体定义规范

**强制要求**: req入参如果是struct的话，需要用 `type struct` 写法，并且放在调用该结构体的方法前面，写好注释。

#### 正确示例

```go
// ✅ 正确 - 请求结构体定义在方法前面，带有注释

// LoginRequest 登录请求参数
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
    var req LoginRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ResponseError(ctx, errors.CodeParamError, err)
        return
    }
    // ...
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
}

// Register 用户注册
func (c *AuthController) Register(ctx *gin.Context) {
    var req RegisterRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ResponseError(ctx, errors.CodeParamError, err)
        return
    }
    // ...
}
```

#### 错误示例

```go
// ❌ 错误 - 结构体定义在方法后面
func (c *AuthController) Login(ctx *gin.Context) {
    var req LoginRequest
    // ...
}

type LoginRequest struct {  // 错误！应该定义在方法前面
    Username string `json:"username"`
    Password string `json:"password"`
}

// ❌ 错误 - 结构体没有注释
type LoginRequest struct {  // 错误！缺少注释说明
    Username string `json:"username"`
    Password string `json:"password"`
}

// ❌ 错误 - 使用匿名结构体
func (c *AuthController) Login(ctx *gin.Context) {
    var req struct {  // 错误！应该定义命名结构体
        Username string `json:"username"`
        Password string `json:"password"`
    }
    // ...
}
```

### 12.3 方法注释规范

**强制要求**: 每个方法最好写一下简单的功能注释，说明方法的作用。

#### 正确示例

```go
// ✅ 正确 - 方法带有功能注释

// GetUserByID 根据用户ID获取用户信息
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    return s.userRepo.FindByID(userID)
}

// CreateOrder 创建订单，包含库存检查和订单号生成
func (s *OrderService) CreateOrder(order *models.Order) error {
    // 检查库存
    if err := s.checkStock(order.ProductID, order.Quantity); err != nil {
        return err
    }
    // 生成订单号
    order.OrderNo = s.generateOrderNo()
    return s.orderRepo.Create(order)
}

// CalculatePrice 计算商品折扣价格
func (s *ProductService) CalculatePrice(price float64, discount float64) float64 {
    return price * (1 - discount)
}
```

#### 错误示例

```go
// ❌ 错误 - 方法没有注释
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    return s.userRepo.FindByID(userID)
}

// ❌ 错误 - 注释过于简单，没有说明功能
func (s *OrderService) CreateOrder(order *models.Order) error {  // 创建
    return s.orderRepo.Create(order)
}

// ❌ 错误 - 注释与实际功能不符
// DeleteOrder 删除订单
func (s *OrderService) CreateOrder(order *models.Order) error {  // 错误！方法名是CreateOrder
    return s.orderRepo.Create(order)
}
```

### 12.4 控制器错误处理规范

**强制要求**: `c.ResponseError` 需要传入原始错误 `err`，方便日志记录。`c.ResponseError` 一般在控制器层 controllers 使用。

#### 正确示例

```go
// ✅ 正确 - 传入原始错误err
func (c *UserController) GetUser(ctx *gin.Context) {
    userID, _ := strconv.Atoi(ctx.Param("id"))
    
    user, err := c.userService.GetUserByID(userID)
    if err != nil {
        c.ResponseError(ctx, errors.CodeUserNotFound, err)  // 传入原始错误
        return
    }
    
    c.ResponseSuccess(ctx, user)
}

// ✅ 正确 - 数据库错误传入原始错误
func (c *ProductController) CreateProduct(ctx *gin.Context) {
    var req CreateProductRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ResponseError(ctx, errors.CodeParamError, err)  // 传入原始错误
        return
    }
    
    product, err := c.productService.CreateProduct(&req)
    if err != nil {
        c.ResponseError(ctx, errors.CodeDBError, err)  // 传入原始错误
        return
    }
    
    c.ResponseSuccess(ctx, product)
}
```

#### 错误示例

```go
// ❌ 错误 - 未传入原始错误
func (c *UserController) GetUser(ctx *gin.Context) {
    user, err := c.userService.GetUserByID(userID)
    if err != nil {
        c.ResponseError(ctx, errors.CodeUserNotFound, nil)  // 错误！应该传入err
        return
    }
}

// ❌ 错误 - 传入自定义错误而非原始错误
func (c *UserController) GetUser(ctx *gin.Context) {
    user, err := c.userService.GetUserByID(userID)
    if err != nil {
        c.ResponseError(ctx, errors.CodeUserNotFound, errors.New(errors.CodeUserNotFound, "用户不存在"))  // 错误！应该传入原始err
        return
    }
}

// ❌ 错误 - 在service层使用ResponseError
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        s.ResponseError(ctx, errors.CodeUserNotFound, err)  // 错误！ResponseError应该在controller层使用
        return nil, err
    }
    return user, nil
}
```

### 12.5 服务层错误处理规范

**强制要求**: services层使用官方的 `errors` 包返回错误，不使用 `fmt.Errorf`，一律返回到 controllers 层交给 `ResponseError` 处理。

#### 正确示例

```go
// ✅ 正确 - services层使用官方errors包，直接返回错误
import "errors"

func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    if userID <= 0 {
        return nil, errors.New("用户ID无效")
    }
    
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, err  // 直接返回原始错误，由controller层处理
    }
    
    if user == nil {
        return nil, errors.New("用户不存在")
    }
    
    return user, nil
}

// ✅ 正确 - services层直接返回错误，不做额外包装
func (s *OrderService) CreateOrder(order *models.Order) error {
    if order.ProductID <= 0 {
        return errors.New("商品ID无效")
    }
    
    if err := s.checkStock(order.ProductID, order.Quantity); err != nil {
        return err  // 直接返回原始错误
    }
    
    return s.orderRepo.Create(order)
}

// ✅ 正确 - controller层统一处理错误
func (c *OrderController) CreateOrder(ctx *gin.Context) {
    var req CreateOrderRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ResponseError(ctx, errors.CodeParamError, err)
        return
    }
    
    order, err := c.orderService.CreateOrder(&req)
    if err != nil {
        c.ResponseError(ctx, errors.CodeInternalError, err)  // 统一由ResponseError处理
        return
    }
    
    c.ResponseSuccess(ctx, order)
}
```

#### 错误示例

```go
// ❌ 错误 - services层使用自定义errors包
import "goshopadmin/backend/errors"

func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, errors.New(errors.CodeDBError, "查询失败")  // 错误！services层应该用官方errors包
    }
    return user, nil
}

// ❌ 错误 - services层使用fmt.Errorf
import "fmt"

func (s *OrderService) CreateOrder(order *models.Order) error {
    if err := s.checkStock(order.ProductID, order.Quantity); err != nil {
        return fmt.Errorf("库存检查失败: %w", err)  // 错误！不使用fmt.Errorf，直接返回原始错误
    }
    return s.orderRepo.Create(order)
}

// ❌ 错误 - services层返回nil错误但无数据
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    user, _ := s.userRepo.FindByID(userID)
    return user, nil  // 错误！应该处理错误
}

// ❌ 错误 - services层包装错误信息
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, errors.New("查询用户失败: " + err.Error())  // 不推荐，直接返回原始错误即可
    }
    return user, nil
}
```

---

## AI执行检查清单

在执行任何计划前，AI必须检查：

- [ ] 响应是否使用了 `ResponseSuccess` 或 `ResponseError` 方法（例外情况除外）
- [ ] 结构体整数类型是否使用了 `int` 而非 `uint`
- [ ] 数据库表设计是否避免了 UNSIGNED 并使用了 `utf8mb4_unicode_ci`
- [ ] **数据库表设计是否避免了外键约束（FOREIGN KEY）**
- [ ] **数据库状态字段是否使用了 ENUM 类型**
- [ ] **代码中是否使用了常量而非硬编码状态值**
- [ ] **错误码是否使用了常量而非硬编码数字**
- [ ] **新增错误码是否同步更新了 errors/code.go**
- [ ] 路由写法是否与参考文件风格一致
- [ ] 修改代码后是否重启了对应的Docker容器
- [ ] 修改API后是否更新了 `API接口和前端对接情况.md`
- [ ] 修改数据库后是否更新了 `数据库表结构说明.md`
- [ ] 用户命名是否符合规则8
- [ ] **新增状态后是否同步更新了 constants/constants.go**
- [ ] **日志记录是否使用了统一的日志工具类（utils.Info/Warn/Error）**
- [ ] **代码中是否避免使用TODO注释代替功能实现（除非用户明确要求）**
- [ ] **if语句是否避免了过长条件，保持可读性**
- [ ] **请求结构体是否定义在方法前面并添加了注释**
- [ ] **方法是否添加了功能注释**
- [ ] **控制器层ResponseError是否传入了原始错误err**
- [ ] **服务层是否使用了官方errors包**

---

## 违规示例汇总

以下代码是**不允许**的：

```go
// ❌ 错误：直接返回JSON，未使用统一响应
c.JSON(200, gin.H{"code": 0, "data": user})

// ❌ 错误：使用了uint
type Product struct {
    ID    uint    `gorm:"primaryKey"`
    Stock uint    `gorm:"default:0"`
}

// ❌ 错误：路由风格不一致
router.GET("/getUser", handler)  // 应该使用 /users
router.POST("/create_product", handler)  // 应该使用 /products

// ❌ 错误 - backend 项目使用了 customer
type Customer struct {  // backend 项目应该用 User
    ID int
}

// ❌ 错误 - shop-backend 项目使用了 user
type User struct {  // shop-backend 项目应该用 Customer
    ID int
}

// ❌ 错误 - 硬编码状态值
if status == "active" {  // 应该用常量 constants.StatusActive
    // ...
}
order.Status = "completed"  // 应该用常量 constants.OrderStatusCompleted

// ❌ 错误 - 使用数字状态值（无业务含义）
order.Status = 1  // 应该用ENUM和常量
product.Status = 0  // 应该用ENUM和常量

// ❌ 错误 - 硬编码错误码
return errors.New(4001, "参数错误")  // 应该用 errors.CodeParamError
s.ResponseError(ctx, 4040, err)       // 应该用 errors.CodeNotFound

// ❌ 错误 - 直接使用 fmt 打印
fmt.Println("用户登录成功")

// ❌ 错误 - 直接使用 log 包
log.Println("库存不足")

// ❌ 错误 - if条件过长
if order.PaymentStatus == "success" && order.Stock > 0 && order.Status == "pending" && order.User != nil {
    // 应该提取为有意义的变量或方法
}

// ❌ 错误 - 请求结构体定义在方法后面
func (c *AuthController) Login(ctx *gin.Context) {
    var req LoginRequest
}
type LoginRequest struct {  // 错误！应该定义在方法前面
    Username string
}

// ❌ 错误 - 方法没有注释
func (s *UserService) GetUserByID(userID int) (*models.User, error) {
    return s.userRepo.FindByID(userID)
}

// ❌ 错误 - ResponseError未传入原始错误
c.ResponseError(ctx, errors.CodeUserNotFound, nil)  // 应该传入err

// ❌ 错误 - services层使用自定义errors包
import "goshopadmin/backend/errors"
return errors.New(errors.CodeDBError, "查询失败")  // services层应该用官方errors包

// ❌ 错误 - services层使用fmt.Errorf
import "fmt"
return fmt.Errorf("库存检查失败: %w", err)  // services层不使用fmt.Errorf，直接返回原始错误
```

以下SQL是**不允许**的：

```sql
-- ❌ 错误 - 使用了外键约束
CREATE TABLE orders (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_orders_user_id (user_id),
    CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ❌ 错误 - 使用了UNSIGNED
CREATE TABLE products (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    stock INT UNSIGNED DEFAULT 0,
    PRIMARY KEY (id)
);

-- ❌ 错误 - 排序规则不正确
CREATE TABLE categories (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100),
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ❌ 错误 - 状态字段使用VARCHAR而非ENUM
CREATE TABLE products (
    id INT NOT NULL AUTO_INCREMENT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 应该用ENUM('active', 'inactive')
    PRIMARY KEY (id)
);

-- ❌ 错误 - 状态字段使用INT而非ENUM
CREATE TABLE orders (
    id INT NOT NULL AUTO_INCREMENT,
    status INT NOT NULL DEFAULT 0,  -- 应该用ENUM('pending', 'paid', ...)
    PRIMARY KEY (id)
);
```

---

*规则版本: 1.9*
*最后更新: 2026-03-28*
