# GoShopAdmin API 接口文档

本文档涵盖 GoShopAdmin 系统所有后端服务的 API 接口，按项目分类组织。

---

## 项目概览

| 项目 | 目录 | 说明 | 前端路由文件 |
| :--- | :--- | :--- | :--- |
| 后台管理系统（backend） | `backend/` | 管理员后台 API 服务 | `frontend/src/router/index.js` |
| C 端商城（shop-backend） | `shop-backend/` | C 端消费者商城 API 服务 | `shop-frontend/src/router/index.js` |
| 搜索服务（search-service） | `search-service/` | 基于 Elasticsearch 的搜索服务 | 由 backend/shop-backend 代理 |

---

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

> HTTP 状态码始终为 200，前端根据 `body.code` 判断成功/失败。

### 例外情况

以下接口不走统一响应格式：
- 验证码图片接口（返回 PNG 图片流）
- 健康检查接口（直接返回 `gin.H`）
- 后台首页接口（直接返回 `gin.H`）

### 认证方式

需要认证的接口需在请求头中携带 JWT Token：

```
Authorization: Bearer <token>
```

### 错误码

| 错误码 | 常量名 | 说明 |
| :--- | :--- | :--- |
| 0 | `CodeSuccess` | 成功 |
| 4001 | `CodeParamError` | 参数错误 |
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
| 4080 | `CodeSearchError` | 搜索服务错误 |
| 4081 | `CodeSearchTimeout` | 搜索请求超时 |
| 4082 | `CodeSearchRateLimited` | 搜索请求过于频繁 |
| 4083 | `CodeESUnavailable` | 搜索服务暂不可用 |
| 4090 | `CodeConflict` | 资源冲突 |
| 4091 | `CodeDuplicate` | 数据重复 |
| 4092 | `CodeStockInsufficient` | 库存不足 |
| 4093 | `CodeCaptchaError` | 验证码错误 |
| 4094 | `CodeUserExists` | 用户名已存在 |
| 5000 | `CodeInternalError` | 内部错误 |
| 5001 | `CodeDBError` | 数据库错误 |
| 5002 | `CodeCacheError` | 缓存错误 |
| 5003 | `CodeExternalError` | 外部服务错误 |
| 5004 | `CodeESError` | 搜索引擎错误 |

---
---

# 一、后台管理系统（backend）

> **后端目录**: `backend/`
> **前端路由文件**: `frontend/src/router/index.js`
> **权限中间件**: 用户管理需 `user:manage`，角色/权限管理需 `role:manage`，商户管理需 `merchant:manage`，商品相关管理需 `product:manage`，活动/兑换码/搜索仅需认证

---

## 1.1 健康检查

**接口路径**: `GET /health`

**功能描述**: 检查后台管理系统服务健康状态。无需认证。

### 请求参数

无

### 成功响应示例

```json
{
  "status": "ok",
  "message": "Goshopadmin backend service is running"
}
```

---

## 1.2 首页

**接口路径**: `GET /`

**功能描述**: 后台首页欢迎信息。无需认证。

### 请求参数

无

### 成功响应示例

```json
{
  "message": "Hello World!",
  "status": "success",
  "service": "Goshopadmin Backend"
}
```

---

## 1.3 认证

### 1.3.1 获取验证码

**接口路径**: `GET /api/auth/captcha`

**功能描述**: 获取图形验证码，用于登录验证。无需认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "captcha_abc123",
    "image": "data:image/png;base64,...",
    "ans": 7
  }
}
```

---

### 1.3.2 验证验证码

**接口路径**: `POST /api/auth/captcha/verify`

**功能描述**: 验证图形验证码是否正确。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `captcha_id` | `string` | 是 | 验证码ID，从获取验证码接口返回 |
| `captcha_ans` | `int` | 是 | 验证码答案 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.3.3 用户登录

**接口路径**: `POST /api/auth/login`

**功能描述**: 后台管理员登录，返回 JWT Token。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `username` | `string` | 是 | 用户名 |
| `password` | `string` | 是 | 密码 |
| `captcha_id` | `string` | 是 | 验证码ID |
| `captcha_ans` | `int` | 是 | 验证码答案 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "username": "admin",
      "role_name": "超级管理员"
    }
  }
}
```

### 错误响应示例

```json
{
  "code": 4013,
  "message": "登录失败: 用户名或密码错误",
  "data": null
}
```

---

### 1.3.4 用户登出

**接口路径**: `POST /api/auth/logout`

**功能描述**: 退出登录，使当前 Token 失效。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.3.5 刷新 Token

**接口路径**: `POST /api/auth/refresh`

**功能描述**: 刷新 JWT Token，返回新 Token。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 1.3.6 获取当前用户信息

**接口路径**: `GET /api/auth/me`

**功能描述**: 获取当前登录用户的信息和权限列表。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "username": "admin",
    "role_name": "超级管理员",
    "permissions": ["user:manage", "role:manage", "product:manage"]
  }
}
```

---

## 1.4 用户管理

> 需要 `user:manage` 权限

### 1.4.1 获取用户列表

**接口路径**: `GET /api/users`

**功能描述**: 获取后台管理系统用户列表。需要认证 + `user:manage` 权限。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 按用户名筛选 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role_id": 1,
      "status": "active",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "role": {
        "id": 1,
        "name": "超级管理员"
      }
    }
  ]
}
```

---

### 1.4.2 获取单个用户

**接口路径**: `GET /api/users/:id`

**功能描述**: 获取指定用户详细信息。需要认证 + `user:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 用户ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "role_id": 1,
    "status": "active",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "role": {
      "id": 1,
      "name": "超级管理员"
    }
  }
}
```

---

### 1.4.3 创建用户

**接口路径**: `POST /api/users`

**功能描述**: 创建后台管理系统用户。需要认证 + `user:manage` 权限。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `username` | `string` | 是 | 用户名 |
| `password` | `string` | 是 | 密码 |
| `role_id` | `int` | 是 | 角色ID |
| `status` | `string` | 否 | 状态：`active` 启用 / `inactive` 禁用 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "username": "operator",
    "role_id": 2,
    "status": "active",
    "created_at": "2026-06-03T10:00:00Z",
    "updated_at": "2026-06-03T10:00:00Z"
  }
}
```

---

### 1.4.4 更新用户

**接口路径**: `PUT /api/users/:id`

**功能描述**: 更新指定用户信息。需要认证 + `user:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 用户ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `password` | `string` | 否 | 新密码 |
| `role_id` | `int` | 否 | 角色ID |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "username": "operator",
    "role_id": 3,
    "status": "active",
    "created_at": "2026-06-03T10:00:00Z",
    "updated_at": "2026-06-03T11:00:00Z"
  }
}
```

---

### 1.4.5 删除用户

**接口路径**: `DELETE /api/users/:id`

**功能描述**: 删除指定用户。需要认证 + `user:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 用户ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.5 角色管理

> 需要 `role:manage` 权限

### 1.5.1 获取角色列表

**接口路径**: `GET /api/roles`

**功能描述**: 获取角色列表，包含关联的权限信息。需要认证 + `role:manage` 权限。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "超级管理员",
      "description": "拥有所有权限",
      "status": "active",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "permissions": [
        {"id": 1, "name": "用户管理", "code": "user:manage"}
      ]
    }
  ]
}
```

---

### 1.5.2 获取单个角色

**接口路径**: `GET /api/roles/:id`

**功能描述**: 获取指定角色详细信息。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 角色ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "超级管理员",
    "description": "拥有所有权限",
    "status": "active",
    "permissions": [
      {"id": 1, "name": "用户管理", "code": "user:manage"}
    ]
  }
}
```

---

### 1.5.3 创建角色

**接口路径**: `POST /api/roles`

