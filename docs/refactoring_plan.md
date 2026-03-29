# 代码重构方案 - 统一架构版

## 概述

本文档提供将 `backend` 和 `shop-backend` 两个项目统一为一致架构的方案。

**统一后的架构目标**:
- 消除全局变量，采用依赖注入
- 结构体方法风格的 Controller
- 清晰的 Service 层
- 统一的项目结构和命名规范

---

## 当前代码状态分析

### 项目结构对比

```
backend/                    shop-backend/
├── config/                 ├── config/
│   ├── config.go           │   ├── config.go       # 全局 AppConfig
│   └── (无db.go)           │   └── db.go           # 全局 DB/Redis
├── controllers/            ├── controllers/
│   ├── base_controller.go  │   ├── base_controller.go
│   ├── auth_controller.go  │   ├── auth_controller.go  # 函数式
│   ├── user_controller.go  │   ├── customer_controller.go # 函数式
│   ├── role_controller.go  │   ├── cart_controller.go
│   ├── permission_controller.go  ├── order_controller.go
│   ├── merchant_controller.go    ├── product_controller.go
│   └── product_controller.go     └── payment_controller.go
├── services/               ├── services/           # ❌ 不存在
│   ├── auth_service.go     │
│   ├── merchant_service.go │
│   └── product_service.go  │
├── models/                 ├── models/
├── middleware/             ├── middleware/
├── routes/                 ├── routes/
│   └── routes.go           │   └── routes.go       # 函数式
├── utils/                  ├── utils/
└── main.go                 └── main.go
```

### 当前实现对比

| 特性 | backend | shop-backend |
|------|---------|--------------|
| **配置管理** | 全局 `AppConfig` | 全局 `AppConfig` |
| **数据库访问** | Service 注入 `*gorm.DB` | 全局 `config.DB` |
| **Controller 风格** | 结构体方法 | 函数式 |
| **Service 层** | ✅ 有 | ❌ 无 |
| **依赖注入** | 部分实现（DB 注入） | ❌ 无 |

### 两个项目共同存在的问题

1. **全局配置变量** - 两个项目都使用 `var AppConfig Config`
2. **Service 层依赖全局配置** - `auth_service.go` 直接访问 `config.AppConfig.JWTSecret`

---

## 统一后的目标架构

### 目录结构

```
project-root/
├── config/
│   ├── config.go           # 配置结构定义 + LoadConfig() (*Config, error)
│   └── db.go               # DBConnection 结构体 + InitDB(cfg) (*DBConnection, error)
├── controllers/            # 控制器层 - 结构体方法风格
│   ├── base_controller.go  # 基础控制器，提供通用方法
│   ├── auth_controller.go
│   └── ...
├── services/               # 服务层 - 业务逻辑
│   ├── auth_service.go
│   └── ...
├── models/                 # 模型层 - 数据结构
├── middleware/             # 中间件
├── routes/                 # 路由配置
│   └── routes.go           # SetupRouter(deps *Dependencies) *gin.Engine
├── utils/                  # 工具函数
└── main.go                 # 程序入口 - 组装所有依赖
```

### 架构层次

```
┌─────────────────────────────────────────────────────────────┐
│                        main.go                               │
│  1. 加载配置: cfg, _ := config.LoadConfig()                  │
│  2. 初始化DB: conn, _ := config.InitDB(cfg)                  │
│  3. 创建Service: svc := services.NewXxxService(conn.DB, cfg) │
│  4. 创建Controller: ctrl := controllers.NewXxxController(svc)│
│  5. 设置路由: router := routes.SetupRouter(deps)             │
│  6. 启动服务: router.Run(...)                                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      routes/routes.go                        │
│  type Dependencies struct {                                  │
│      AuthController *controllers.AuthController              │
│      ...                                                     │
│  }                                                           │
│  func SetupRouter(deps *Dependencies) *gin.Engine            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   controllers/*.go                           │
│  type AuthController struct {                                │
│      authService *services.AuthService                       │
│  }                                                           │
│  func (c *AuthController) Login(ctx *gin.Context)            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     services/*.go                            │
│  type AuthService struct {                                   │
│      db *gorm.DB                                             │
│      jwtSecret string                                        │
│  }                                                           │
│  func (s *AuthService) Login(...) (string, error)            │
└─────────────────────────────────────────────────────────────┘
```

---

## 重构方案详情

### 阶段 1: 配置层重构 (config/)

#### 目标: 消除全局变量，改为实例化模式

