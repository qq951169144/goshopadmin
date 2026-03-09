# API接口和前端对接情况

## 一、后端API接口实现

### 1. 用户管理API

| 接口路径 | 方法 | 功能描述 | 请求参数 | 成功响应 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/users` | `GET` | 获取用户列表 | 无 | `{"code": 200, "message": "获取用户列表成功", "data": [...]}` |
| `/api/users/:id` | `GET` | 获取单个用户信息 | 无 | `{"code": 200, "message": "获取用户信息成功", "data": {...}}` |
| `/api/users` | `POST` | 创建用户 | `{"username": "...", "password": "...", "role_id": 1}` | `{"code": 200, "message": "创建用户成功", "data": {...}}` |
| `/api/users/:id` | `PUT` | 更新用户 | `{"password": "...", "role_id": 1, "status": "active"}` | `{"code": 200, "message": "更新用户成功", "data": {...}}` |
| `/api/users/:id` | `DELETE` | 删除用户 | 无 | `{"code": 200, "message": "删除用户成功"}` |

### 2. 角色管理API

| 接口路径 | 方法 | 功能描述 | 请求参数 | 成功响应 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/roles` | `GET` | 获取角色列表 | 无 | `{"code": 200, "message": "获取角色列表成功", "data": [...]}` |
| `/api/roles/:id` | `GET` | 获取单个角色信息 | 无 | `{"code": 200, "message": "获取角色信息成功", "data": {...}}` |
| `/api/roles` | `POST` | 创建角色 | `{"name": "...", "description": "..."}` | `{"code": 200, "message": "创建角色成功", "data": {...}}` |
| `/api/roles/:id` | `PUT` | 更新角色 | `{"name": "...", "description": "..."}` | `{"code": 200, "message": "更新角色成功", "data": {...}}` |
| `/api/roles/:id` | `DELETE` | 删除角色 | 无 | `{"code": 200, "message": "删除角色成功"}` |
| `/api/roles/:id/permissions` | `POST` | 为角色分配权限 | `{"permission_ids": [1, 2, 3]}` | `{"code": 200, "message": "分配权限成功"}` |

### 3. 权限管理API

| 接口路径 | 方法 | 功能描述 | 请求参数 | 成功响应 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/permissions` | `GET` | 获取权限列表 | 无 | `{"code": 200, "message": "获取权限列表成功", "data": [...]}` |
| `/api/permissions/:id` | `GET` | 获取单个权限信息 | 无 | `{"code": 200, "message": "获取权限信息成功", "data": {...}}` |
| `/api/permissions` | `POST` | 创建权限 | `{"name": "...", "code": "...", "description": "..."}` | `{"code": 200, "message": "创建权限成功", "data": {...}}` |
| `/api/permissions/:id` | `PUT` | 更新权限 | `{"name": "...", "code": "...", "description": "..."}` | `{"code": 200, "message": "更新权限成功", "data": {...}}` |
| `/api/permissions/:id` | `DELETE` | 删除权限 | 无 | `{"code": 200, "message": "删除权限成功"}` |

### 4. 商户管理API

| 接口路径 | 方法 | 功能描述 | 请求参数 | 成功响应 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/merchants` | `GET` | 获取商户列表 | 无 | `{"code": 200, "message": "获取商户列表成功", "data": [...]}` |
| `/api/merchants/:id` | `GET` | 获取单个商户信息 | 无 | `{"code": 200, "message": "获取商户信息成功", "data": {...}}` |
| `/api/merchants` | `POST` | 创建商户 | `{"name": "...", "contact_person": "...", "contact_phone": "...", "address": "..."}` | `{"code": 200, "message": "创建商户成功", "data": {...}}` |
| `/api/merchants/:id` | `PUT` | 更新商户信息 | `{"name": "...", "contact_person": "...", "contact_phone": "...", "address": "...", "status": "..."}` | `{"code": 200, "message": "更新商户成功", "data": {...}}` |
| `/api/merchants/:id` | `DELETE` | 禁用商户 | 无 | `{"code": 200, "message": "禁用商户成功"}` |
| `/api/merchants/:id/audit` | `PUT` | 审核商户 | `{"audit_status": "approved", "audit_note": "..."}` | `{"code": 200, "message": "审核商户成功", "data": {...}}` |
| `/api/merchants/:id/users` | `GET` | 获取商户用户列表 | 无 | `{"code": 200, "message": "获取商户用户列表成功", "data": [...]}` |
| `/api/merchants/:id/users` | `POST` | 为商户添加用户 | `{"user_id": 1}` | `{"code": 200, "message": "添加商户用户成功"}` |
| `/api/merchants/:id/users/:user_id` | `DELETE` | 从商户移除用户 | 无 | `{"code": 200, "message": "移除商户用户成功"}` |

### 5. 商品管理API