**功能描述**: 创建新角色。需要认证 + `role:manage` 权限。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 角色名称 |
| `description` | `string` | 否 | 角色描述 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "name": "运营人员",
    "description": "负责日常运营",
    "status": "active"
  }
}
```

---

### 1.5.4 更新角色

**接口路径**: `PUT /api/roles/:id`

**功能描述**: 更新指定角色信息。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 角色ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 角色名称 |
| `description` | `string` | 否 | 角色描述 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "name": "运营人员",
    "description": "负责日常运营管理",
    "status": "active"
  }
}
```

---

### 1.5.5 删除角色

**接口路径**: `DELETE /api/roles/:id`

**功能描述**: 删除指定角色。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 角色ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.5.6 为角色分配权限

**接口路径**: `POST /api/roles/:id/permissions`

**功能描述**: 为指定角色分配权限列表，覆盖原有权限。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 角色ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `permission_ids` | `[]int` | 是 | 权限ID列表 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.6 权限管理

> 需要 `role:manage` 权限

### 1.6.1 获取权限列表

**接口路径**: `GET /api/permissions`

**功能描述**: 获取所有权限列表。需要认证 + `role:manage` 权限。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "用户管理",
      "code": "user:manage",
      "description": "管理系统用户",
      "category": "系统管理",
      "status": "active",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### 1.6.2 获取单个权限

**接口路径**: `GET /api/permissions/:id`

**功能描述**: 获取指定权限详细信息。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 权限ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "用户管理",
    "code": "user:manage",
    "description": "管理系统用户",
    "category": "系统管理",
    "status": "active"
  }
}
```

---

### 1.6.3 创建权限

**接口路径**: `POST /api/permissions`

**功能描述**: 创建新权限。需要认证 + `role:manage` 权限。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 权限名称 |
| `code` | `string` | 是 | 权限代码，如 `user:manage` |
| `description` | `string` | 否 | 权限描述 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 5,
    "name": "活动管理",
    "code": "activity:manage",
    "description": "管理营销活动",
    "status": "active"
  }
}
```

---

### 1.6.4 更新权限

**接口路径**: `PUT /api/permissions/:id`

**功能描述**: 更新指定权限信息。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 权限ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 权限名称 |
| `code` | `string` | 否 | 权限代码 |
| `description` | `string` | 否 | 权限描述 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 5,
    "name": "活动管理",
    "code": "activity:manage",
    "description": "管理营销活动和兑换码",
    "status": "active"
  }
}
```

---

### 1.6.5 删除权限

**接口路径**: `DELETE /api/permissions/:id`

**功能描述**: 删除指定权限。需要认证 + `role:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 权限ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.7 商户管理

> 需要 `merchant:manage` 权限

### 1.7.1 获取商户列表

**接口路径**: `GET /api/merchants`

**功能描述**: 获取商户列表。需要认证 + `merchant:manage` 权限。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "苹果官方旗舰店",
      "contact_name": "张经理",
      "contact_phone": "13800138000",
      "email": "apple@example.com",
      "address": "北京市朝阳区",
      "business_license": "LICENSE001",
      "tax_number": "TAX001",
      "audit_status": "approved",
      "status": "active",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### 1.7.2 获取单个商户

**接口路径**: `GET /api/merchants/:id`

**功能描述**: 获取指定商户详细信息，包含关联用户、审核记录等。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "苹果官方旗舰店",
    "contact_name": "张经理",
    "contact_phone": "13800138000",
    "email": "apple@example.com",
    "address": "北京市朝阳区",
    "business_license": "LICENSE001",
    "tax_number": "TAX001",
    "audit_status": "approved",
    "status": "active",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "users": [],
    "audits": [],
    "banks": []
  }
}
```

---

### 1.7.3 创建商户

**接口路径**: `POST /api/merchants`

**功能描述**: 创建新商户。需要认证 + `merchant:manage` 权限。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 商户名称 |
| `contact_name` | `string` | 是 | 联系人姓名 |
| `contact_phone` | `string` | 是 | 联系电话 |
| `email` | `string` | 是 | 邮箱 |
| `address` | `string` | 是 | 地址 |
| `business_license` | `string` | 是 | 营业执照号 |
| `tax_number` | `string` | 是 | 税号 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "华为旗舰店",
    "contact_name": "李经理",
    "status": "inactive",
    "audit_status": "pending"
  }
}
```

---

### 1.7.4 更新商户

**接口路径**: `PUT /api/merchants/:id`

**功能描述**: 更新指定商户信息。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 商户名称 |
| `contact_name` | `string` | 否 | 联系人姓名 |
| `contact_phone` | `string` | 否 | 联系电话 |
| `email` | `string` | 否 | 邮箱 |
| `address` | `string` | 否 | 地址 |
| `business_license` | `string` | 否 | 营业执照号 |
| `tax_number` | `string` | 否 | 税号 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "华为旗舰店",
    "status": "active"
  }
}
```

---

### 1.7.5 禁用商户

**接口路径**: `DELETE /api/merchants/:id`

**功能描述**: 禁用指定商户。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.7.6 审核商户

**接口路径**: `PUT /api/merchants/:id/audit`

**功能描述**: 审核商户，设置审核状态和备注。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `audit_status` | `string` | 是 | 审核状态：`pending` / `approved` / `rejected` |
| `audit_note` | `string` | 否 | 审核备注 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.7.7 获取商户用户列表

**接口路径**: `GET /api/merchants/:id/users`

**功能描述**: 获取指定商户关联的用户列表。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "merchant_id": 1,
      "user_id": 1,
      "role": "owner",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### 1.7.8 添加商户用户

**接口路径**: `POST /api/merchants/:id/users`

**功能描述**: 为指定商户添加用户。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `user_id` | `int` | 是 | 用户ID |
| `role` | `string` | 是 | 角色：`owner` / `manager` / `staff` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.7.9 移除商户用户

**接口路径**: `DELETE /api/merchants/:id/users/:user_id`

**功能描述**: 从指定商户移除用户。需要认证 + `merchant:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商户ID |
| `user_id` | `int` | 是 | 用户ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.8 商品管理

> 需要 `product:manage` 权限

### 1.8.1 获取商品列表

**接口路径**: `GET /api/products`

**功能描述**: 获取商品列表。需要认证 + `product:manage` 权限。返回数据包含 SKU 列表，`stock` 字段为所有启用状态 SKU 的库存总和。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 按商品名称筛选 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "苹果手机 iPhone 15 Pro",
      "description": "全新苹果旗舰手机",
      "detail": "<p>商品详情HTML</p>",
      "price": 7999.00,
      "stock": 200,
      "category_id": 1,
      "merchant_id": 1,
      "status": "active",
      "is_activity": 0,
      "created_at": "2026-01-15T10:30:00Z",
      "updated_at": "2026-06-01T08:00:00Z",
      "category": {
        "id": 1,
        "name": "手机"
      },
      "images": [
        {
          "id": 1,
          "product_id": 1,
          "image_url": "https://example.com/iphone15.jpg",
          "is_main": true,
          "sort": 0
        }
      ],
      "skus": [
        {
          "id": 1,
          "product_id": 1,
          "sku_code": "SKU-001",
          "price": 7999.00,
          "original_price": 8999.00,
          "stock": 100,
          "status": "active"
        }
      ]
    }
  ]
}
```

> **说明**：`stock` 字段为所有 `status: active` 的 SKU 库存总和，非数据库原始值。

---

### 1.8.2 获取商品详情

**接口路径**: `GET /api/products/:id`

**功能描述**: 获取指定商品详细信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "苹果手机 iPhone 15 Pro",
    "description": "全新苹果旗舰手机",
    "detail": "<p>商品详情HTML</p>",
    "category_id": 1,
    "status": "active",
    "created_at": "2026-01-15T10:30:00Z",
    "updated_at": "2026-06-01T08:00:00Z"
  }
}
```

---

