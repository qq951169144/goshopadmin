# GoShopAdmin 前端开发规则

本文件定义了AI在执行前端开发任务时必须强制遵守的规范。

---

## 规则1：项目结构规范

### 1.1 目录结构

**强制要求**：所有前端项目必须遵循以下目录结构：

```
├── src/
│   ├── api/         # API 调用相关
│   ├── assets/      # 静态资源
│   ├── components/  # 公共组件
│   ├── router/      # 路由配置
│   ├── store/       # 状态管理（如 Pinia）
│   ├── views/       # 页面组件
│   ├── App.vue      # 根组件
│   └── main.js      # 入口文件
├── public/          # 公共静态文件
├── .env.example     # 环境变量示例
├── index.html       # HTML 模板
├── package.json     # 项目配置
└── vite.config.js   # Vite 配置
```

### 1.2 页面组件组织

**推荐做法**：
- 页面组件放在 `views/` 目录下
- 复杂页面可以按功能创建子目录（如 `views/products/`）
- 公共组件放在 `components/` 目录下

---

## 规则2：技术栈规范

### 2.1 核心依赖

**强制要求**：所有前端项目必须使用以下核心依赖：

| 依赖 | 版本 | 用途 |
| :--- | :--- | :--- |
| Vue | ^3.5.0 | 前端框架 |
| Vue Router | ^4.4.0 | 路由管理 |
| Axios | ^1.13.0 | 网络请求 |
| Element Plus | ^2.13.0 | UI 组件库 |
| Vite | ^7.3.0 | 构建工具 |

### 2.2 可选依赖

**推荐使用**：
- Pinia (^2.1.0) - 状态管理（C端商城推荐）
- Quill (^2.0.0) - 富文本编辑器

---

## 规则3：路由规范

### 3.1 路由配置

**强制要求**：路由配置必须统一使用 `createRouter` 和 `createWebHistory`：

```javascript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 路由配置
  ]
})

export default router
```

### 3.2 路由命名

**强制要求**：
- 后台管理系统：使用 PascalCase 命名路由和组件（如 `ProductCategories`）
- C端商城：使用 kebab-case 命名路由（如 `customer-profile`），使用动态导入

### 3.3 路由结构

**推荐做法**：
- 后台管理系统：使用嵌套路由结构，Home 作为父路由
- C端商城：使用扁平路由结构，直接定义所有页面路由

### 3.4 路由守卫

**强制要求**：需要认证的页面必须使用路由守卫进行保护：

```javascript
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path === '/login') {
    next()
  } else {
    if (token) {
      next()
    } else {
      next('/login')
    }
  }
})
```

---

## 规则4：API 调用规范

### 4.1 API 配置

**强制要求**：所有 API 调用必须通过统一的 Axios 实例：

```javascript
import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})
```

### 4.2 拦截器

**强制要求**：必须配置请求和响应拦截器：

1. **请求拦截器**：
   - 添加 token 到请求头
   - 添加时间戳防止缓存
   - 处理重复请求

2. **响应拦截器**：
   - 统一处理业务错误
   - 统一处理 HTTP 错误
   - 提取响应数据

### 4.3 API 方法命名

**强制要求**：API 方法必须按功能分组导出：

```javascript
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  logout: () => api.post('/auth/logout')
}

export const productAPI = {
  getProducts: (params) => api.get('/products', { params }),
  getProductDetail: (id) => api.get(`/products/${id}`)
}
```

---

## 规则5：代码风格规范

### 5.1 组件写法

**强制要求**：使用 Composition API (setup 语法)：

```vue
<script setup>
import { ref, reactive, onMounted } from 'vue'

// 组件逻辑
</script>
```

### 5.2 命名规范

**强制要求**：
- 变量：使用 camelCase
- 组件：使用 PascalCase
- 路由：
  - 后台管理系统：使用 PascalCase
  - C端商城：使用 kebab-case
- API 对象：使用 camelCase（如 `authApi`）或大写（如 `authAPI`），保持项目内一致

