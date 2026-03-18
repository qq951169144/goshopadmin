# API接口和前端对接情况

## 一、后端API接口实现

### 1. 后台用户管理API

| 接口路径             | 方法       | 功能描述     | 请求参数                                                                                                   | 成功响应                                                  |
| :--------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/users`     | `GET`    | 获取用户列表   | 无                                                                                                          | `{"code": 200, "message": "获取用户列表成功", "data": [...]}` |
| `/api/users/:id` | `GET`    | 获取单个用户信息 | 无                                                                                                          | `{"code": 200, "message": "获取用户信息成功", "data": {...}}` |
| `/api/users`     | `POST`   | 创建用户     | `{"username": "string", "password": "string", "role_id": 1, "status": "active"}`  | `{"code": 200, "message": "创建用户成功", "data": {...}}`   |
| `/api/users/:id` | `PUT`    | 更新用户     | `{"username": "string", "role_id": 1, "status": "active"}` | `{"code": 200, "message": "更新用户成功", "data": {...}}`   |
| `/api/users/:id` | `DELETE` | 删除用户     | 无                                                                                                          | `{"code": 200, "message": "删除用户成功"}`                  |

### 2. 角色管理API

| 接口路径                         | 方法       | 功能描述     | 请求参数                                                                                                   | 成功响应                                                  |
| :--------------------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/roles`                 | `GET`    | 获取角色列表   | 无                                                                                                          | `{"code": 200, "message": "获取角色列表成功", "data": [...]}` |
| `/api/roles/:id`             | `GET`    | 获取单个角色信息 | 无                                                                                                          | `{"code": 200, "message": "获取角色信息成功", "data": {...}}` |
| `/api/roles`                 | `POST`   | 创建角色     | `{"name": "string", "description": "string", "status": "active"}` | `{"code": 200, "message": "创建角色成功", "data": {...}}`   |
| `/api/roles/:id`             | `PUT`    | 更新角色     | `{"name": "string", "description": "string", "status": "active"}` | `{"code": 200, "message": "更新角色成功", "data": {...}}`   |
| `/api/roles/:id`             | `DELETE` | 删除角色     | 无                                                                                                          | `{"code": 200, "message": "删除角色成功"}`                  |
| `/api/roles/:id/permissions` | `POST`   | 为角色分配权限  | `{"permission_ids": [1, 2, 3]}`         | `{"code": 200, "message": "分配权限成功"}`                  |

### 3. 权限管理API