### 1.8.3 创建商品

**接口路径**: `POST /api/products`

**功能描述**: 创建新商品。需要认证 + `product:manage` 权限。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 商品名称 |
| `description` | `string` | 否 | 商品描述 |
| `detail` | `string` | 否 | 商品详情（富文本HTML） |
| `stock` | `int` | 否 | 库存 |
| `category_id` | `int` | 是 | 分类ID |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "新商品",
    "category_id": 1,
    "status": "active"
  }
}
```

---

### 1.8.4 更新商品

**接口路径**: `PUT /api/products/:id`

**功能描述**: 更新指定商品信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 商品名称 |
| `description` | `string` | 否 | 商品描述 |
| `detail` | `string` | 否 | 商品详情 |
| `stock` | `int` | 否 | 库存 |
| `category_id` | `int` | 否 | 分类ID |
| `status` | `string` | 否 | 状态 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"id": 1}
}
```

---

### 1.8.5 删除商品

**接口路径**: `DELETE /api/products/:id`

**功能描述**: 删除指定商品。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.9 商品分类管理

> 需要 `product:manage` 权限

### 1.9.1 获取分类列表

**接口路径**: `GET /api/product-categories`

**功能描述**: 获取商品分类列表。需要认证 + `product:manage` 权限。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "手机",
      "parent_id": 0,
      "level": 1,
      "sort": 1,
      "status": "active",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### 1.9.2 获取分类详情

**接口路径**: `GET /api/product-categories/:id`

**功能描述**: 获取指定分类详细信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 分类ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "手机",
    "parent_id": 0,
    "level": 1,
    "sort": 1,
    "status": "active"
  }
}
```

---

### 1.9.3 创建分类

**接口路径**: `POST /api/product-categories`

**功能描述**: 创建商品分类。需要认证 + `product:manage` 权限。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 分类名称 |
| `parent_id` | `int` | 否 | 父分类ID，0 表示顶级分类 |
| `level` | `int` | 否 | 层级 |
| `sort` | `int` | 否 | 排序 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "手机配件",
    "parent_id": 1,
    "level": 2,
    "sort": 0,
    "status": "active"
  }
}
```

---

### 1.9.4 更新分类

**接口路径**: `PUT /api/product-categories/:id`

**功能描述**: 更新指定分类信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 分类ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 分类名称 |
| `parent_id` | `int` | 否 | 父分类ID |
| `level` | `int` | 否 | 层级 |
| `sort` | `int` | 否 | 排序 |
| `status` | `string` | 否 | 状态 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "手机配件",
    "parent_id": 1,
    "status": "active"
  }
}
```

---

### 1.9.5 删除分类

**接口路径**: `DELETE /api/product-categories/:id`

**功能描述**: 删除指定分类。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 分类ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.10 商品图片管理

> 需要 `product:manage` 权限

### 1.10.1 添加商品图片

**接口路径**: `POST /api/product-images`

**功能描述**: 为商品添加图片。需要认证 + `product:manage` 权限。使用 `multipart/form-data` 上传。

### 请求参数（Body - multipart/form-data）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `product_id` | `int` | 是 | 商品ID |
| `image_url` | `string` | 是 | 图片URL |
| `is_main` | `bool` | 否 | 是否为主图 |
| `sort` | `int` | 否 | 排序 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "product_id": 1,
    "image_url": "https://example.com/image.jpg",
    "is_main": true,
    "sort": 0
  }
}
```

---

### 1.10.2 更新商品图片

**接口路径**: `PUT /api/product-images/:id`

**功能描述**: 更新指定商品图片信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 图片ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `product_id` | `int` | 是 | 商品ID |
| `image_url` | `string` | 否 | 图片URL |
| `is_main` | `bool` | 否 | 是否为主图 |
| `sort` | `int` | 否 | 排序 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "product_id": 1,
    "image_url": "https://example.com/new-image.jpg",
    "is_main": true,
    "sort": 1
  }
}
```

---

### 1.10.3 删除商品图片

**接口路径**: `DELETE /api/product-images/:id`

**功能描述**: 删除指定商品图片。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 图片ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.11 规格管理

> 需要 `product:manage` 权限

### 1.11.1 获取商品规格列表

**接口路径**: `GET /api/products/:id/specifications`

**功能描述**: 获取指定商品的规格列表（含规格值）。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "product_id": 1,
      "name": "颜色",
      "sort": 1,
      "values": [
        {"id": 1, "spec_id": 1, "value": "暗紫色", "sort": 1, "status": "active"},
        {"id": 2, "spec_id": 1, "value": "金色", "sort": 2, "status": "active"}
      ]
    }
  ]
}
```

---

### 1.11.2 创建规格

**接口路径**: `POST /api/products/:id/specifications`

**功能描述**: 为指定商品创建规格。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 规格名称，如"颜色" |
| `sort` | `int` | 否 | 排序 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "product_id": 1,
    "name": "存储",
    "sort": 2
  }
}
```

---

### 1.11.3 更新规格

**接口路径**: `PUT /api/specifications/:id`

**功能描述**: 更新指定规格信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 规格ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 规格名称 |
| `sort` | `int` | 否 | 排序 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.11.4 删除规格

**接口路径**: `DELETE /api/specifications/:id`

**功能描述**: 删除指定规格。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 规格ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.11.5 创建规格值

**接口路径**: `POST /api/specifications/:id/values`

**功能描述**: 为指定规格创建规格值。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 规格ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `value` | `string` | 是 | 规格值，如"红色" |
| `image` | `string` | 否 | 规格值图片URL |
| `sort` | `int` | 否 | 排序 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "spec_id": 1,
    "value": "银色",
    "sort": 3,
    "status": "active"
  }
}
```

---

### 1.11.6 更新规格值

**接口路径**: `PUT /api/specification-values/:id`

**功能描述**: 更新指定规格值。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 规格值ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `value` | `string` | 是 | 规格值 |
| `image` | `string` | 否 | 图片URL |
| `sort` | `int` | 否 | 排序 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.11.7 删除规格值

**接口路径**: `DELETE /api/specification-values/:id`

**功能描述**: 删除指定规格值。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 规格值ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.12 SKU 管理

> 需要 `product:manage` 权限

### 1.12.1 创建 SKU

**接口路径**: `POST /api/products/:id/skus`

**功能描述**: 为指定商品创建 SKU。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `sku_code` | `string` | 是 | SKU 编码 |
| `price` | `float` | 是 | 价格 |
| `original_price` | `float` | 否 | 原价 |
| `stock` | `int` | 否 | 库存 |
| `status` | `string` | 否 | 状态：`active` / `inactive` |
| `spec_combinations` | `[]object` | 否 | 规格组合列表 |

`spec_combinations` 中每项：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `spec_id` | `int` | 否 | 规格ID |
| `spec_value_id` | `int` | 否 | 规格值ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "product_id": 1,
    "sku_code": "SKU-001",
    "price": 7999.00,
    "original_price": 8999.00,
    "stock": 100,
    "status": "active"
  }
}
```

---

### 1.12.2 批量创建 SKU

**接口路径**: `POST /api/products/:id/skus/batch`

**功能描述**: 批量为指定商品创建 SKU。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `skus` | `[]object` | 是 | SKU 列表 |

`skus` 中每项：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `sku_code` | `string` | 否 | SKU 编码 |
| `price` | `float` | 否 | 价格 |
| `original_price` | `float` | 否 | 原价 |
| `stock` | `int` | 否 | 库存 |
| `status` | `string` | 否 | 状态 |
| `is_activity` | `bool` | 否 | 是否为活动SKU |
| `activity_id` | `int` | 否 | 活动ID |
| `spec_combinations` | `[]object` | 否 | 规格组合列表 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.12.3 获取商品 SKU 列表

**接口路径**: `GET /api/products/:id/skus`

**功能描述**: 获取指定商品的 SKU 列表。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "product_id": 1,
      "sku_code": "SKU-001",
      "price": 7999.00,
      "original_price": 8999.00,
      "stock": 100,
      "status": "active"
    }
  ]
}
```