### 5.3 代码风格

**推荐做法**：
- 使用 ES6+ 语法
- 保持代码缩进一致（2或4个空格）
- 项目内保持分号使用一致
- 添加适当的注释说明复杂逻辑

### 5.4 错误处理

**强制要求**：
- 使用 try/catch 处理异步操作错误
- 统一使用 Element Plus 的 ElMessage 显示错误信息
- 登录失败或 token 过期时跳转到登录页

---

## 规则6：状态管理规范

### 6.1 Pinia 使用

**推荐做法**：C端商城项目推荐使用 Pinia 进行状态管理：

```javascript
// store/cart.js
import { defineStore } from 'pinia'

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: []
  }),
  actions: {
    addItem(item) {
      this.items.push(item)
    }
  }
})
```

### 6.2 本地存储

**强制要求**：
- 使用 localStorage 存储 token 和用户信息
- 登录成功后保存 token：`localStorage.setItem('token', response.token)`
- 登出或 token 过期时清除 localStorage：`localStorage.removeItem('token')`

---

## 规则7：UI 规范

### 7.1 组件使用

**强制要求**：
- 后台管理系统：使用 Element Plus 组件
- C端商城：可以使用 Element Plus 组件或原生 HTML 元素

### 7.2 样式规范

**推荐做法**：
- 使用 scoped 样式
- 保持 CSS 命名一致（如 BEM 规范）
- 使用 CSS 变量管理主题色
- 确保响应式设计

---

## 规则8：性能优化

### 8.1 代码分割

**推荐做法**：
- C端商城使用动态导入路由组件
- 大型组件使用懒加载

### 8.2 网络请求优化

**强制要求**：
- 实现重复请求取消机制
- 添加请求超时设置
- 合理使用缓存策略

---

## 规则9：安全规范

### 9.1 XSS 防护

**推荐做法**：
- 使用 Vue 的 v-html 指令时注意内容安全
- 对用户输入进行验证和转义

### 9.2 CSRF 防护

**推荐做法**：
- 使用 token 验证
- 遵循后端 API 的安全要求

---

## 规则10：部署规范

### 10.1 构建命令

**强制要求**：使用 Vite 的构建命令：

```bash
npm run build
```

### 10.2 环境变量

**强制要求**：使用 .env 文件管理环境变量：

```env
VITE_API_BASE_URL=http://localhost:8000/api
```

---

## 项目类型区分

### 后台管理系统（frontend）

**特点**：
- 使用嵌套路由结构
- 更复杂的 UI 组件
- 功能导向的页面设计
- 推荐使用 PascalCase 命名

### C端商城（shop-frontend）

**特点**：
- 使用扁平路由结构
- 更简洁的 UI 设计
- 用户体验导向的页面设计
- 推荐使用 kebab-case 命名
- 推荐使用 Pinia 进行状态管理

---

## 开发流程规范

1. **初始化项目**：使用 Vite 创建 Vue 3 项目
2. **安装依赖**：添加核心依赖和可选依赖
3. **配置路由**：根据项目类型配置路由结构
4. **配置 API**：创建统一的 API 调用实例
5. **开发页面**：按照规范开发页面组件
6. **测试**：确保功能正常，无错误
7. **构建**：运行构建命令生成生产版本
8. **部署**：部署到服务器

---

## 违规示例

以下代码是**不允许**的：

```javascript
// ❌ 错误：使用 CommonJS 语法
const Vue = require('vue')

// ❌ 错误：直接使用 axios 而不是统一实例
axios.get('/api/products')

// ❌ 错误：硬编码 API 地址
const api = axios.create({
  baseURL: 'http://localhost:8000/api'
})

// ❌ 错误：没有使用路由守卫
router.beforeEach(() => {
  // 没有验证 token
})

// ❌ 错误：使用 Options API
<script>
export default {
  data() {
    return {}
  }
}
</script>
```

---

*规则版本: 1.0*
*最后更新: 2026-03-28*