| 接口路径                   | 方法       | 功能描述     | 请求参数                                                                                                                   | 成功响应                                                  |
| :--------------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/permissions`     | `GET`    | 获取权限列表   | 无                                                                                                                      | `{"code": 200, "message": "获取权限列表成功", "data": [...]}` |
| `/api/permissions/:id` | `GET`    | 获取单个权限信息 | 无                                                                                                                      | `{"code": 200, "message": "获取权限信息成功", "data": {...}}` |
| `/api/permissions`     | `POST`   | 创建权限     | `{"name": "string", "code": "string", "description": "string", "status": "active"}` | `{"code": 200, "message": "创建权限成功", "data": {...}}`   |
| `/api/permissions/:id` | `PUT`    | 更新权限     | `{"name": "string", "code": "string", "description": "string", "status": "active"}` | `{"code": 200, "message": "更新权限成功", "data": {...}}`   |
| `/api/permissions/:id` | `DELETE` | 删除权限     | 无                                                                                                                      | `{"code": 200, "message": "删除权限成功"}`                  |

### 4. 商户管理API

| 接口路径                                | 方法       | 功能描述     | 请求参数                                                                                                  | 成功响应                                                    |
| :---------------------------------- | :------- | :------- | :---------------------------------------------------------------------------------------------------- | :------------------------------------------------------ |
| `/api/merchants`                    | `GET`    | 获取商户列表   | 无                                                                                                     | `{"code": 200, "message": "获取商户列表成功", "data": [...]}`   |
| `/api/merchants/:id`                | `GET`    | 获取单个商户信息 | 无                                                                                                     | `{"code": 200, "message": "获取商户信息成功", "data": {...}}`   |
| `/api/merchants`                    | `POST`   | 创建商户     | `{"name": "string", "contact_name": "string", "contact_phone": "string", "email": "string", "address": "string", "business_license": "string", "tax_number": "string"}`                  | `{"code": 200, "message": "创建商户成功", "data": {...}}`     |
| `/api/merchants/:id`                | `PUT`    | 更新商户信息   | `{"name": "string", "contact_name": "string", "contact_phone": "string", "email": "string", "address": "string", "business_license": "string", "tax_number": "string", "status": "active"}` | `{"code": 200, "message": "更新商户成功", "data": {...}}`     |
| `/api/merchants/:id`                | `DELETE` | 禁用商户     | 无                                                                                                     | `{"code": 200, "message": "禁用商户成功"}`                    |
| `/api/merchants/:id/audit`          | `PUT`    | 审核商户     | `{"audit_status": "pending/approved/rejected", "audit_note": "string"}`                                                   | `{"code": 200, "message": "审核商户成功", "data": {...}}`     |
| `/api/merchants/:id/users`          | `GET`    | 获取商户用户列表 | 无                                                                                                     | `{"code": 200, "message": "获取商户用户列表成功", "data": [...]}` |
| `/api/merchants/:id/users`          | `POST`   | 为商户添加用户  | `{"user_id": 1, "role": "owner/manager/staff"}`                                                                                      | `{"code": 200, "message": "添加商户用户成功"}`                  |
| `/api/merchants/:id/users/:user_id` | `DELETE` | 从商户移除用户  | 无                                                                                                     | `{"code": 200, "message": "移除商户用户成功"}`                  |

### 5. 商品管理API

| 接口路径                          | 方法       | 功能描述     | 请求参数                                                                                                                        | 成功响应                                                  |
| :---------------------------- | :------- | :------- | :-------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/products`               | `GET`    | 获取商品列表   | 无                                                                                                                           | `{"code": 200, "message": "获取商品列表成功", "data": [...]}` |
| `/api/products/:id`           | `GET`    | 获取商品详情   | 无                                                                                                                           | `{"code": 200, "message": "获取商品详情成功", "data": {...}}` |
| `/api/products`               | `POST`   | 创建商品     | `{"name": "string", "description": "string", "price": 100.00, "stock": 10, "category_id": 1, "status": "active"}`                                        | `{"code": 200, "message": "创建商品成功", "data": {...}}`   |
| `/api/products/:id`           | `PUT`    | 更新商品     | `{"name": "string", "description": "string", "price": 100.00, "stock": 10, "category_id": 1, "status": "active"}`                    | `{"code": 200, "message": "更新商品成功", "data": {...}}`   |
| `/api/products/:id`           | `DELETE` | 删除商品     | 无                                                                                                                           | `{"code": 200, "message": "删除商品成功"}`                  |
| `/api/product-categories`     | `GET`    | 获取商品分类列表 | 无                                                                                                                           | `{"code": 200, "message": "获取分类列表成功", "data": [...]}` |
| `/api/product-categories/:id` | `GET`    | 获取商品分类详情 | 无                                                                                                                           | `{"code": 200, "message": "获取分类详情成功", "data": {...}}` |
| `/api/product-categories`     | `POST`   | 创建商品分类   | `{"name": "string", "parent_id": 0, "level": 1, "sort": 0, "status": "active"}`                                                                                | `{"code": 200, "message": "创建分类成功", "data": {...}}`   |
| `/api/product-categories/:id` | `PUT`    | 更新商品分类   | `{"name": "string", "parent_id": 0, "level": 1, "sort": 0, "status": "active"}`                                                            | `{"code": 200, "message": "更新分类成功", "data": {...}}`   |
| `/api/product-categories/:id` | `DELETE` | 删除商品分类   | 无                                                                                                                           | `{"code": 200, "message": "删除分类成功"}`                  |
| `/api/product-images`         | `POST`   | 添加商品图片   | `{"product_id": 1, "image_url": "string", "is_main": true, "sort": 0}`                                                         | `{"code": 200, "message": "添加图片成功", "data": {...}}`   |
| `/api/product-images/:id`     | `DELETE` | 删除商品图片   | 无                                                                                                                           | `{"code": 200, "message": "删除图片成功"}`                  |
| `/api/product-skus`           | `POST`   | 添加商品SKU  | `{"product_id": 1, "sku_code": "string", "attributes": "{\"color\": \"red\", \"size\": \"M\"}", "price": 100.00, "original_price": 120.00, "stock": 10, "status": "active"}`    | `{"code": 200, "message": "添加SKU成功", "data": {...}}`  |
| `/api/product-skus/:id`       | `PUT`    | 更新商品SKU  | `{"product_id": 1, "sku_code": "string", "attributes": "{\"color\": \"red\", \"size\": \"M\"}", "price": 100.00, "original_price": 120.00, "stock": 10, "status": "active"}` | `{"code": 200, "message": "更新SKU成功", "data": {...}}`  |
| `/api/product-skus/:id`       | `DELETE` | 删除商品SKU  | 无                                                                                                                           | `{"code": 200, "message": "删除SKU成功"}`                 |
| `/api/products/:id/specifications` | `GET`    | 获取商品规格列表 | 无                                                                                                                           | `{"code": 200, "message": "获取规格列表成功", "data": [...]}` |
| `/api/products/:id/specifications` | `POST`   | 创建商品规格   | `{"name": "颜色", "sort": 1}`                                                                                              | `{"code": 200, "message": "创建规格成功", "data": {...}}`   |
| `/api/specifications/:id`     | `PUT`    | 更新商品规格   | `{"name": "颜色", "sort": 1}`                                                                                              | `{"code": 200, "message": "更新规格成功", "data": {...}}`   |
| `/api/specifications/:id`     | `DELETE` | 删除商品规格   | 无                                                                                                                           | `{"code": 200, "message": "删除规格成功"}`                  |
| `/api/specifications/:id/values` | `GET`    | 获取规格值列表  | 无                                                                                                                           | `{"code": 200, "message": "获取规格值列表成功", "data": [...]}` |
| `/api/specifications/:id/values` | `POST`   | 创建规格值     | `{"value": "红色", "sort": 1, "status": "active"}`                                                                         | `{"code": 200, "message": "创建规格值成功", "data": {...}}` |
| `/api/specification-values/:id` | `PUT`    | 更新规格值     | `{"value": "红色", "sort": 1, "status": "active"}`                                                                         | `{"code": 200, "message": "更新规格值成功", "data": {...}}` |
| `/api/specification-values/:id` | `DELETE` | 删除规格值     | 无                                                                                                                           | `{"code": 200, "message": "删除规格值成功"}`                |
| `/api/product-skus/:id/specs` | `GET`    | 获取SKU规格关联 | 无                                                                                                                           | `{"code": 200, "message": "获取SKU规格成功", "data": [...]}` |
| `/api/product-skus/:id/specs` | `POST`   | 设置SKU规格关联 | `{"spec_id": 1, "spec_value_id": 2}`                                                                                       | `{"code": 200, "message": "设置SKU规格成功", "data": {...}}` |

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
  - 支持创建新商品，填写商品名称、描述、详情、分类等信息
  - 支持编辑现有商品，修改商品信息和状态
  - 支持删除商品，带确认对话框
  - 支持管理商品图片，添加和删除图片，设置主图
  - 支持管理商品规格，创建规格（如颜色、尺寸）和规格值（如红色、M码）
  - 支持管理商品SKU，添加、编辑和删除SKU，关联规格组合
  - 支持预览C端商品展示页面
