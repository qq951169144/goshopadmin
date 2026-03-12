# C端商城实现计划

## 一、项目结构设计

### 1. 目录结构

```
goshopadmin/
├── backend/            # 现有Go后端服务
├── docker/             # Docker配置
│   ├── mysql/
│   ├── nginx/
│   └── docker-compose.yml
├── frontend/           # 现有后台管理前端
├── shop-frontend/      # 新增C端商城前端
│   ├── public/
│   ├── src/
│   │   ├── api/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── views/
│   │   ├── router/
│   │   ├── store/
│   │   ├── utils/
│   │   ├── App.vue
│   │   └── main.js
│   ├── .env.example
│   ├── .gitignore
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── shop-backend/       # 新增C端商城后端API
│   ├── config/
│   ├── controllers/
│   ├── middleware/
│   ├── models/
│   ├── routes/
│   ├── services/
│   ├── utils/
│   ├── .env.example
│   ├── go.mod
│   └── main.go
└── 开发与部署流程.md
```

### 2. 技术栈选择

| 分类 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 前端框架 | Vue.js | 3.x | 响应式前端框架 |
| 构建工具 | Vite | 5.x | 快速构建工具 |
| 状态管理 | Pinia | 2.x | 轻量级状态管理 |
| 路由 | Vue Router | 4.x | 前端路由管理 |
| HTTP客户端 | Axios | 1.x | API请求 |
| 后端语言 | Go | 1.25.8 | 高性能后端语言 |
| 数据库 | MySQL | 8.0 | 关系型数据库 |
| 缓存 | Redis | 7.0 | 用于购物车和会话管理 |
| 容器化 | Docker | 20.x+ | 应用容器化 |
| 反向代理 | Nginx | 1.20 | 静态文件服务和负载均衡 |

## 二、核心功能模块设计

### 1. 购物车模块 ✅

#### 1.1 无登录状态
- **实现方式**：使用LocalStorage存储购物车数据
- **数据结构**：
  ```javascript
  {
    "cart_id": "local_123456",
    "items": [
      {
        "product_id": 1,
        "quantity": 2,
        "price": 99.99,
        "sku": "red-medium"
      }
    ],
    "created_at": "2026-03-12T10:00:00Z"
  }
  ```
- **同步机制**：用户登录后，将本地购物车数据同步到服务器

#### 1.2 有登录状态
- **实现方式**：使用Redis存储购物车数据，关联用户ID
- **数据结构**：
  ```go
  type CartItem struct {
    ProductID  uint    `json:"product_id"`
    Quantity   int     `json:"quantity"`
    Price      float64 `json:"price"`
    SKU        string  `json:"sku"`
  }
  
  type Cart struct {
    UserID    uint        `json:"user_id"`
    Items     []CartItem  `json:"items"`
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
  }
  ```
- **API接口**：
  - `GET /api/cart` - 获取购物车
  - `POST /api/cart/items` - 添加商品到购物车
  - `PUT /api/cart/items/:id` - 更新商品数量
  - `DELETE /api/cart/items/:id` - 移除商品
  - `POST /api/cart/sync` - 同步本地购物车

### 2. 支付模块（伪代码实现） ✅

#### 2.1 订单创建
- **API接口**：`POST /api/orders`
- **响应数据**：
  ```json
  {
    "order_id": "ORD202603120001",
    "amount": 199.98,
    "payment_url": "/api/payment/fake-pay?order_id=ORD202603120001",
    "status": "pending"
  }
  ```

#### 2.2 伪支付流程
1. 客户端跳转到 `payment_url`
2. 服务端返回支付成功页面
3. 服务端模拟支付回调：`POST /api/payment/callback`
4. 服务端更新订单状态为已支付

#### 2.3 支付回调接口
- **API接口**：`POST /api/payment/callback`
- **请求数据**：
  ```json
  {
    "order_id": "ORD202603120001",
    "transaction_id": "TRX1234567890",
    "status": "success",
    "amount": 199.98
  }
  ```

### 3. 验证码模块 ✅

#### 3.1 验证码生成
- **实现方式**：使用Go的图像处理库生成验证码
- **API接口**：`GET /api/captcha`
- **响应**：返回验证码图片（PNG格式）
- **存储**：验证码值存储在Redis中，设置5分钟过期

#### 3.2 验证码验证
- **API接口**：`POST /api/captcha/verify`
- **请求数据**：
  ```json
  {
    "captcha_id": "cap_123456",
    "value": "1234"
  }
  ```
