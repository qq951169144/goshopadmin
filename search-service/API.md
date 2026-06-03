# Search Service API 接口文档

搜索服务提供基于 Elasticsearch 的商品、订单、用户、客户搜索功能，以及搜索建议和健康检查接口。

## 通用说明

### 响应格式

所有接口统一返回以下 JSON 格式：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `code` | `int` | 业务状态码，0 表示成功，非 0 表示失败 |
| `message` | `string` | 提示信息 |
| `data` | `object` | 返回数据，失败时为 `null` |

### 错误码

| 错误码 | 常量名 | 说明 |
| :--- | :--- | :--- |
| 0 | `CodeSuccess` | 成功 |
| 4001 | `CodeParamError` | 参数错误 |
| 4002 | `CodeParamMissing` | 参数缺失 |
| 4003 | `CodeParamInvalid` | 参数格式无效 |
| 4004 | `CodeParamOutOfRange` | 参数超出范围 |
| 4040 | `CodeNotFound` | 资源不存在 |
| 4080 | `CodeSearchError` | 搜索服务错误 |
| 4081 | `CodeSearchTimeout` | 搜索请求超时 |
| 4082 | `CodeSearchRateLimited` | 搜索请求过于频繁 |
| 4083 | `CodeESUnavailable` | 搜索服务暂不可用 |
| 5000 | `CodeInternalError` | 内部错误 |
| 5001 | `CodeDBError` | 数据库错误 |
| 5002 | `CodeESError` | 搜索引擎错误 |

### 限流说明

搜索 API（`/api/search/*`）限制为 50 QPS，超出限制返回错误码 `4082`。

---

## 1. 商品搜索

**接口路径**: `GET /api/search/products`

**功能描述**: 搜索商品，支持关键词搜索（IK 中文分词）、分类/商户/状态/价格区间筛选、排序和分页。搜索结果中匹配的关键词会高亮显示（`<em>` 标签包裹）。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `keyword` | `string` | 否 | 搜索关键词，使用 IK 分词匹配商品名称和描述 |
| `category_id` | `int` | 否 | 分类ID，精确筛选指定分类 |
| `merchant_id` | `int` | 否 | 商户ID，精确筛选指定商户 |
| `status` | `string` | 否 | 商品状态：`active` 上架 / `inactive` 下架 |
| `min_price` | `float` | 否 | 最低价格筛选 |
| `max_price` | `float` | 否 | 最高价格筛选 |
| `sort` | `string` | 否 | 排序字段：`relevance`（相关度，默认）/ `price`（价格）/ `sales`（销量）/ `created_at`（创建时间） |
| `order` | `string` | 否 | 排序方向：`desc`（降序，默认）/ `asc`（升序） |
| `page` | `int` | 否 | 页码，从 1 开始，默认 1 |
| `page_size` | `int` | 否 | 每页记录数，默认 20，最大 100 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 156,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "name": "<em>苹果</em>手机 iPhone 15 Pro",
        "description": "全新<em>苹果</em>旗舰手机，搭载A17芯片",
        "category_id": 10,
        "category_name": "手机",
        "merchant_id": 5,
        "merchant_name": "苹果官方旗舰店",
        "status": "active",
        "min_price": 7999.00,
        "max_price": 12999.00,
        "main_image": "https://example.com/iphone15.jpg",
        "sales": 5680,
        "skus": [
          {
            "id": 1,
            "product_id": 1,
            "sku_name": "暗紫色-256GB",
            "price": 7999.00,
            "stock": 100,
            "image": "https://example.com/iphone15-purple.jpg",
            "spec_values": {
              "颜色": "暗紫色",
              "存储": "256GB"
            }
          }
        ],
        "specs": [
          {
            "spec_name": "颜色",
            "spec_values": ["暗紫色", "金色", "银色", "黑色"]
          },
          {
            "spec_name": "存储",
            "spec_values": ["256GB", "512GB", "1TB"]
          }
        ],
        "created_at": "2026-01-15T10:30:00Z",
        "updated_at": "2026-06-01T08:00:00Z"
      }
    ]
  }
}
```

### 错误响应示例

```json
{
  "code": 4080,
  "message": "搜索服务错误: Elasticsearch 客户端未初始化",
  "data": null
}
```

---

## 2. 订单搜索

**接口路径**: `GET /api/search/orders`

**功能描述**: 搜索订单，支持关键词搜索（订单号精确匹配 + 订单明细商品名称模糊匹配）、多状态筛选、时间范围和金额范围筛选。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `keyword` | `string` | 否 | 搜索关键词，匹配订单号（精确）或商品名称（模糊） |
| `customer_id` | `int` | 否 | 客户ID，筛选指定客户的订单 |
| `merchant_id` | `int` | 否 | 商户ID，筛选指定商户的订单 |
| `status` | `string` | 否 | 订单状态：`pending` / `paid` / `shipped` / `completed` / `cancelled` |
| `payment_status` | `string` | 否 | 支付状态：`pending` / `success` / `failed` |
| `shipping_status` | `string` | 否 | 物流状态：`pending` / `shipped` / `delivered` / `returned` |
| `start_date` | `string` | 否 | 开始日期，格式 `2006-01-02` |
| `end_date` | `string` | 否 | 结束日期，格式 `2006-01-02` |
| `min_amount` | `float` | 否 | 最小金额筛选 |
| `max_amount` | `float` | 否 | 最大金额筛选 |
| `sort` | `string` | 否 | 排序字段：`relevance`（默认）/ `total_amount` / `created_at` |
| `order` | `string` | 否 | 排序方向：`desc`（默认）/ `asc` |
| `page` | `int` | 否 | 页码，默认 1 |
| `page_size` | `int` | 否 | 每页记录数，默认 20，最大 100 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 42,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1001,
        "order_no": "ORD202606030001",
        "customer_id": 50,
        "customer_name": "张三",
        "merchant_id": 5,
        "status": "paid",
        "payment_status": "success",
        "shipping_status": "pending",
        "total_amount": 7999.00,
        "items": [
          {
            "id": 2001,
            "product_id": 1,
            "product_name": "苹果手机 iPhone 15 Pro",
            "sku_id": 1,
            "sku_name": "暗紫色-256GB",
            "price": 7999.00,
            "quantity": 1,
            "subtotal": 7999.00
          }
        ],
        "created_at": "2026-06-03T14:30:00Z",
        "updated_at": "2026-06-03T14:31:00Z"
      }
    ]
  }
}
```