- **实现**：
  - 使用Element Plus的Table组件展示商品列表
  - 使用Dialog组件实现创建和编辑商品的表单
  - 使用富文本编辑器编辑商品详情
  - 使用Dialog组件实现商品图片管理功能
  - 使用Dialog组件实现商品规格管理功能
  - 使用Dialog组件实现商品SKU管理功能，支持规格组合选择
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

- **用户表** (`users`)：id, username, password, role\_id, status, created\_at, updated\_at
- **角色表** (`roles`)：id, name, description, created\_at, updated\_at
- **权限表** (`permissions`)：id, name, code, description, created\_at, updated\_at
- **角色权限关联表** (`role_permissions`)：role\_id, permission\_id
- **商户表** (`merchants`)：id, name, contact\_person, contact\_phone, address, audit\_status, status, created\_at, updated\_at
- **商户用户关联表** (`merchant_users`)：merchant\_id, user\_id, created\_at
- **商户审核表** (`merchant_audits`)：id, merchant\_id, audit\_status, audit\_note, audited\_by, audited\_at
- **商户银行信息表** (`merchant_banks`)：id, merchant\_id, bank\_name, account\_name, account\_number, status, created\_at, updated\_at
- **商户提现表** (`merchant_withdraws`)：id, merchant\_id, amount, status, bank\_id, created\_at, updated\_at
- **商户对账单表** (`merchant_statements`)：id, merchant\_id, type, amount, balance, description, created\_at

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