| 接口路径 | 方法 | 功能描述 | 请求参数 | 成功响应 |
| :--- | :--- | :--- | :--- | :--- |
| `/api/products` | `GET` | 获取商品列表 | 无 | `{"code": 200, "message": "获取商品列表成功", "data": [...]}` |
| `/api/products/:id` | `GET` | 获取商品详情 | 无 | `{"code": 200, "message": "获取商品详情成功", "data": {...}}` |
| `/api/products` | `POST` | 创建商品 | `{"name": "...", "description": "...", "price": 100, "stock": 10, "category_id": 1}` | `{"code": 200, "message": "创建商品成功", "data": {...}}` |
| `/api/products/:id` | `PUT` | 更新商品 | `{"name": "...", "description": "...", "price": 100, "stock": 10, "category_id": 1, "status": "active"}` | `{"code": 200, "message": "更新商品成功", "data": {...}}` |
| `/api/products/:id` | `DELETE` | 删除商品 | 无 | `{"code": 200, "message": "删除商品成功"}` |
| `/api/product-categories` | `GET` | 获取商品分类列表 | 无 | `{"code": 200, "message": "获取分类列表成功", "data": [...]}` |
| `/api/product-categories/:id` | `GET` | 获取商品分类详情 | 无 | `{"code": 200, "message": "获取分类详情成功", "data": {...}}` |
| `/api/product-categories` | `POST` | 创建商品分类 | `{"name": "...", "parent_id": 0, "sort": 0}` | `{"code": 200, "message": "创建分类成功", "data": {...}}` |
| `/api/product-categories/:id` | `PUT` | 更新商品分类 | `{"name": "...", "parent_id": 0, "sort": 0, "status": "active"}` | `{"code": 200, "message": "更新分类成功", "data": {...}}` |
| `/api/product-categories/:id` | `DELETE` | 删除商品分类 | 无 | `{"code": 200, "message": "删除分类成功"}` |
| `/api/product-images` | `POST` | 添加商品图片 | `{"product_id": 1, "image_url": "...", "is_main": true, "sort": 0}` | `{"code": 200, "message": "添加图片成功", "data": {...}}` |
| `/api/product-images/:id` | `DELETE` | 删除商品图片 | 无 | `{"code": 200, "message": "删除图片成功"}` |
| `/api/product-skus` | `POST` | 添加商品SKU | `{"product_id": 1, "sku_code": "...", "attributes": "{\"color\": \"red\", \"size\": \"M\"}", "price": 100, "stock": 10}` | `{"code": 200, "message": "添加SKU成功", "data": {...}}` |
| `/api/product-skus/:id` | `PUT` | 更新商品SKU | `{"sku_code": "...", "attributes": "{\"color\": \"red\", \"size\": \"M\"}", "price": 100, "stock": 10, "status": "active"}` | `{"code": 200, "message": "更新SKU成功", "data": {...}}` |
| `/api/product-skus/:id` | `DELETE` | 删除商品SKU | 无 | `{"code": 200, "message": "删除SKU成功"}` |

## 二、前端页面功能实现

### 1. 用户管理页面

- **功能**：
  - 显示用户列表，包括ID、用户名、角色、状态等信息
  - 支持创建新用户，填写用户名、密码、角色、状态
  - 支持编辑现有用户，修改密码、角色、状态
  - 支持删除用户，带确认对话框
  - 状态显示为标签，活跃为绿色，禁用为红色

- **实现**：
  - 使用Element Plus的Table组件展示用户列表
  - 使用Dialog组件实现创建和编辑用户的表单
  - 使用Select组件选择角色和状态
  - 使用MessageBox组件实现删除确认
  - 调用后端API接口实现数据交互

### 2. 角色管理页面

- **功能**：
  - 显示角色列表，包括ID、角色名称、描述等信息
  - 支持创建新角色，填写角色名称、描述
  - 支持编辑现有角色，修改角色名称、描述
  - 支持删除角色，带确认对话框
  - 支持为角色分配权限，选择多个权限

- **实现**：
  - 使用Element Plus的Table组件展示角色列表
  - 使用Dialog组件实现创建和编辑角色的表单
  - 使用CheckboxGroup组件选择权限
  - 使用MessageBox组件实现删除确认
  - 调用后端API接口实现数据交互

### 3. 权限管理页面

- **功能**：
  - 显示权限列表，包括ID、权限名称、权限代码、描述等信息
  - 支持创建新权限，填写权限名称、权限代码、描述
  - 支持编辑现有权限，修改权限名称、权限代码、描述
  - 支持删除权限，带确认对话框

- **实现**：
  - 使用Element Plus的Table组件展示权限列表
  - 使用Dialog组件实现创建和编辑权限的表单
  - 使用MessageBox组件实现删除确认
  - 调用后端API接口实现数据交互

### 4. 商户管理页面

