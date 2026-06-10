# GoShopAdmin 项目索引

> AI 助手每次会话开始时读取此文件，快速了解项目全貌和关键信息位置。
> 详细账号密码见：`默认登录账号密码.md`

## 项目架构概览

GoShopAdmin 是一个多服务电商管理系统，包含后台管理、C端商城、搜索服务三大业务线，
以及 MySQL/Redis/RabbitMQ/Elasticsearch 等基础设施。

| 服务 | 目录 | 容器名 | 端口映射 | 技术栈 | 启动命令 |
|:---|:---|:---|:---|:---|:---|
| 后台管理后端 | backend/ | goshopadmin-backend | 8080 | Go + Gin + GORM | go run main.go |
| 后台管理前端 | frontend/ | goshopadmin-frontend | 5173→3000 | Vue3 + Element Plus + Vite | npm run dev |
| C端商城后端 | shop-backend/ | goshopadmin-shop-backend | 8081(仅本机) | Go + Gin + GORM + RabbitMQ | go run main.go |
| C端商城前端 | shop-frontend/ | goshopadmin-shop-frontend | 3001 | Vue3 + Vite | npm run dev |
| 搜索服务 | search-service/ | goshopadmin-search-service | 8082 | Go + Gin + Elasticsearch(v7) | go run main.go |

### 基础设施

| 服务 | 容器名 | 端口 | 用途 |
|:---|:---|:---|:---|
| MySQL 8.0 | goshopadmin-mysql | 3306 | 主数据库（数据库名: goshopadmin） |
| Redis Stack | goshopadmin-redis | 6377→6379 | 缓存 + Bloom 过滤器模块 |
| RabbitMQ | goshopadmin-rabbitmq | 5672 / 15672 | 消息队列（管理界面15672） |
| Elasticsearch | goshopadmin-elasticsearch | 9200 | 搜索引擎（8.17.0, 单节点） |
| Kibana | goshopadmin-kibana | 5601 | ES 可视化界面 |
| Logstash | goshopadmin-logstash | 9600 | 数据同步引擎 |
| Prometheus | goshopadmin-prometheus | 9090 | 监控指标收集 |
| Grafana | goshopadmin-grafana | 3000 | 监控可视化面板 |
| Nginx | goshopadmin-nginx | 80 / 443 | 反向代理（生产环境静态资源） |
| Filebeat | goshopadmin-filebeat | - | 日志收集器 |

## 访问地址速查

### 开发环境

| 系统 | 地址 | 说明 |
|:---|:---|:---|
| 后台管理前端 | http://localhost:3000 (Docker映射5173→3000) | Element Plus 管理界面 |
| 后台管理API | http://localhost:8080/api | Gin RESTful API |
| C端商城前端 | http://localhost:3001 | Vue3 商城界面 |
| C端商城API | http://localhost:8081/api | Gin RESTful API |
| 搜索服务API | http://localhost:8082/api/search | ES 搜索接口 |
| RabbitMQ管理 | http://localhost:15672 | guest/guest 登录 |
| Kibana | http://localhost:5601 | ES 数据可视化 |
| Grafana | http://localhost:3000 | 监控面板（与前端端口冲突时需停frontend） |
| ES REST API | http://localhost:9200 | 直接访问 ES |

### 生产环境

| 系统 | 地址 |
|:---|:---|
| 后台管理 | http://admin.服务器IP (Nginx 80端口) |
| C端商城 | http://shop.服务器IP (Nginx 80端口) |

## 账号与密钥索引

> 完整列表见 `默认登录账号密码.md`，此处为速查摘要。

### 业务系统账号

| 用户名 | 密码 | 系统 | 角色 |
|:---|:---|:---|:---|
| admin | 123456 | 后台管理系统 | 超级管理员（所有权限） |
| platform | 123456 | 后台管理系统 | 平台管理员（所有权限） |
| merchant | 123456 | 后台管理系统 | 商户账号（商品/订单/活动/物流等） |
| pxj | （非默认） | 后台管理系统 | 商户账号 |

C端商城无预置账号，需通过注册接口 `/api/auth/register` 创建。

### 基础设施账号

| 服务 | 用户名 | 密码 | 说明 |
|:---|:---|:---|:---|
| MySQL | root | password | 数据库超级管理员 |
| MySQL | goshopadmin | password | 应用专用账号 |
| Redis | - | （空） | 开发环境无密码 |
| RabbitMQ | guest | guest | 管理界面 http://localhost:15672 |
| Grafana | admin | admin | 监控面板 |
| Elasticsearch | - | 无需认证 | xpack.security.enabled=false |
| Kibana | - | 无需认证 | 同上 |

### JWT Secret（各服务 .env 文件）

| 服务 | .env 路径 | 变量名 | 值前缀 |
|:---|:---|:---|:---|
| backend | backend/.env | JWT_SECRET | pqe9SIY... |
| shop-backend | shop-backend/.env | JWT_SECRET | 1a4tx4p... |
| search-service | search-service/.env | JWT_SECRET_ADMIN | pqe9SIY... (与backend一致) |
| search-service | search-service/.env | JWT_SECRET_CUSTOMER | 1a4tx4p... (与shop-backend一致) |