## 6. C端商城API

### 6.1 认证API

| 接口路径                 | 方法       | 功能描述     | 请求参数                                                                                                   | 成功响应                                                  |
| :------------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/auth/register` | `POST`   | 注册     | `{"username": "string", "password": "string", "captcha_id": "string", "captcha": "string"}` | `{"message": "Register success", "token": "..."}` |
| `/api/auth/login`    | `POST`   | 登录     | `{"username": "string", "password": "string", "captcha_id": "string", "captcha": "string"}` | `{"token": "..."}` |
| `/api/auth/logout`   | `POST`   | 登出     | 无                                                                                                          | `{"message": "Logout success"}`                  |

### 6.2 验证码API

| 接口路径                 | 方法       | 功能描述     | 请求参数 | 成功响应                                                  |
| :------------------- | :------- | :------- | :--- | :---------------------------------------------------- |
| `/api/captcha`       | `GET`    | 生成验证码   | 无    | 返回验证码图片（PNG格式），响应头包含X-Captcha-ID |
| `/api/captcha/verify` | `POST`   | 验证验证码   | `{"captcha_id": "string", "value": "string"}` | `{"valid": true}` |

### 6.3 用户API

| 接口路径                 | 方法       | 功能描述     | 请求参数 | 成功响应                                                  |
| :------------------- | :------- | :------- | :--- | :---------------------------------------------------- |
| `/api/user/profile`  | `GET`    | 获取个人信息  | 无    | `{"username": "testuser", "email": "test@example.com"}` |
| `/api/user/profile`  | `PUT`    | 更新个人信息  | `{"username": "string", "email": "string"}` | `{"message": "Profile updated", "username": "...", "email": "..."}` |
| `/api/user/orders`   | `GET`    | 获取订单列表  | `page`, `limit` | `{"orders": [...], "total": 2}` |

### 6.4 商品API

| 接口路径                 | 方法       | 功能描述     | 请求参数 | 成功响应                                                  |
| :------------------- | :------- | :------- | :--- | :---------------------------------------------------- |
| `/api/products`      | `GET`    | 获取商品列表  | `page`, `limit`, `category`, `keyword` | `{"products": [...], "total": 2}` |
| `/api/products/:id`  | `GET`    | 获取商品详情  | 无    | `{"id": 1, "name": "iPhone 13", "description": "...", "price": 5999.99, "sku": "...", "stock": 100, "image": "..."}` |

### 6.5 购物车API

| 接口路径                 | 方法       | 功能描述     | 请求参数                                                                                                   | 成功响应                                                  |
| :------------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/cart`          | `GET`    | 获取购物车   | 无                                                                                                          | `{"items": [{"id": 1, "product_id": 1, "product_name": "iPhone 13", "main_image": "...", "sku_id": 1, "sku_code": "128GB-Black", "quantity": 2, "price": 99.99}]}` |
| `/api/cart/items`    | `POST`   | 添加商品到购物车 | `{"product_id": 1, "sku_id": 1, "quantity": 2, "price": 99.99}` | `{"message": "Item added to cart"}` |
| `/api/cart/items/:id` | `PUT`    | 更新购物车项  | `{"quantity": 3}`                                                                                      | `{"message": "Cart item updated", "item_id": "...", "quantity": 3}` |
| `/api/cart/items/:id` | `DELETE` | 移除购物车项  | 无                                                                                                          | `{"message": "Cart item removed", "item_id": "..."}` |
| `/api/cart/sync`     | `POST`   | 同步购物车   | `{"items": [...]}`                                                                                        | `{"message": "Cart synced", "items": [...]}` |