### 错误响应示例

```json
{
  "code": 4081,
  "message": "搜索请求超时: 订单搜索超时",
  "data": null
}
```

---

## 3. 用户搜索

**接口路径**: `GET /api/search/users`

**功能描述**: 搜索后台管理系统用户，支持关键词搜索（用户名精确匹配 + 邮箱模糊匹配）、角色和状态筛选。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `keyword` | `string` | 否 | 搜索关键词，匹配用户名（精确）或邮箱（模糊） |
| `role_id` | `int` | 否 | 角色ID，筛选指定角色的用户 |
| `status` | `string` | 否 | 用户状态：`active` 启用 / `inactive` 禁用 |
| `page` | `int` | 否 | 页码，默认 1 |
| `page_size` | `int` | 否 | 每页记录数，默认 20，最大 100 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 8,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "phone": "13800138000",
        "role_id": 1,
        "role_name": "超级管理员",
        "status": "active",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 错误响应示例

```json
{
  "code": 4080,
  "message": "搜索服务错误: Elasticsearch 客户端未初始化",
  "data": null
}
```

---

## 4. 客户搜索

**接口路径**: `GET /api/search/customers`

**功能描述**: 搜索 C 端商城客户，支持关键词搜索（用户名/手机号精确匹配 + 邮箱模糊匹配 + 昵称 IK 分词搜索）、状态筛选。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `keyword` | `string` | 否 | 搜索关键词，匹配用户名/手机号（精确）或邮箱（模糊）或昵称（IK 分词） |
| `status` | `string` | 否 | 客户状态：`active` 启用 / `inactive` 禁用 |
| `page` | `int` | 否 | 页码，默认 1 |
| `page_size` | `int` | 否 | 每页记录数，默认 20，最大 100 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 25,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13900139000",
        "nickname": "小明",
        "avatar": "https://example.com/avatar/1.jpg",
        "status": "active",
        "created_at": "2026-02-15T10:00:00Z",
        "updated_at": "2026-05-20T14:30:00Z"
      }
    ]
  }
}
```

### 错误响应示例

```json
{
  "code": 4080,
  "message": "搜索服务错误: Elasticsearch 客户端未初始化",
  "data": null
}
```

---

## 5. 搜索建议

**接口路径**: `GET /api/search/suggest`

**功能描述**: 根据用户输入的前缀返回搜索建议词，用于搜索框自动补全功能。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `prefix` | `string` | 是 | 搜索前缀，至少 1 个字符 |
| `type` | `string` | 是 | 建议类型：`product`（商品名称）/ `order`（订单号）/ `user`（用户名）/ `customer`（客户昵称） |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "suggestions": [
      "苹果手机 iPhone 15 Pro",
      "苹果手机 iPhone 15",
      "苹果手机 iPhone 14"
    ]
  }
}
```

### 错误响应示例

```json
{
  "code": 4003,
  "message": "参数格式无效: type 参数必须是 product, order, user, customer 之一",
  "data": null
}
```

---

## 6. 健康检查

**接口路径**: `GET /health`

**功能描述**: 检查搜索服务健康状态，包括 Elasticsearch 连接状态和数据新鲜度。此接口无需限流，供监控系统调用。

### 请求参数

无

### 成功响应示例

**健康状态**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "healthy",
    "elasticsearch": {
      "connected": true,
      "cluster_status": "green",
      "ik_plugin": true
    },
    "data_freshness": {
      "last_sync_time": "2026-06-03T14:30:00+08:00",
      "sync_interval": 60
    }
  }
}
```

**降级状态**（ES 集群异常）：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "degraded",
    "elasticsearch": {
      "connected": true,
      "cluster_status": "red",
      "ik_plugin": true
    },
    "data_freshness": {
      "last_sync_time": "2026-06-03T14:30:00+08:00",
      "sync_interval": 60
    }
  }
}
```

**不可用状态**（ES 无法连接）：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "unhealthy",
    "elasticsearch": {
      "connected": false,
      "cluster_status": "unknown",
      "ik_plugin": false
    },
    "data_freshness": {
      "last_sync_time": "2026-06-03T14:30:00+08:00",
      "sync_interval": 60
    }
  }
}
```

### 健康状态说明

| 状态 | 说明 |
| :--- | :--- |
| `healthy` | 服务完全正常，ES 连接正常且集群状态为 green/yellow |
| `degraded` | 服务降级，ES 集群状态为 red 或 IK 插件未安装 |
| `unhealthy` | 服务不可用，ES 无法连接 |