- **响应**：
  ```json
  {
    "valid": true
  }
  ```

### 4. 用户模块 ✅

#### 4.1 注册/登录
- **API接口**：
  - `POST /api/auth/register` - 注册
  - `POST /api/auth/login` - 登录
  - `POST /api/auth/logout` - 登出

#### 4.2 个人中心
- **API接口**：
  - `GET /api/user/profile` - 获取个人信息
  - `PUT /api/user/profile` - 更新个人信息
  - `GET /api/user/orders` - 获取订单列表

### 5. 商品模块 ✅

#### 5.1 商品列表
- **API接口**：`GET /api/products`
- **参数**：
  - `page` - 页码
  - `limit` - 每页数量
  - `category` - 分类ID
  - `keyword` - 搜索关键词

#### 5.2 商品详情
- **API接口**：`GET /api/products/:id`

## 三、环境配置

### 1. 开发环境

#### 1.1 Docker配置
- 修改 `docker-compose.yml`，添加C端前端和后端服务
- 服务配置：
  - `shop-frontend`：端口 3001
  - `shop-backend`：端口 8081

#### 1.2 前端配置
- **Vite配置**：
  ```javascript
  // shop-frontend/vite.config.js
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://shop-backend:8081',
        changeOrigin: true
      }
    }
  }
  ```

### 2. 生产环境

#### 2.1 Nginx配置
- **配置文件**：`docker/nginx/nginx.conf`
- **配置内容**：
  ```nginx
  # 后台管理系统
  server {
    listen 80;
    server_name admin.example.com;

    root /usr/share/nginx/html/admin;
    index index.html;

    location / {
      try_files $uri $uri/ /index.html;
    }

    location /api/ {
      proxy_pass http://backend:8080/api/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }

  # C端商城
  server {
    listen 80;
    server_name shop.example.com;

    root /usr/share/nginx/html/shop;
    index index.html;

    location / {
      try_files $uri $uri/ /index.html;
    }

    location /api/ {
      proxy_pass http://shop-backend:8081/api/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }
  }
  ```

#### 2.2 静态文件存储
- 后台管理系统：`/usr/share/nginx/html/admin`
- C端商城：`/usr/share/nginx/html/shop`

## 四、数据库设计

### 1. 购物车表