---

### 1.12.4 根据规格生成 SKU

**接口路径**: `POST /api/products/:id/skus/generate`

**功能描述**: 根据商品规格自动生成 SKU 组合。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `base_price` | `float` | 否 | 基础价格 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "sku_code": "AUTO-001",
      "price": 7999.00,
      "spec_combinations": [
        {"spec_id": 1, "spec_value_id": 1},
        {"spec_id": 2, "spec_value_id": 3}
      ]
    }
  ]
}
```

---

### 1.12.5 更新 SKU

**接口路径**: `PUT /api/skus/:id`

**功能描述**: 更新指定 SKU 信息。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | SKU ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `sku_code` | `string` | 否 | SKU 编码 |
| `price` | `float` | 否 | 价格 |
| `original_price` | `float` | 否 | 原价 |
| `stock` | `int` | 否 | 库存 |
| `status` | `string` | 否 | 状态 |
| `spec_combinations` | `[]object` | 否 | 规格组合列表 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.12.6 删除 SKU

**接口路径**: `DELETE /api/skus/:id`

**功能描述**: 删除指定 SKU。需要认证 + `product:manage` 权限。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | SKU ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 1.13 活动管理

> 需要认证（无特殊权限要求）

### 1.13.1 获取活动列表

**接口路径**: `GET /api/activities`

**功能描述**: 获取活动列表。需要认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `page` | `int` | 否 | 页码，默认 1 |
| `size` | `int` | 否 | 每页数量，默认 10 |
| `type` | `string` | 否 | 活动类型筛选 |
| `status` | `string` | 否 | 活动状态筛选 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "618大促",
        "type": "seckill",
        "start_time": "2026-06-18 00:00:00",
        "end_time": "2026-06-18 23:59:59",
        "status": "active",
        "product_count": 3,
        "created_at": "2026-06-01 10:00:00"
      }
    ],
    "total": 10
  }
}
```

---

### 1.13.2 获取活动详情

**接口路径**: `GET /api/activities/:id`

**功能描述**: 获取指定活动详细信息，包含活动商品和兑换码规则。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "618大促",
    "type": "seckill",
    "start_time": "2026-06-18 00:00:00",
    "end_time": "2026-06-18 23:59:59",
    "status": "active",
    "products": [
      {
        "product_id": 1,
        "product_name": "苹果手机 iPhone 15 Pro",
        "sku_id": 1,
        "sku_code": "SKU-001",
        "price": 7999.00,
        "stock": 100
      }
    ],
    "redeem_code_rules": {
      "type": "alphanumeric",
      "length": 8,
      "exclude_chars": "0O1lI"
    }
  }
}
```

---

### 1.13.3 创建活动

**接口路径**: `POST /api/activities`

**功能描述**: 创建新活动。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 活动名称 |
| `type` | `string` | 是 | 活动类型 |
| `start_time` | `string` | 是 | 开始时间 |
| `end_time` | `string` | 是 | 结束时间 |
| `status` | `string` | 否 | 状态 |
| `products` | `[]object` | 否 | 活动商品列表 |
| `redeem_setting` | `object` | 否 | 兑换码设置 |

`products` 中每项：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `product_id` | `int` | 是 | 商品ID |
| `sku_id` | `int` | 是 | SKU ID |
| `product_type` | `string` | 否 | 商品类型 |

`redeem_setting`：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `code_type` | `string` | 否 | 兑换码类型 |
| `code_length` | `int` | 否 | 兑换码长度 |
| `exclude_chars` | `string` | 否 | 排除字符 |
| `total_quantity` | `int` | 否 | 总数量 |
| `limit_per_user` | `int` | 否 | 每用户限制 |
| `need_verify` | `int` | 否 | 是否需要核销 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "活动创建成功"}
}
```

---

### 1.13.4 更新活动

**接口路径**: `PUT /api/activities/:id`

**功能描述**: 更新指定活动信息。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 活动名称 |
| `start_time` | `string` | 是 | 开始时间 |
| `end_time` | `string` | 是 | 结束时间 |
| `status` | `string` | 是 | 状态 |
| `products` | `[]object` | 否 | 活动商品列表 |
| `redeem_setting` | `object` | 否 | 兑换码设置 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "活动更新成功"}
}
```

---

### 1.13.5 删除活动

**接口路径**: `DELETE /api/activities/:id`

**功能描述**: 删除指定活动。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "活动删除成功"}
}
```

---

### 1.13.6 更新活动状态

**接口路径**: `PUT /api/activities/:id/status`

**功能描述**: 更新指定活动状态。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `status` | `string` | 是 | 活动状态 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "活动状态更新成功"}
}
```

---

## 1.14 兑换码管理

> 需要认证（无特殊权限要求）

### 1.14.1 生成兑换码

**接口路径**: `POST /api/activities/:id/redeem-codes/generate`

**功能描述**: 为指定活动批量生成兑换码。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `activity_id` | `int` | 是 | 活动ID |
| `quantity` | `int` | 是 | 生成数量（1-10000） |
| `code_type` | `string` | 否 | 兑换码类型 |
| `code_length` | `int` | 否 | 兑换码长度 |
| `exclude_chars` | `string` | 否 | 排除字符 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "generated_count": 10,
    "codes": [
      {"id": 1, "code": "ABCD1234", "status": "unused"}
    ]
  }
}
```

---

### 1.14.2 获取兑换码列表

**接口路径**: `GET /api/activities/:id/redeem-codes`

**功能描述**: 获取指定活动的兑换码列表。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `status` | `string` | 否 | 状态筛选 |
| `page` | `int` | 否 | 页码 |
| `size` | `int` | 否 | 每页数量 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "code": "ABCD1234",
        "status": "unused",
        "valid_start_time": "2026-06-01T00:00:00Z",
        "valid_end_time": "2026-06-30T23:59:59Z",
        "created_at": "2026-06-01T10:00:00Z"
      }
    ],
    "total": 50,
    "stats": {
      "total": 50,
      "used": 10,
      "unused": 35,
      "expired": 5
    }
  }
}
```

---

### 1.14.3 获取兑换码统计

**接口路径**: `GET /api/activities/:id/redeem-codes/stats`

**功能描述**: 获取指定活动的兑换码统计数据。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "used": 10,
    "unused": 35,
    "expired": 5
  }
}
```

---

### 1.14.4 导出兑换码

**接口路径**: `GET /api/activities/:id/redeem-codes/export`

**功能描述**: 导出指定活动的兑换码。需要认证。（功能开发中）

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "导出功能开发中"}
}
```

---

### 1.14.5 导入兑换码

**接口路径**: `POST /api/activities/:id/redeem-codes/import`

**功能描述**: 导入兑换码。需要认证。（功能开发中）

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "imported_count": 0,
    "failed_count": 0,
    "failed_codes": []
  }
}
```

---

### 1.14.6 核销兑换码

**接口路径**: `POST /api/redeem-codes/verify`

**功能描述**: 核销兑换码。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `code` | `string` | 是 | 兑换码 |
| `remark` | `string` | 否 | 核销备注 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "code": "ABCD1234",
    "status": "used",
    "verified_at": "2026-06-03T14:30:00Z"
  }
}
```

---

### 1.14.7 获取核销记录

**接口路径**: `GET /api/redeem-codes/logs`

**功能描述**: 获取兑换码核销记录。需要认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `activity_id` | `int` | 否 | 活动ID筛选 |
| `page` | `int` | 否 | 页码 |
| `size` | `int` | 否 | 每页数量 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "code": "ABCD1234",
        "customer_id": 1,
        "status": "verified",
        "verify_by": 1,
        "verify_at": "2026-06-03T14:30:00Z",
        "remark": "线下核销",
        "created_at": "2026-06-03T14:30:00Z"
      }
    ],
    "total": 10
  }
}
```