**config/config.go**
```go
package config

import (
    "fmt"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    ServerPort    string
    DBHost        string
    DBPort        string
    DBUser        string
    DBPassword    string
    DBName        string
    RedisHost     string
    RedisPort     string
    RedisPassword string
    RedisDB       int
    JWTSecret     string
    JWTExpireHour int
}

// LoadConfig 返回配置实例（非全局变量）
func LoadConfig() (*Config, error) {
    _ = godotenv.Load()

    jwtExpireHour, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOUR", "24"))
    redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

    return &Config{
        ServerPort:    getEnv("SERVER_PORT", "8080"),
        DBHost:        getEnv("DB_HOST", "localhost"),
        DBPort:        getEnv("DB_PORT", "3306"),
        DBUser:        getEnv("DB_USER", "root"),
        DBPassword:    getEnv("DB_PASSWORD", "password"),
        DBName:        getEnv("DB_NAME", "goshopadmin"),
        RedisHost:     getEnv("REDIS_HOST", "localhost"),
        RedisPort:     getEnv("REDIS_PORT", "6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       redisDB,
        JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
        JWTExpireHour: jwtExpireHour,
    }, nil
}

func (c *Config) GetDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) GetRedisAddr() string {
    return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
```

**config/db.go**
```go
package config

import (
    "context"
    "fmt"
    "log"

    "github.com/go-redis/redis/v8"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// DBConnection 封装数据库连接（替代全局变量）
type DBConnection struct {
    DB    *gorm.DB
    Redis *redis.Client
}

// InitDB 初始化数据库连接，返回实例
func InitDB(cfg *Config) (*DBConnection, error) {
    dsn := cfg.GetDSN()
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    // 自动迁移
    // 注意：具体模型在各自项目的 main.go 中传入
    log.Println("Database connected successfully")

    // 连接 Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     cfg.GetRedisAddr(),
        Password: cfg.RedisPassword,
        DB:       cfg.RedisDB,
    })

    ctx := context.Background()
    if err := redisClient.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect redis: %w", err)
    }

    log.Println("Redis connected successfully")

    return &DBConnection{
        DB:    db,
        Redis: redisClient,
    }, nil
}
```

---

### 阶段 2: 服务层重构 (services/)

#### 目标: Service 不再依赖全局配置，通过构造函数注入

**services/auth_service.go**
```go
package services

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
    "your-project/models"
)

// AuthService 认证服务
type AuthService struct {
    db            *gorm.DB
    jwtSecret     string
    jwtExpireHour int
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpireHour int) *AuthService {
    return &AuthService{
        db:            db,
        jwtSecret:     jwtSecret,
        jwtExpireHour: jwtExpireHour,
    }
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
    var user models.User
    result := s.db.Where("username = ?", username).First(&user)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil, errors.New("用户名或密码错误")
        }
        return "", nil, result.Error
    }

    // 验证密码
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return "", nil, errors.New("用户名或密码错误")
    }

    // 生成 Token
    token, err := s.generateToken(user.ID)
    if err != nil {
        return "", nil, err
    }

    return token, &user, nil
}

func (s *AuthService) generateToken(userID uint) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * time.Duration(s.jwtExpireHour)).Unix(),
    })
    return token.SignedString([]byte(s.jwtSecret))
}
```

---

### 阶段 3: 控制器层重构 (controllers/)

#### 目标: 统一为结构体方法风格，注入 Service

**controllers/base_controller.go**
```go
package controllers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// BaseController 基础控制器
type BaseController struct{}

// ResponseSuccess 返回成功响应
func (c *BaseController) ResponseSuccess(ctx *gin.Context, data interface{}) {
    ctx.JSON(http.StatusOK, data)
}

// ResponseError 返回错误响应
func (c *BaseController) ResponseError(ctx *gin.Context, code int, message string) {
    ctx.JSON(code, gin.H{"error": message})
}

// GetUserID 从上下文获取用户ID
func (c *BaseController) GetUserID(ctx *gin.Context) (uint, bool) {
    userID, exists := ctx.Get("userID")
    if !exists {
        return 0, false
    }
    return userID.(uint), true
}
```

**controllers/auth_controller.go**
```go
package controllers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "your-project/services"
)

// AuthController 认证控制器
type AuthController struct {
    BaseController
    authService *services.AuthService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(authService *services.AuthService) *AuthController {
    return &AuthController{
        authService: authService,
    }
}

// LoginRequest 登录请求结构
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
    var req LoginRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.ResponseError(ctx, http.StatusBadRequest, "参数错误")
        return
    }

    token, user, err := c.authService.Login(req.Username, req.Password)
    if err != nil {
        c.ResponseError(ctx, http.StatusUnauthorized, err.Error())
        return
    }

    c.ResponseSuccess(ctx, gin.H{
        "code":    200,
        "message": "登录成功",
        "data": gin.H{
            "token":    token,
            "user_id":  user.ID,
            "username": user.Username,
        },
    })
}
```