```sql
CREATE TABLE `carts` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned DEFAULT NULL,
  `session_id` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_session_id` (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `cart_items` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `cart_id` int unsigned NOT NULL,
  `product_id` int unsigned NOT NULL,
  `quantity` int NOT NULL DEFAULT '1',
  `price` decimal(10,2) NOT NULL,
  `sku` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_cart_id` (`cart_id`),
  KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. 订单表

```sql
CREATE TABLE `orders` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `order_id` varchar(50) NOT NULL,
  `user_id` int unsigned NOT NULL,
  `amount` decimal(10,2) NOT NULL,
  `status` enum('pending','paid','shipping','delivered','cancelled') NOT NULL DEFAULT 'pending',
  `payment_method` varchar(50) DEFAULT NULL,
  `transaction_id` varchar(100) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `order_items` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `order_id` int unsigned NOT NULL,
  `product_id` int unsigned NOT NULL,
  `quantity` int NOT NULL,
  `price` decimal(10,2) NOT NULL,
  `sku` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 五、实现步骤

### 1. 项目初始化

1. **创建目录结构**：
   ```bash
   mkdir -p shop-frontend/src/{api,assets,components,views,router,store,utils}
   mkdir -p shop-backend/{config,controllers,middleware,models,routes,services,utils}
   ```

2. **初始化前端项目**：
   ```bash
   cd shop-frontend
   npm create vite@latest . -- --template vue
   npm install vue-router@4 pinia axios
   ```

3. **初始化后端项目**：
   ```bash
   cd shop-backend
   go mod init shop-backend
   go get github.com/gin-gonic/gin github.com/go-redis/redis/v8 github.com/go-sql-driver/mysql github.com/golang-jwt/jwt/v5 github.com/joho/godotenv
   ```

### 2. 配置文件

1. **前端配置**：
   - 创建 `.env.example` 文件
   - 配置 `vite.config.js`

2. **后端配置**：
   - 创建 `.env.example` 文件
   - 配置 `config/config.go`

3. **Docker配置**：
   - 修改 `docker-compose.yml`，添加C端服务
   - 修改 `docker/nginx/nginx.conf`

### 3. 核心功能实现

1. **购物车功能**：
   - 前端：购物车组件、本地存储管理
   - 后端：购物车API、Redis存储

2. **支付功能**：
   - 前端：支付流程、订单确认
   - 后端：订单创建、伪支付接口、回调处理

3. **验证码功能**：
   - 后端：验证码生成、验证逻辑
   - 前端：验证码组件、表单验证

4. **用户功能**：
   - 前端：注册/登录表单、个人中心
   - 后端：用户API、JWT认证

5. **商品功能**：
   - 前端：商品列表、详情页
   - 后端：商品API、搜索功能

### 4. 测试与部署

1. **开发环境测试**：
   - 启动Docker容器：`cd docker && docker-compose up -d`
   - 访问C端商城：`http://localhost:3001`
   - 测试API接口

2. **生产环境部署**：
   - 构建前端：`cd shop-frontend && npm run build`
   - 构建后端：`cd shop-backend && go build -o shop-backend-server main.go`
   - 启动生产容器：`cd docker && NODE_ENV=production RESTART_POLICY=always docker-compose up -d`

## 六、技术要点

1. **购物车实现**：
   - 无登录状态使用LocalStorage
   - 登录状态使用Redis
   - 登录时自动同步本地购物车

2. **支付流程**：
   - 伪代码实现，模拟支付过程
   - 保持真实的支付流程体验

3. **验证码**：
   - 使用Go的图像处理库生成验证码
   - Redis存储验证码值，设置过期时间

4. **环境配置**：
   - 开发环境：独立的前端容器
   - 生产环境：Nginx静态文件服务

5. **安全性**：
   - JWT token认证
   - 密码加密存储
   - 防止SQL注入
   - CORS配置

## 七、预期成果

1. **功能完整**：实现购物车、支付、验证码等核心功能
2. **用户体验**：流畅的购物流程，支持无登录购物
3. **技术架构**：清晰的前后端分离架构
4. **部署便捷**：Docker容器化部署，支持开发和生产环境

## 八、风险评估

1. **风险点**：购物车数据同步可能出现冲突
   - **解决方案**：使用乐观锁，或在同步时进行冲突检测

2. **风险点**：伪支付流程可能影响用户体验
   - **解决方案**：在开发环境明确标识为测试模式

3. **风险点**：验证码识别难度
   - **解决方案**：调整验证码复杂度，确保可读性

4. **风险点**：生产环境部署配置错误
   - **解决方案**：详细的部署文档，配置检查脚本

## 九、后续优化方向

1. **性能优化**：
   - 前端：代码分割、懒加载
   - 后端：数据库索引优化、缓存策略

2. **功能扩展**：
   - 优惠券系统
   - 评价系统
   - 物流跟踪

3. **安全性**：
   - HTTPS配置
   - 防止XSS攻击
   - 防止CSRF攻击

4. **用户体验**：
   - 响应式设计
   - 加载动画
   - 错误处理

## 十、实现状态

### 已完成的功能

1. **用户模块**：
   - ✅ 注册/登录功能
   - ✅ 个人信息管理
   - ✅ 订单列表查询

2. **购物车模块**：
   - ✅ 无登录状态（LocalStorage存储）
   - ✅ 有登录状态（数据库存储）
   - ✅ 购物车同步功能

3. **订单模块**：
   - ✅ 订单创建
   - ✅ 订单详情查询

4. **支付模块**：
   - ✅ 伪支付流程
   - ✅ 支付回调处理

5. **验证码模块**：
   - ✅ 验证码生成
   - ✅ 验证码验证

6. **商品模块**：
   - ✅ 商品列表
   - ✅ 商品详情

### 技术实现

1. **后端**：
   - ✅ Go 1.25.8 + Gin框架
   - ✅ MySQL数据库
   - ✅ Redis缓存
   - ✅ JWT认证

2. **前端**：
   - ✅ Vue 3 + Vite
   - ✅ Pinia状态管理
   - ✅ Vue Router路由
   - ✅ Axios API请求

3. **环境配置**：
   - ✅ Docker容器化
   - ✅ Nginx反向代理

---

本计划详细说明了C端商城的实现方案，包括项目结构、技术栈选择、功能模块设计、环境配置等内容。按照此计划实施，可以快速构建一个功能完整的C端商城系统。所有核心功能已实现，系统可以正常运行。