### 6.6 订单API

| 接口路径                 | 方法       | 功能描述     | 请求参数                                                                                                   | 成功响应                                                  |
| :------------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/orders`        | `POST`   | 创建订单     | `{"items": [{"product_id": 1, "quantity": 2, "price": 99.99, "sku": "red-medium"}]}` | `{"order_id": "ORD202603120001", "amount": 199.98, "payment_url": "...", "status": "pending", "created_at": "..."}` |
| `/api/orders/:id`    | `GET`    | 获取订单详情  | 无                                                                                                          | `{"order_id": "...", "amount": 199.98, "status": "paid", "items": [...], "created_at": "...", "paid_at": "..."}` |

### 6.7 支付API

| 接口路径                 | 方法       | 功能描述     | 请求参数                                                                                                   | 成功响应                                                  |
| :------------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------- | :---------------------------------------------------- |
| `/api/payment/fake-pay` | `GET`    | 伪支付页面   | `order_id`                                                                                               | 返回支付成功页面（HTML）                                      |
| `/api/payment/callback` | `POST`   | 支付回调     | `{"order_id": "ORD202603120001", "transaction_id": "TRX1234567890", "status": "success", "amount": 199.98}` | `{"message": "Payment callback received", "order_id": "...", "transaction_id": "...", "status": "..."}` |

## 二、前端页面功能实现

### 1. 后台用户管理页面

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

### 7. C端商城页面

#### 7.1 首页
- **功能**：
  - ✅ 展示热门商品
  - ✅ 展示活动信息
  - ✅ 导航到商品列表和详情页
- **实现**：
  - ✅ 使用Vue 3的组件系统
  - ✅ 调用商品API获取商品列表

#### 7.2 商品列表页
- **功能**：
  - ✅ 展示商品列表
  - ✅ 支持分类筛选
  - ✅ 支持关键词搜索
  - ✅ 支持分页
- **实现**：
  - ✅ 使用Vue 3的组件系统
  - ✅ 调用商品API获取商品列表

#### 7.3 商品详情页
- **功能**：
  - ✅ 展示商品详情
  - ✅ 展示商品图片
  - ✅ 选择商品SKU
  - ✅ 添加商品到购物车
- **实现**：
  - ✅ 使用Vue 3的组件系统
  - ✅ 调用商品API获取商品详情
  - ✅ 调用购物车API添加商品

#### 7.4 购物车页面
- **功能**：
  - ✅ 展示购物车商品
  - ✅ 调整商品数量
  - ✅ 移除商品
  - ✅ 结算下单
- **实现**：
  - ✅ 使用Vue 3的组件系统
  - ✅ 使用Pinia进行状态管理
  - ✅ 调用购物车API获取和更新购物车

#### 7.5 登录/注册页面
- **功能**：
  - ✅ 用户登录
  - ✅ 用户注册
  - ✅ 验证码验证
- **实现**：
  - ✅ 使用Vue 3的组件系统
  - ✅ 调用认证API进行登录和注册
  - ✅ 调用验证码API获取和验证验证码

#### 7.6 个人中心页面
- **功能**：
  - ✅ 查看个人信息
  - ✅ 编辑个人信息
  - ✅ 查看订单列表
- **实现**：
  - ✅ 使用Vue 3的组件系统
  - ✅ 调用用户API获取和更新个人信息
  - ✅ 调用订单API获取订单列表

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
- **状态管理**：Vue的ref和reactive（后台），Pinia（C端）

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
- **C端用户表** (`customers`)：id, username, password, phone, email, status, created_at, updated_at, nickname, avatar, last_login_at, last_login_ip
- **地址表** (`addresses`)：id, customer_id, name, phone, province, city, district, detail_address, is_default, status, created_at, updated_at
- **订单表** (`orders`)：id, order_no, customer_id, merchant_id, total_amount, status, address_id, created_at, updated_at, payment_method, transaction_id, paid_at, shipped_at, delivered_at, cancelled_at
- **订单明细表** (`order_items`)：id, order_id, product_id, sku_id, product_name, sku_attributes, price, quantity, total_amount, created_at, updated_at
- **支付记录表** (`payments`)：id, order_id, payment_no, amount, payment_method, transaction_id, status, created_at, updated_at, paid_at
- **购物车表** (`carts`)：id, user_id, session_id, created_at, updated_at
- **购物车项表** (`cart_items`)：id, cart_id, product_id, sku_id, quantity, price, created_at, updated_at
- **商品规格表** (`product_specifications`)：id, product_id, name, sort, created_at, updated_at
- **规格值表** (`product_specification_values`)：id, spec_id, value, sort, status, created_at, updated_at
- **SKU规格关联表** (`product_sku_specs`)：id, sku_id, spec_id, spec_value_id, created_at

## 四、测试结果

### 1. 后端API测试

- ✅ 后台用户管理API：所有接口正常工作
- ✅ 角色管理API：所有接口正常工作
- ✅ 权限管理API：所有接口正常工作
- ✅ 商户管理API：所有接口正常工作
- ✅ C端认证API：所有接口正常工作
- ✅ C端购物车API：所有接口正常工作
- ✅ C端订单API：所有接口正常工作
- ✅ C端支付API：所有接口正常工作
- ✅ 认证中间件：正确保护需要认证的接口

### 2. 前端页面测试

- ✅ 后台用户管理页面：所有功能正常，数据显示正确
- ✅ 角色管理页面：所有功能正常，权限分配功能正常
- ✅ 权限管理页面：所有功能正常，数据显示正确
- ✅ 商户管理页面：所有功能正常，审核流程顺畅
- ✅ C端首页：商品展示正常，导航功能正常
- ✅ C端商品列表页：筛选和搜索功能正常
- ✅ C端商品详情页：商品信息展示完整，添加购物车功能正常
- ✅ C端购物车页面：商品管理功能正常，结算功能正常
- ✅ C端登录/注册页面：认证功能正常，验证码功能正常
- ✅ C端个人中心页面：个人信息管理和订单查看功能正常
- ✅ 页面交互：响应及时，操作流畅

## 五、总结

本次实现完成了商城后台管理系统和C端商城系统的功能，包括：

1. **后端API接口**：
   - 后台管理API：实现了用户、角色、权限、商户、商品等管理功能
   - C端商城API：实现了认证、购物车、订单、支付等核心功能

2. **前端页面**：
   - 后台管理页面：直观易用的管理界面，支持各种操作功能
   - C端商城页面：用户友好的购物界面，支持完整的购物流程

3. **数据交互**：
   - 前后端对接正常，数据流转顺畅
   - 支持无登录购物车功能

4. **安全性**：
   - 使用JWT认证，密码加密存储
   - 权限控制严格，保护敏感接口

5. **商户管理**：
   - 实现了商户注册、审核、信息管理和用户关联等功能

6. **C端功能**：
   - 实现了完整的购物流程，包括商品浏览、购物车管理、订单创建、支付等功能

系统已经具备了完整的商城后台管理和C端购物功能，可以满足商城的日常运营需求。