---

### 阶段 4: 路由层重构 (routes/)

#### 目标: 支持依赖注入

**routes/routes.go**
```go
package routes

import (
    "github.com/gin-gonic/gin"
    "your-project/controllers"
    "your-project/middleware"
)

// Dependencies 包含所有依赖
type Dependencies struct {
    AuthController *controllers.AuthController
    UserController *controllers.UserController
    // 其他控制器...
}

// SetupRouter 设置路由
func SetupRouter(deps *Dependencies) *gin.Engine {
    router := gin.Default()

    // 跨域中间件
    router.Use(middleware.CORS())

    // 健康检查
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API路由组
    api := router.Group("/api")
    api.Use(middleware.RequestLogger())
    {
        // 认证路由
        auth := api.Group("/auth")
        {
            auth.POST("/login", deps.AuthController.Login)
            auth.POST("/logout", deps.AuthController.Logout)
        }

        // 需要认证的路由
        authorized := api.Group("/")
        authorized.Use(middleware.Auth())
        {
            authorized.GET("/users", deps.UserController.GetUsers)
            // 其他需要认证的路由...
        }
    }

    return router
}
```

---

### 阶段 5: 程序入口重构 (main.go)

#### 目标: 组装所有依赖

**main.go**
```go
package main

import (
    "log"
    "your-project/config"
    "your-project/controllers"
    "your-project/models"
    "your-project/routes"
    "your-project/services"
)

func main() {
    // 1. 加载配置
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 2. 初始化数据库连接
    conn, err := config.InitDB(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // 3. 自动迁移模型
    conn.DB.AutoMigrate(
        &models.User{},
        &models.Role{},
        // 其他模型...
    )

    // 4. 创建服务层实例（依赖注入）
    authService := services.NewAuthService(conn.DB, cfg.JWTSecret, cfg.JWTExpireHour)
    userService := services.NewUserService(conn.DB)
    // 其他服务...

    // 5. 创建控制器实例（依赖注入）
    deps := &routes.Dependencies{
        AuthController: controllers.NewAuthController(authService),
        UserController: controllers.NewUserController(userService),
        // 其他控制器...
    }

    // 6. 设置路由
    router := routes.SetupRouter(deps)

    // 7. 启动服务器
    log.Printf("Server starting on port %s", cfg.ServerPort)
    if err := router.Run(":" + cfg.ServerPort); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

---

## 重构实施计划

### 项目优先级

1. **先重构 `shop-backend`** - 变化更大，需要新增 Service 层
2. **后调整 `backend`** - 主要是消除全局配置依赖

### 详细实施步骤

#### shop-backend 重构步骤

| 阶段 | 任务 | 文件 | 优先级 |
|------|------|------|--------|
| 1 | 修改配置层 | `config/config.go` | 高 |
| 1 | 修改数据库初始化 | `config/db.go` | 高 |
| 2 | 创建 AuthService | `services/auth_service.go` | 高 |
| 2 | 创建 CustomerService | `services/customer_service.go` | 高 |
| 3 | 重构 AuthController | `controllers/auth_controller.go` | 高 |
| 3 | 重构 CustomerController | `controllers/customer_controller.go` | 中 |
| 4 | 重构路由层 | `routes/routes.go` | 中 |
| 5 | 重构 main.go | `main.go` | 中 |
| 6 | 逐步迁移其他模块 | Cart/Order/Product | 低 |

#### backend 调整步骤

| 阶段 | 任务 | 文件 | 优先级 |
|------|------|------|--------|
| 1 | 修改配置层 | `config/config.go` | 高 |
| 2 | 修改 Service 层 | `services/*.go` | 高 |
| 3 | 修改 Controller 构造函数 | `controllers/*.go` | 中 |
| 4 | 修改路由层 | `routes/routes.go` | 中 |
| 5 | 重构 main.go | `main.go` | 中 |

---

## 代码对比：重构前后

### 配置层对比

| 项目 | 重构前 | 重构后 |
|------|--------|--------|
| **返回值** | `LoadConfig() error` | `LoadConfig() (*Config, error)` |
| **配置访问** | `config.AppConfig.Xxx` | 通过实例 `cfg.Xxx` |
| **数据库** | `config.DB` 全局变量 | `conn.DB` 实例 |

### Controller 对比

**重构前 (shop-backend)**:
```go
func Register(c *gin.Context) {
    // 直接访问全局变量
    result := config.DB.Where("username = ?", req.Username).First(&existingUser)
}
```

**重构后**:
```go
type AuthController struct {
    authService *services.AuthService
}

func (c *AuthController) Register(ctx *gin.Context) {
    // 通过 Service 调用
    user, err := c.authService.Register(...)
}
```

### Service 对比

**重构前 (backend)**:
```go
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
    // 直接访问全局配置
    token, err := utils.GenerateToken(user.ID, config.AppConfig.JWTSecret)
}
```

**重构后**:
```go
func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpireHour int) *AuthService {
    return &AuthService{
        db:            db,
        jwtSecret:     jwtSecret,      // 通过构造函数注入
        jwtExpireHour: jwtExpireHour,  // 通过构造函数注入
    }
}
```

---

## 测试示例

### 单元测试

```go
package controllers

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/models"
    "your-project/services"
)

func TestAuthController_Login(t *testing.T) {
    // 1. 创建内存数据库
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    assert.NoError(t, err)
    db.AutoMigrate(&models.User{})

    // 2. 插入测试数据
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
    db.Create(&models.User{
        ID:       1,
        Username: "test",
        Password: string(hashedPassword),
    })

    // 3. 创建服务和控制器（依赖注入）
    authService := services.NewAuthService(db, "secret", 24)
    controller := NewAuthController(authService)

    // 4. 设置 Gin 测试上下文
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    ctx, _ := gin.CreateTestContext(w)
    ctx.Request = httptest.NewRequest("POST", "/login", 
        strings.NewReader(`{"username":"test","password":"password"}`))
    ctx.Request.Header.Set("Content-Type", "application/json")

    // 5. 执行测试
    controller.Login(ctx)

    // 6. 验证结果
    assert.Equal(t, http.StatusOK, w.Code)
}
```

---

## 重构收益

| 方面 | 改进前 | 改进后 |
|------|--------|--------|
| **可测试性** | ❌ 难测试（全局依赖） | ✅ 易测试（依赖注入） |
| **代码耦合** | ❌ 高耦合 | ✅ 低耦合 |
| **并发安全** | ❌ 全局变量风险 | ✅ 实例隔离 |
| **多数据库支持** | ❌ 困难 | ✅ 容易 |
| **事务管理** | ❌ 不灵活 | ✅ 灵活控制 |
| **代码可读性** | ⚠️ 依赖隐藏 | ✅ 依赖清晰 |
| **架构一致性** | ❌ 两个项目不一致 | ✅ 统一架构 |

---

## 注意事项

1. **渐进式重构**: 不要一次性重构所有代码，逐步迁移降低风险
2. **保持接口兼容性**: 确保前端 API 调用不受影响
3. **边重构边测试**: 确保每个阶段功能正常
4. **优先重构核心模块**: Auth/User 模块优先，验证架构可行性后再扩展
5. **代码审查**: 确保团队成员理解新架构

---

## 依赖关系图

```
┌─────────────────────────────────────────────────────────────────┐
│                            main.go                               │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐  │
│  │  LoadConfig  │  │    InitDB    │  │   创建 Service 实例    │  │
│  │              │  │              │  │                       │  │
│  │ 返回 *Config  │  │ 返回 *DBConn │  │ AuthService           │  │
│  │              │  │ (DB+Redis)   │  │ UserService           │  │
│  └──────┬───────┘  └──────┬───────┘  │ ...                   │  │
│         │                 │          └───────────┬───────────┘  │
│         │                 │                      │              │
│         └─────────────────┴──────────────────────┘              │
│                           │                                      │
│                           ▼                                      │
│               ┌───────────────────────┐                          │
│               │   创建 Controller 实例 │                          │
│               │  NewAuthController()  │                          │
│               │ NewUserController()   │                          │
│               └───────────┬───────────┘                          │
│                           │                                      │
│                           ▼                                      │
│               ┌───────────────────────┐                          │
│               │    routes.SetupRouter │                          │
│               │    (依赖注入 deps)     │                          │
│               └───────────┬───────────┘                          │
│                           │                                      │
│                           ▼                                      │
│               ┌───────────────────────┐                          │
│               │   router.Run(...)     │                          │
│               └───────────────────────┘                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 总结

本方案旨在统一 `backend` 和 `shop-backend` 两个项目的架构，核心改进包括：

1. **消除全局变量**: 配置和数据库连接通过实例传递
2. **依赖注入**: Service 和 Controller 通过构造函数注入依赖
3. **统一代码风格**: 两个项目采用一致的架构模式
4. **提高可测试性**: 便于编写单元测试和集成测试

实施顺序建议：先完成 `shop-backend` 的重构（变化更大），再调整 `backend` 以匹配统一架构。