- **功能**：
  - 显示商户列表，包括ID、商户名称、联系人、联系电话、审核状态、状态等信息
  - 支持创建新商户，填写商户名称、联系人、联系电话、地址
  - 支持编辑现有商户，修改商户信息和状态
  - 支持审核商户，设置审核状态和审核备注
  - 支持禁用商户，带确认对话框
  - 支持管理商户用户，查看、添加和移除商户用户

- **实现**：
  - 使用Element Plus的Table组件展示商户列表
  - 使用Dialog组件实现创建和编辑商户的表单
  - 使用Dialog组件实现商户审核功能
  - 使用Dialog组件实现商户用户管理功能
  - 使用MessageBox组件实现禁用确认
  - 调用后端API接口实现数据交互

### 5. 商品管理页面

- **功能**：
  - 显示商品列表，包括ID、商品名称、价格、库存、分类、状态等信息
  - 支持创建新商品，填写商品名称、描述、价格、库存、分类等信息
  - 支持编辑现有商品，修改商品信息和状态
  - 支持删除商品，带确认对话框
  - 支持管理商品图片，添加和删除图片
  - 支持管理商品SKU，添加、编辑和删除SKU
  - 支持预览C端商品展示页面

- **实现**：
  - 使用Element Plus的Table组件展示商品列表
  - 使用Dialog组件实现创建和编辑商品的表单
  - 使用富文本编辑器编辑商品详情
  - 使用Dialog组件实现商品图片管理功能
  - 使用Dialog组件实现商品SKU管理功能
  - 使用MessageBox组件实现删除确认
  - 调用后端API接口实现数据交互

### 6. 商品分类管理页面

- **功能**：
  - 显示商品分类列表，包括ID、分类名称、父分类、排序等信息
  - 支持创建新分类，填写分类名称、父分类、排序等信息
  - 支持编辑现有分类，修改分类信息和状态
  - 支持删除分类，带确认对话框

- **实现**：
  - 使用Element Plus的Table组件展示分类列表
  - 使用Dialog组件实现创建和编辑分类的表单
  - 使用Tree组件展示分类层级关系
  - 使用MessageBox组件实现删除确认
  - 调用后端API接口实现数据交互

## 三、技术实现细节

### 1. 后端技术栈

- **语言**：Go 1.20+
- **Web框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL
- **认证**：JWT
- **密码加密**：bcrypt

### 2. 前端技术栈

- **框架**：Vue 3
- **构建工具**：Vite
- **UI库**：Element Plus
- **HTTP客户端**：Axios
- **状态管理**：Vue的ref和reactive

### 3. 数据结构

- **用户表** (`users`)：id, username, password, role_id, status, created_at, updated_at
- **角色表** (`roles`)：id, name, description, created_at, updated_at
- **权限表** (`permissions`)：id, name, code, description, created_at, updated_at
- **角色权限关联表** (`role_permissions`)：role_id, permission_id
- **商户表** (`merchants`)：id, name, contact_person, contact_phone, address, audit_status, status, created_at, updated_at
- **商户用户关联表** (`merchant_users`)：merchant_id, user_id, created_at
- **商户审核表** (`merchant_audits`)：id, merchant_id, audit_status, audit_note, audited_by, audited_at
- **商户银行信息表** (`merchant_banks`)：id, merchant_id, bank_name, account_name, account_number, status, created_at, updated_at
- **商户提现表** (`merchant_withdraws`)：id, merchant_id, amount, status, bank_id, created_at, updated_at
- **商户对账单表** (`merchant_statements`)：id, merchant_id, type, amount, balance, description, created_at

## 四、测试结果

### 1. 后端API测试

- ✅ 用户管理API：所有接口正常工作
- ✅ 角色管理API：所有接口正常工作
- ✅ 权限管理API：所有接口正常工作
- ✅ 商户管理API：所有接口正常工作
- ✅ 认证中间件：正确保护需要认证的接口

### 2. 前端页面测试

- ✅ 用户管理页面：所有功能正常，数据显示正确
- ✅ 角色管理页面：所有功能正常，权限分配功能正常
- ✅ 权限管理页面：所有功能正常，数据显示正确
- ✅ 商户管理页面：所有功能正常，审核流程顺畅
- ✅ 页面交互：响应及时，操作流畅

## 五、总结

本次实现完成了商城后台管理系统的用户管理、角色管理、权限管理和商户管理功能，包括：

1. **后端API接口**：实现了完整的CRUD操作，支持用户、角色、权限和商户的管理
2. **前端页面**：实现了直观易用的管理界面，支持各种操作功能
3. **数据交互**：前后端对接正常，数据流转顺畅
4. **安全性**：使用JWT认证，密码加密存储，权限控制严格
5. **商户管理**：实现了商户注册、审核、信息管理和用户关联等功能

系统已经具备了完整的用户-角色-权限管理体系和商户管理功能，可以满足商城后台的管理需求。