**安全提醒**：以上均为开发环境默认值，生产环境必须更换。

## 关键配置文件速查

### 后端配置

| 配置项 | 文件路径 | 说明 |
|:---|:---|:---|
| Docker 编排（全部服务） | docker/docker-compose.yml | 13个服务的定义、端口、挂载、依赖 |
| 后台管理Go配置 | backend/config/config.go | LoadConfig() 读取 .env |
| C端商城Go配置 | shop-backend/config/config.go | 同上，DB_HOST默认mysql |
| 搜索服务配置 | search-service/config/config.go | 全局单例模式，含双JWT Secret |
| MQ硬编码配置 | shop-backend/config/mq_config.go | RabbitMQ连接（硬编码Docker服务名） |

### 前端配置

| 配置项 | 文件路径 | 说明 |
|:---|:---|:---|
| 后台管理API代理 | frontend/vite.config.js | /api → backend:8080 |
| 后台管理API定义 | frontend/src/api/auth.js | 所有 API 调用封装 |
| 后台管理路由 | frontend/src/router/index.js | 嵌套路由，Home父路由 |
| C端API代理 | shop-frontend/vite.config.js | /api → shop-backend:8081 |
| C端API定义 | shop-frontend/src/api/index.js | customerAPI/orderAPI/paymentAPI |
| Nginx配置 | docker/nginx/nginx.conf | 反向代理+静态资源 |

### 搜索服务配置

| 配置项 | 文件路径 | 说明 |
|:---|:---|:---|
| ES文档模型 | search-service/models/es_models.go | ProductDoc/OrderDoc/UserDoc/CustomerDoc |
| ES索引模板 | docker/elasticsearch/templates/ | products/orders/users/customers |
| 搜索路由 | search-service/routes/routes.go | admin组(需认证) + customer组(JWT) |
| 数据同步 | search-service/services/sync_service.go | 每60秒从MySQL同步到ES |

### 数据库相关

| 配置项 | 文件路径 | 说明 |
|:---|:---|:---|
| MySQL初始化 | docker/mysql/init.sql | 建表+初始数据（含默认管理员账号） |
| 后端常量 | backend/constants/constants.go | 状态枚举常量 |
| C端常量 | shop-backend/constants/constants.go | 同上 |
| 错误码(后端) | backend/errors/code.go | 4001-5003 错误码体系 |
| 错误码(C端) | shop-backend/errors/code.go | 含额外4043-4094 |
| API文档 | API接口和前端对接情况.md | 接口清单和对接状态 |

### 日志与监控

| 配置项 | 文件路径 | 说明 |
|:---|:---|:---|
| 日志工具(后端) | backend/utils/logger.go | 统一日志（禁止fmt.Println） |
| 日志工具(C端) | shop-backend/utils/logger.go | 同上 |
| 日志工具(搜索) | search-service/utils/logger.go | 同上 |
| Filebeat配置 | docker/elk/filebeat/filebeat.yml | 收集backend/shop-backend日志 |
| Prometheus | docker/prometheus/prometheus.yml | 抓取shop-backend指标 |

## MCP 工具能力清单

当前可用的 MCP 工具（已内置在 Trae IDE 中）：

| 工具名称 | 核心能力 | 典型使用场景 |
|:---|:---|:---|
| Chrome DevTools MCP | 页面导航、DOM快照/截图、控制台日志、网络请求抓取、性能分析、JS执行、元素交互 | UI验证、页面调试、接口测试、样式检查 |
| Playwright MCP | 多浏览器导航、点击/输入/选择、上传、截图、页面断言、文件操作 | E2E自动化测试、表单填写测试 |

MCP 构建方法：参考 `.trae/skills/mcp-builder` 技能（完整的255行方法论指南）

## Docker 操作速查

详细规则见 `.trae/rules/docker_rules.md`。

```powershell
# 重启单个容器（修改代码后）
docker restart goshopadmin-backend
docker restart goshopadmin-shop-backend
docker restart goshopadmin-search-service
docker restart goshopadmin-frontend
docker restart goshopadmin-shop-frontend

# 查看容器日志
docker logs <container-name> --tail 50

# 全量启动/停止
cd docker; docker compose up -d
cd docker; docker compose down

# 端口冲突解决（管理员CMD）
net stop winnat; net start winnat
```

## 项目规范索引

| 规范 | 文件路径 | 核心要求 |
|:---|:---|:---|
| 代码编写规则 | .trae/rules/code_rules.md | 统一响应/路由风格/枚举/命名/错误码/日志 |
| 数据库设计规则 | .trae/rules/database_rules.md | int不用uint/ENUM状态/禁用外键/utf8mb4 |
| Docker操作规则 | .trae/rules/docker_rules.md | 改代码必重启/端口冲突处理 |
| 文档更新规则 | .trae/rules/documentation_rules.md | 改API必改文档/改库必更说明 |
| 前端开发规则 | .trae/rules/frontend_rules.md | Composition API/Vue Router/Axios/Element Plus |
| Git提交规范 | .trae/rules/git-commit-message.md | Conventional Commits中文适配 |
| AI行为规则 | .trae/rules/ai_behavior.md | MCP检测/调用/创建/敏感信息查找 |