---

### 1.14.8 更新兑换码状态

**接口路径**: `PUT /api/redeem-codes/:id/status`

**功能描述**: 更新指定兑换码状态。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 兑换码ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `status` | `string` | 是 | 状态 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "兑换码状态更新成功"}
}
```

---

## 1.15 搜索（代理）

> 需要认证。搜索请求代理到 search-service。

### 1.15.1 搜索商品

**接口路径**: `GET /api/search/products`

**功能描述**: 搜索商品，代理到 search-service。需要认证。

### 请求参数

同 [3.1 商品搜索](#31-商品搜索)

### 响应

同 [3.1 商品搜索](#31-商品搜索)

---

### 1.15.2 搜索订单

**接口路径**: `GET /api/search/orders`

**功能描述**: 搜索订单，代理到 search-service。需要认证。

### 请求参数

同 [3.2 订单搜索](#32-订单搜索)

### 响应

同 [3.2 订单搜索](#32-订单搜索)

---

### 1.15.3 搜索用户

**接口路径**: `GET /api/search/users`

**功能描述**: 搜索用户，代理到 search-service。需要认证。

### 请求参数

同 [3.3 用户搜索](#33-用户搜索)

### 响应

同 [3.3 用户搜索](#33-用户搜索)

---

### 1.15.4 搜索客户

**接口路径**: `GET /api/search/customers`

**功能描述**: 搜索客户，代理到 search-service。需要认证。

### 请求参数

同 [3.4 客户搜索](#34-客户搜索)

### 响应

同 [3.4 客户搜索](#34-客户搜索)

---

### 1.15.5 搜索建议

**接口路径**: `GET /api/search/suggest`

**功能描述**: 搜索建议，代理到 search-service。需要认证。

### 请求参数

同 [3.5 搜索建议](#35-搜索建议)

### 响应

同 [3.5 搜索建议](#35-搜索建议)

---

## 1.16 监控

> 需要认证

### 1.16.1 监控指标

**接口路径**: `GET /api/monitor/*`

**功能描述**: 监控相关接口，由 `utils.Monitor.RegisterHTTPHandlers` 注册。需要认证。

### 请求参数

由 Monitor 工具内部定义

### 响应

由 Monitor 工具内部定义

---
---

# 二、C 端商城（shop-backend）

> **后端目录**: `shop-backend/`
> **前端路由文件**: `shop-frontend/src/router/index.js`
> **命名约定**: C 端用户统一使用 `customer` 命名，对应 `customers` 数据库表

---

## 2.1 健康检查

### 2.1.1 服务健康检查

**接口路径**: `GET /health`

**功能描述**: 检查 C 端商城服务健康状态。无需认证。

### 请求参数

无

### 成功响应示例

```json
{
  "status": "ok"
}
```

---

### 2.1.2 MQ 健康检查

**接口路径**: `GET /api/health/mq`

**功能描述**: 检查消息队列连接状态。无需认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "MQ连接正常",
  "data": null
}
```

---

## 2.2 验证码

### 2.2.1 生成验证码

**接口路径**: `GET /api/captcha`

**功能描述**: 生成图形验证码，返回 PNG 图片流。无需认证。

### 请求参数

无

### 成功响应

- **Content-Type**: `image/png`
- **响应头**: `X-Captcha-ID` 包含验证码ID
- **响应体**: PNG 图片流

> 此接口不走统一响应格式，直接返回图片流。

---

### 2.2.2 验证验证码

**接口路径**: `POST /api/captcha/verify`

**功能描述**: 验证图形验证码是否正确。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `captcha_id` | `string` | 是 | 验证码ID |
| `value` | `string` | 是 | 验证码值 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"valid": true}
}
```

---

## 2.3 认证

### 2.3.1 注册

**接口路径**: `POST /api/auth/register`

**功能描述**: C 端客户注册。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `username` | `string` | 是 | 用户名 |
| `password` | `string` | 是 | 密码 |
| `captcha_id` | `string` | 是 | 验证码ID |
| `captcha` | `string` | 是 | 验证码值 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Register success",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 2.3.2 登录

**接口路径**: `POST /api/auth/login`

**功能描述**: C 端客户登录。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `username` | `string` | 是 | 用户名 |
| `password` | `string` | 是 | 密码 |
| `captcha_id` | `string` | 是 | 验证码ID |
| `captcha` | `string` | 是 | 验证码值 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 2.3.3 登出

**接口路径**: `POST /api/auth/logout`

**功能描述**: C 端客户登出。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "Logout success"}
}
```

---

## 2.4 客户信息

> 需要认证

### 2.4.1 获取个人信息

**接口路径**: `GET /api/user/profile`

**功能描述**: 获取当前登录客户的个人信息。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "nickname": "小明",
    "avatar": "http://localhost:8000/uploads/avatars/1/avatar.jpg"
  }
}
```

---

### 2.4.2 上传客户头像

**接口路径**: `POST /api/user/avatar`

**功能描述**: 上传当前登录客户的头像。需要认证。使用 `multipart/form-data` 上传。

### 请求参数（Body - multipart/form-data）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `avatar` | `file` | 是 | 头像图片文件 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "avatar": "http://localhost:8000/uploads/avatars/1/avatar.jpg"
  }
}
```

---

### 2.4.3 更新个人信息

**接口路径**: `PUT /api/user/profile`

**功能描述**: 更新当前登录客户的个人信息。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `username` | `string` | 否 | 用户名 |
| `email` | `string` | 否 | 邮箱 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Profile updated",
    "username": "zhangsan",
    "email": "new@example.com",
    "phone": "13800138000",
    "nickname": "小明",
    "avatar": "http://localhost:8000/uploads/avatars/1/avatar.jpg"
  }
}
```

---

### 2.4.4 获取订单列表

**接口路径**: `GET /api/user/orders`

**功能描述**: 获取当前登录客户的订单列表。需要认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `page` | `int` | 否 | 页码，默认 1 |
| `limit` | `int` | 否 | 每页数量，默认 10 |
| `status` | `string` | 否 | 订单状态筛选 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "orders": [
      {
        "order_id": 1,
        "order_no": "ORD202606030001",
        "amount": 7999.00,
        "status": "paid",
        "created_at": "2026-06-03T14:30:00Z",
        "items": [
          {
            "product_id": 1,
            "product_name": "苹果手机 iPhone 15 Pro",
            "product_image": "https://example.com/iphone15.jpg",
            "sku_code": "SKU-001",
            "sku_attributes": "{\"颜色\":\"暗紫色\",\"存储\":\"256GB\"}",
            "price": 7999.00,
            "quantity": 1
          }
        ]
      }
    ],
    "total": 2
  }
}
```

---

## 2.5 地址管理

> 需要认证

### 2.5.1 获取地址列表

**接口路径**: `GET /api/customer/addresses`

**功能描述**: 获取当前登录客户的收货地址列表。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "addresses": [
      {
        "id": 1,
        "name": "张三",
        "phone": "13800138000",
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail_address": "科技园路1号",
        "is_default": true
      }
    ]
  }
}
```

---

### 2.5.2 创建地址

**接口路径**: `POST /api/customer/addresses`

**功能描述**: 创建收货地址。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 是 | 收件人姓名 |
| `phone` | `string` | 是 | 手机号 |
| `province` | `string` | 是 | 省份 |
| `city` | `string` | 是 | 城市 |
| `district` | `string` | 是 | 区/县 |
| `detail_address` | `string` | 是 | 详细地址 |
| `is_default` | `bool` | 否 | 是否设为默认地址 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "地址创建成功",
    "address": {
      "id": 2,
      "name": "张三",
      "phone": "13800138000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "科技园路2号",
      "is_default": false
    }
  }
}
```

---

### 2.5.3 获取地址详情

**接口路径**: `GET /api/customer/addresses/:id`

**功能描述**: 获取指定收货地址详情。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 地址ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "张三",
    "phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail_address": "科技园路1号",
    "is_default": true
  }
}
```

---

### 2.5.4 更新地址

**接口路径**: `PUT /api/customer/addresses/:id`

**功能描述**: 更新指定收货地址。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 地址ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 否 | 收件人姓名 |
| `phone` | `string` | 否 | 手机号 |
| `province` | `string` | 否 | 省份 |
| `city` | `string` | 否 | 城市 |
| `district` | `string` | 否 | 区/县 |
| `detail_address` | `string` | 否 | 详细地址 |
| `is_default` | `bool` | 否 | 是否设为默认地址 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "地址更新成功",
    "address": {
      "id": 1,
      "name": "张三",
      "phone": "13900139000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "科技园路1号",
      "is_default": true
    }
  }
}
```

---

### 2.5.5 删除地址

**接口路径**: `DELETE /api/customer/addresses/:id`

**功能描述**: 删除指定收货地址。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 地址ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "地址删除成功"}
}
```

---

### 2.5.6 设置默认地址

**接口路径**: `PUT /api/customer/addresses/:id/default`

**功能描述**: 将指定地址设为默认地址。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 地址ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "默认地址设置成功"}
}
```

---

### 2.5.7 获取默认地址

**接口路径**: `GET /api/customer/addresses/default`

**功能描述**: 获取当前客户的默认收货地址。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "address": {
      "id": 1,
      "name": "张三",
      "phone": "13800138000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "科技园路1号",
      "is_default": true
    }
  }
}
```

---

## 2.6 商品浏览

> 无需认证

### 2.6.1 获取商品列表

**接口路径**: `GET /api/products`

**功能描述**: 获取 C 端商品列表，支持分页和关键词搜索。无需认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `page` | `int` | 否 | 页码，默认 1 |
| `limit` | `int` | 否 | 每页数量，默认 10 |
| `keyword` | `string` | 否 | 搜索关键词 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "products": [
      {
        "id": 1,
        "name": "苹果手机 iPhone 15 Pro",
        "description": "全新苹果旗舰手机",
        "price": 7999.00,
        "sku": "SKU-001",
        "stock": 100,
        "image": "https://example.com/iphone15.jpg",
        "images": ["https://example.com/iphone15-1.jpg"],
        "default_sku_price": 7999.00,
        "sales": 5680
      }
    ],
    "total": 100
  }
}
```

---

### 2.6.2 获取商品详情

**接口路径**: `GET /api/products/:id`

**功能描述**: 获取商品详情，包含规格、SKU 和价格区间。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "苹果手机 iPhone 15 Pro",
    "description": "全新苹果旗舰手机",
    "detail": "<p>商品详情HTML</p>",
    "price": 7999.00,
    "image": "https://example.com/iphone15.jpg",
    "images": ["https://example.com/iphone15-1.jpg"],
    "specifications": [
      {
        "id": 1,
        "name": "颜色",
        "values": [
          {"id": 1, "value": "暗紫色", "image": ""},
          {"id": 2, "value": "金色", "image": ""}
        ]
      }
    ],
    "sku_list": [
      {
        "id": 1,
        "sku_code": "SKU-001",
        "price": 7999.00,
        "original_price": 8999.00,
        "stock": 100,
        "status": "active",
        "spec_combination": {"1": 1, "2": 3}
      }
    ],
    "price_range": {"min": 7999.00, "max": 12999.00},
    "sales": 5680,
    "reviews_count": 0
  }
}
```

---

### 2.6.3 获取商品 SKU 列表

**接口路径**: `GET /api/products/:id/skus`

**功能描述**: 获取商品 SKU 列表（含规格组合）。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "skus": [
      {
        "id": 1,
        "sku_code": "SKU-001",
        "price": 7999.00,
        "original_price": 8999.00,
        "stock": 100,
        "status": "active",
        "spec_combination": {"1": 1, "2": 3}
      }
    ]
  }
}
```

---

### 2.6.4 根据规格组合获取 SKU

**接口路径**: `GET /api/products/:id/sku`

**功能描述**: 根据规格组合查询对应的 SKU。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 商品ID |

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `specs` | `string` | 是 | 规格组合，格式：`1:1,2:4`（spec_id:spec_value_id） |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "sku_code": "SKU-001",
    "price": 7999.00,
    "original_price": 8999.00,
    "stock": 100,
    "status": "active",
    "spec_combination": {"1": 1, "2": 3}
  }
}
```

---

## 2.7 购物车

> 需要认证

### 2.7.1 获取购物车

**接口路径**: `GET /api/cart`

**功能描述**: 获取当前客户的购物车内容。需要认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "product_name": "苹果手机 iPhone 15 Pro",
        "main_image": "https://example.com/iphone15.jpg",
        "sku_id": 1,
        "sku_code": "SKU-001",
        "quantity": 2,
        "price": 7999.00
      }
    ]
  }
}
```

---

### 2.7.2 添加商品到购物车

**接口路径**: `POST /api/cart/items`

**功能描述**: 添加商品到购物车。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `product_id` | `int` | 是 | 商品ID |
| `sku_id` | `int` | 否 | SKU ID |
| `quantity` | `int` | 是 | 数量 |
| `price` | `float` | 是 | 价格 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Item added to cart",
    "item": {
      "product_id": 1,
      "sku_id": 1,
      "quantity": 2,
      "price": 7999.00
    }
  }
}
```

---

### 2.7.3 更新购物车项

**接口路径**: `PUT /api/cart/items/:id`

**功能描述**: 更新购物车中指定商品的数量。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 购物车项ID |

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `quantity` | `int` | 是 | 新数量 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Cart item updated",
    "item_id": "1",
    "quantity": 3
  }
}
```

---

### 2.7.4 移除购物车项

**接口路径**: `DELETE /api/cart/items/:id`

**功能描述**: 从购物车移除指定商品。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 购物车项ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Cart item removed",
    "item_id": "1"
  }
}
```

---

### 2.7.5 同步购物车

**接口路径**: `POST /api/cart/sync`

**功能描述**: 同步本地购物车到服务端。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `items` | `[]object` | 否 | 购物车项列表 |

`items` 中每项：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 否 | 购物车项ID |
| `product_id` | `int` | 否 | 商品ID |
| `product_name` | `string` | 否 | 商品名称 |
| `main_image` | `string` | 否 | 主图URL |
| `sku_id` | `int` | 否 | SKU ID |
| `sku_code` | `string` | 否 | SKU 编码 |
| `quantity` | `int` | 否 | 数量 |
| `price` | `float` | 否 | 价格 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Cart synced",
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "product_name": "苹果手机 iPhone 15 Pro",
        "main_image": "https://example.com/iphone15.jpg",
        "sku_id": 1,
        "sku_code": "SKU-001",
        "quantity": 2,
        "price": 7999.00
      }
    ]
  }
}
```

---

## 2.8 订单

> 需要认证

### 2.8.1 创建订单

**接口路径**: `POST /api/orders`

**功能描述**: 创建订单，包含库存检查和扣减。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `address_id` | `int` | 是 | 收货地址ID |
| `items` | `[]object` | 是 | 订单商品列表 |
| `remark` | `string` | 否 | 订单备注 |

`items` 中每项：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `product_id` | `int` | 是 | 商品ID |
| `sku_id` | `int` | 否 | SKU ID |
| `quantity` | `int` | 是 | 数量（最小1） |

**说明**：商品价格由后端查询数据库获取实时价格，订单总金额由后端计算。库存不足时返回错误。

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "order_no": "ORD202606030001",
    "amount": 15998.00,
    "payment_url": "/api/payment/fake-pay?orderNo=ORD202606030001",
    "status": "pending",
    "created_at": "2026-06-03T14:30:00Z"
  }
}
```

---

### 2.8.2 获取订单详情

**接口路径**: `GET /api/orders/:orderNo`

**功能描述**: 获取订单详情，包含收货地址和商品信息。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `orderNo` | `string` | 是 | 订单号 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": 1,
    "order_no": "ORD202606030001",
    "total_amount": 7999.00,
    "status": "paid",
    "created_at": "2026-06-03T14:30:00Z",
    "address": {
      "name": "张三",
      "phone": "13800138000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail_address": "科技园路1号"
    },
    "items": [
      {
        "product_id": 1,
        "product_name": "苹果手机 iPhone 15 Pro",
        "product_image": "https://example.com/iphone15.jpg",
        "sku_code": "SKU-001",
        "sku_attributes": "{\"颜色\":\"暗紫色\",\"存储\":\"256GB\"}",
        "price": 7999.00,
        "quantity": 1
      }
    ]
  }
}
```

---

### 2.8.3 取消订单

**接口路径**: `PUT /api/orders/:orderNo/cancel`

**功能描述**: 取消指定订单。需要认证。仅 `pending`（待支付）状态的订单可以取消。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `orderNo` | `string` | 是 | 订单号 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "订单已取消"}
}
```

---

### 2.8.4 确认收货

**接口路径**: `PUT /api/orders/:orderNo/confirm`

**功能描述**: 确认收货。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `orderNo` | `string` | 是 | 订单号 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "确认收货成功"}
}
```

---

## 2.9 支付

### 2.9.1 伪支付

**接口路径**: `GET /api/payment/fake-pay`

**功能描述**: 伪支付页面，用于开发测试。无需认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `orderNo` | `string` | 是 | 订单号 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "ORD202606030001",
    "amount": "7999.00",
    "message": "支付成功"
  }
}
```

---

### 2.9.2 支付回调

**接口路径**: `POST /api/payment/callback`

**功能描述**: 支付结果回调通知。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `order_no` | `string` | 是 | 订单号 |
| `transaction_id` | `string` | 是 | 交易流水号 |
| `status` | `string` | 是 | 支付状态 |
| `amount` | `float` | 是 | 支付金额 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "Payment callback received",
    "order_no": "ORD202606030001",
    "transaction_id": "TRX1234567890",
    "status": "success"
  }
}
```

---

## 2.10 活动

> 部分无需认证

### 2.10.1 获取活动列表

**接口路径**: `GET /api/activities`

**功能描述**: 获取进行中的活动列表。无需认证。

### 请求参数

无

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "618大促",
      "status": "active",
      "start_time": "2026-06-18T00:00:00Z",
      "end_time": "2026-06-18T23:59:59Z",
      "products": []
    }
  ]
}
```

---

### 2.10.2 获取活动详情

**接口路径**: `GET /api/activities/:id`

**功能描述**: 获取指定活动详情，包含活动商品。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "618大促",
    "status": "active",
    "start_time": "2026-06-18T00:00:00Z",
    "end_time": "2026-06-18T23:59:59Z",
    "products": [
      {
        "id": 1,
        "activity_id": 1,
        "product_id": 1,
        "price": 6999.00,
        "stock": 50,
        "limit": 1
      }
    ]
  }
}
```

---

### 2.10.3 获取活动商品列表

**接口路径**: `GET /api/activities/:id/products`

**功能描述**: 获取指定活动的商品列表。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "activity_id": 1,
      "product_id": 1,
      "price": 6999.00,
      "stock": 50,
      "limit": 1
    }
  ]
}
```

---

### 2.10.4 获取活动 SKU 列表

**接口路径**: `GET /api/activities/:id/skus`

**功能描述**: 获取指定活动的 SKU 列表。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "sku_id": 1,
      "sku_code": "SKU-001",
      "price": 6999.00,
      "stock": 50,
      "product_id": 1,
      "product_name": "苹果手机 iPhone 15 Pro",
      "description": "全新苹果旗舰手机",
      "is_activity": 1,
      "main_image": "https://example.com/iphone15.jpg"
    }
  ]
}
```

---

### 2.10.5 获取活动 SKU 详情

**接口路径**: `GET /api/activities/:id/skus/:sku_id`

**功能描述**: 获取指定活动的 SKU 详情。无需认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动ID |
| `sku_id` | `int` | 是 | SKU ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "activity_name": "618大促",
    "activity_id": 1,
    "sku_id": 1,
    "sku_code": "SKU-001",
    "price": 6999.00,
    "stock": 50,
    "sku_status": "active",
    "is_activity": 1,
    "product_id": 1,
    "product_name": "苹果手机 iPhone 15 Pro",
    "product_description": "全新苹果旗舰手机",
    "main_image": "https://example.com/iphone15.jpg"
  }
}
```

---

## 2.11 兑换码

### 2.11.1 验证兑换码

**接口路径**: `POST /api/redeem-codes/verify`

**功能描述**: 验证兑换码是否有效。无需认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `code` | `string` | 是 | 兑换码 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "code": "ABCD1234",
    "activity_id": 1,
    "value": 100.00,
    "status": "unused",
    "expire_time": "2026-06-30T23:59:59Z"
  }
}
```

---

### 2.11.2 兑换码兑换

**接口路径**: `POST /api/redeem-codes/redeem`

**功能描述**: 使用兑换码进行兑换。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `code` | `string` | 是 | 兑换码 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "code": "ABCD1234",
    "activity_id": 1,
    "value": 100.00,
    "status": "used",
    "customer_id": 1,
    "used_at": "2026-06-03T15:00:00Z"
  }
}
```

---

### 2.11.3 获取兑换记录

**接口路径**: `GET /api/redeem-codes/logs`

**功能描述**: 获取当前客户的兑换码使用记录。需要认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `page` | `int` | 否 | 页码，默认 1 |
| `page_size` | `int` | 否 | 每页数量，默认 10 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "activity_id": 1,
        "redeem_code_id": 1,
        "customer_id": 1,
        "code": "ABCD1234",
        "value": 100.00,
        "redeem_time": "2026-06-03T15:00:00Z",
        "status": "redeemed",
        "created_at": "2026-06-03T15:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 2.12 活动订单

> 需要认证

### 2.12.1 创建活动订单

**接口路径**: `POST /api/activity-orders`

**功能描述**: 创建活动商品订单。需要认证。

### 请求参数（Body - JSON）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `activity_id` | `int` | 是 | 活动ID |
| `address_id` | `int` | 是 | 收货地址ID |
| `items` | `[]object` | 是 | 订单商品列表 |

`items` 中每项：

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `product_id` | `int` | 是 | 商品ID |
| `sku_id` | `int` | 是 | SKU ID |
| `quantity` | `int` | 是 | 数量（最小1） |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "订单已提交，正在处理中"}
}
```

---

### 2.12.2 获取活动订单列表

**接口路径**: `GET /api/activity-orders`

**功能描述**: 获取当前客户的活动订单列表。需要认证。

### 请求参数（Query String）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `page` | `int` | 否 | 页码，默认 1 |
| `page_size` | `int` | 否 | 每页数量，默认 10 |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "order_no": "ACT202606030001",
        "customer_id": 1,
        "activity_id": 1,
        "activity_name": "618大促",
        "total_amount": 6999.00,
        "status": "pending",
        "created_at": "2026-06-03T15:00:00Z",
        "items": [
          {
            "id": 1,
            "order_id": 1,
            "product_id": 1,
            "sku_id": 1,
            "product_name": "苹果手机 iPhone 15 Pro",
            "sku_attributes": "{\"颜色\":\"暗紫色\"}",
            "product_image": "https://example.com/iphone15.jpg",
            "price": 6999.00,
            "quantity": 1,
            "total_amount": 6999.00
          }
        ]
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 10
  }
}
```

---

### 2.12.3 获取活动订单详情

**接口路径**: `GET /api/activity-orders/:id`

**功能描述**: 获取指定活动订单详情。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动订单ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "order_no": "ACT202606030001",
    "customer_id": 1,
    "activity_id": 1,
    "activity_name": "618大促",
    "total_amount": 6999.00,
    "status": "pending",
    "created_at": "2026-06-03T15:00:00Z",
    "items": [
      {
        "id": 1,
        "order_id": 1,
        "product_id": 1,
        "sku_id": 1,
        "product_name": "苹果手机 iPhone 15 Pro",
        "sku_attributes": "{\"颜色\":\"暗紫色\"}",
        "product_image": "https://example.com/iphone15.jpg",
        "price": 6999.00,
        "quantity": 1,
        "total_amount": 6999.00
      }
    ]
  }
}
```

---

### 2.12.4 取消活动订单

**接口路径**: `PUT /api/activity-orders/:id/cancel`

**功能描述**: 取消指定活动订单。需要认证。

### 请求参数（Path）

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | `int` | 是 | 活动订单ID |

### 成功响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {"message": "订单取消成功"}
}
```

---

## 2.13 监控

> 需要认证

### 2.13.1 获取当前监控指标

**接口路径**: `GET /api/monitor/stats`

**功能描述**: 获取当前系统监控指标。需要认证。

### 请求参数

无

### 成功响应示例

由 `utils.Monitor` 内部定义，返回当前 CPU、内存、Goroutine 等运行时指标。

---

### 2.13.2 获取历史监控指标

**接口路径**: `GET /api/monitor/stats/history`

**功能描述**: 获取历史系统监控指标。需要认证。

### 请求参数

无

### 成功响应示例

由 `utils.Monitor` 内部定义，返回历史监控数据。

---

## 2.14 搜索（代理）

> 需要认证。搜索请求代理到 search-service。

### 2.14.1 搜索商品

**接口路径**: `GET /api/search/products`

**功能描述**: 搜索商品，代理到 search-service。需要认证。

### 请求参数

同 [3.1 商品搜索](#31-商品搜索)

### 响应

同 [3.1 商品搜索](#31-商品搜索)

---

### 2.14.2 搜索订单

**接口路径**: `GET /api/search/orders`

**功能描述**: 搜索订单，代理到 search-service。需要认证。

### 请求参数

同 [3.2 订单搜索](#32-订单搜索)

### 响应

同 [3.2 订单搜索](#32-订单搜索)

---

### 2.14.3 搜索建议

**接口路径**: `GET /api/search/suggest`

**功能描述**: 搜索建议，代理到 search-service。需要认证。

### 请求参数

同 [3.5 搜索建议](#35-搜索建议)

### 响应

同 [3.5 搜索建议](#35-搜索建议)

---

## 2.15 Prometheus 指标

### 2.15.1 Metrics 端点

**接口路径**: `GET /metrics`

**功能描述**: Prometheus 监控指标端点，仅允许内部网络（Docker 网段、局域网、本机）访问。

### 请求参数

无

### 响应

Prometheus 标准指标格式

---

## 2.16 pprof 性能分析

> 需要认证

### 2.16.1 pprof 端点

**接口路径**: `GET /debug/pprof/*`

**功能描述**: Go runtime 性能分析端点，需要认证后才能访问。

### 请求参数

由 pprof 标准库定义

### 响应

由 pprof 标准库定义

---
---

# 三、搜索服务（search-service）

> **后端目录**: `search-service/`
> **前端路由文件**: 无（由 backend / shop-backend 代理访问）
> **中间件**: CORS 跨域 + RequestLogger 请求日志 + RateLimit 限流（50 QPS）

---

## 3.1 商品搜索

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

## 3.2 订单搜索

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

## 3.3 用户搜索

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

## 3.4 客户搜索

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

## 3.5 搜索建议

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

## 3.6 健康检查

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

---

# 四、前端路由对照表

## 4.1 后台管理系统前端路由

> **路由文件**: `frontend/src/router/index.js`

| 前端路由路径 | 组件 | 对应后端 API | 权限 |
| :--- | :--- | :--- | :--- |
| `/login` | Login | `/api/auth/login` | 无 |
| `/home/dashboard` | Dashboard | - | 无 |
| `/home/users` | Users | `/api/users` | `user:manage` |
| `/home/roles` | Roles | `/api/roles` | `role:manage` |
| `/home/permissions` | Permissions | `/api/permissions` | `role:manage` |
| `/home/merchants` | Merchants | `/api/merchants` | `merchant:manage` |
| `/home/products` | Products | `/api/products` | `product:manage` |
| `/home/product-categories` | ProductCategories | `/api/product-categories` | `product:manage` |
| `/home/products/:id/specifications` | ProductSpecifications | `/api/products/:id/specifications` | `product:manage` |
| `/home/products/:id/skus` | ProductSkus | `/api/products/:id/skus` | `product:manage` |
| `/home/activities` | Activities | `/api/activities` | `activity:manage` |
| `/home/activities/create` | ActivityForm | `/api/activities` (POST) | `activity:manage` |
| `/home/activities/:id/edit` | ActivityForm | `/api/activities/:id` (PUT) | `activity:manage` |
| `/home/activities/:id/redeem-codes` | RedeemCodes | `/api/activities/:id/redeem-codes` | `activity:manage` |
| `/home/activities/:id/redeem-codes/generate` | RedeemCodeGenerate | `/api/activities/:id/redeem-codes/generate` | `activity:manage` |
| `/home/activities/:id/redeem-codes/import-export` | RedeemCodeImportExport | `/api/activities/:id/redeem-codes/export` | `activity:manage` |
| `/home/activities/:id` | ActivityDetail | `/api/activities/:id` | `activity:manage` |
| `/home/redeem-codes/verify` | RedeemCodeVerify | `/api/redeem-codes/verify` | `activity:manage` |
| `/home/search` | Search | `/api/search/*` | 需认证 |

## 4.2 C 端商城前端路由

> **路由文件**: `shop-frontend/src/router/index.js`

| 前端路由路径 | 组件 | 对应后端 API | 认证 |
| :--- | :--- | :--- | :--- |
| `/` | Home | `/api/products`, `/api/activities` | 否 |
| `/products` | Products | `/api/products` | 否 |
| `/product/:id` | ProductDetail | `/api/products/:id` | 否 |
| `/cart` | Cart | `/api/cart` | 是 |
| `/login` | Login | `/api/auth/login` | 否 |
| `/register` | Register | `/api/auth/register` | 否 |
| `/customer/profile` | CustomerProfile | `/api/user/profile` | 是 |
| `/orders` | OrderList | `/api/user/orders` | 是 |
| `/order/:id` | OrderDetail | `/api/orders/:orderNo` | 是 |
| `/addresses` | AddressList | `/api/customer/addresses` | 是 |
| `/address/edit/:id?` | AddressEdit | `/api/customer/addresses/:id` | 是 |
| `/checkout` | OrderConfirm | `/api/orders` (POST) | 是 |
| `/activity/:id` | ActivityDetail | `/api/activities/:id` | 否 |
| `/activity/order/confirm` | ActivityOrderConfirm | `/api/activity-orders` (POST) | 是 |
| `/activity/orders` | ActivityOrderList | `/api/activity-orders` | 是 |
| `/activity/order/:id` | ActivityOrderDetail | `/api/activity-orders/:id` | 是 |
| `/messages` | MessageList | - | 是 |
| `/customer/service` | CustomerService | - | 是 |
| `/search` | SearchResults | `/api/search/*` | 是 |
