# ELK 日志与搜索系统 — 设计与开发方案

| 属性 | 值 |
|------|------|
| 项目名称 | GoShopAdmin ELK 日志与搜索系统 |
| 文档版本 | 1.0 |
| 创建日期 | 2026-06-02 |
| 状态 | 待审核 |
| 关联需求文档 | [需求文档](./2026-06-02-elk-logging-search-requirements.md) |

---

## 1. 整体架构

### 1.1 架构图

```
                        ┌─────────────────────────────────────────────┐
                        │           Docker Network                     │
                        │        goshopadmin-network                   │
                        │                                              │
  ┌──────────────┐      │  ┌──────────────┐   ┌──────────────┐        │
  │   frontend   │──────│──│   backend    │   │ shop-backend │        │
  │  (后台管理)   │      │  │  (Go 后端)   │   │ (C端Go后端)   │        │
  └──────────────┘      │  └──────┬───────┘   └──────┬───────┘        │
  ┌──────────────┐      │         │ HTTP REST          │ HTTP REST    │
  │shop-frontend │──────│─────────│────────────────────│───────────   │
  │  (C端商城)    │      │         ▼                    ▼              │
  └──────────────┘      │  ┌──────────────────────────────────────┐   │
                        │  │       search-service (Go)            │   │
                        │  │    统一搜索 API + ES Go Client        │   │
                        │  └──────────────┬───────────────────────┘   │
                        │                 │ ES Go Client               │
                        │                 ▼                            │
                        │  ┌──────────────────────────────────────┐   │
                        │  │       Elasticsearch 8.x              │   │
                        │  │  ┌───────────┐  ┌────────────────┐   │   │
                        │  │  │ 业务索引   │  │  日志索引       │   │   │
                        │  │  │ products  │  │  app-logs-*     │   │   │
                        │  │  │ orders    │  │  container-logs-*│  │   │
                        │  │  │ users     │  │                 │   │   │
                        │  │  │ customers │  │                 │   │   │
                        │  │  └───────────┘  └────────────────┘   │   │
                        │  └──────────▲───────────────────────────┘   │
                        │             │                                │
                        │  ┌──────────┴────────┐  ┌────────────────┐  │
                        │  │     Logstash      │  │   Filebeat     │  │
                        │  │  JDBC 同步业务数据  │  │ 收集应用+容器  │  │
                        │  └──────────┬────────┘  └───────┬────────┘  │
                        │             │                   │           │
                        │             ▼                   ▼           │
                        │  ┌──────────────┐   ┌──────────────┐       │
                        │  │    MySQL      │   │ Docker Logs  │       │
                        │  │  (业务数据源)  │   │ (容器日志)    │       │
                        │  └──────────────┘   └──────────────┘       │
                        │                                              │
                        │  ┌──────────────┐                           │
                        │  │    Kibana     │                           │
                        │  │  可视化仪表盘  │                           │
                        │  └──────────────┘                           │
                        └─────────────────────────────────────────────┘
```

### 1.2 数据流说明

| 数据流 | 方向 | 说明 |
|--------|------|------|
| MySQL → Logstash → ES | 业务数据同步 | Logstash 通过 JDBC 插件定时轮询 MySQL，增量同步到 ES |
| 应用日志文件 → Filebeat → ES | 应用日志收集 | Filebeat 读取 backend/shop-backend 的日志文件 |
| Docker stdout → Filebeat → ES | 容器日志收集 | Filebeat 通过 Docker input 收集所有容器日志 |
| 前端 → 后端 → search-service → ES | 搜索请求 | 前端搜索请求经业务后端转发到 search-service |
| ES → Kibana | 可视化 | Kibana 读取 ES 数据展示仪表盘 |

### 1.3 新增 Docker 容器清单

| 服务 | 容器名 | 镜像 | 端口映射 | 内存限制 | 说明 |
|------|--------|------|---------|---------|------|
| Elasticsearch | goshopadmin-elasticsearch | elasticsearch:8.17.0 | 9200:9200 | 2GB | 搜索引擎 + 数据存储 |
| Logstash | goshopadmin-logstash | logstash:8.17.0 | 无外部端口 | 1GB | 数据同步管道 |
| Kibana | goshopadmin-kibana | kibana:8.17.0 | 5601:5601 | 512MB | 可视化仪表盘 |
| Filebeat | goshopadmin-filebeat | elastic/filebeat:8.17.0 | 无外部端口 | 256MB | 日志收集器 |
| search-service | goshopadmin-search-service | golang:1.24-alpine | 8082:8082 | 256MB | 搜索 API 服务 |

**总新增内存**：约 4GB

### 1.4 端口分配

| 服务 | 端口 | 用途 | 是否冲突 |
|------|------|------|---------|
| Elasticsearch | 9200 | ES REST API | 否（新增） |
| Kibana | 5601 | Kibana Web UI | 否（新增） |
| search-service | 8082 | 搜索 API | 否（新增） |

现有端口：MySQL(3306)、Redis(6379)、RabbitMQ(5672/15672)、backend(8080)、shop-backend(8081)、frontend(5173→3000)、shop-frontend(3001)、Prometheus(9090)、Grafana(3000)、Nginx(80/443)

---

## 2. Elasticsearch 索引设计

### 2.1 业务数据索引

#### 2.1.1 products 索引

商品索引，SKU 作为 nested 对象嵌入商品文档，预存聚合字段用于快速筛选。

```json
{
  "index_patterns": ["products"],
  "template": {
    "mappings": {
      "properties": {
        "id":              { "type": "integer" },
        "name":            { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart", "fields": { "keyword": { "type": "keyword" } } },
        "description":     { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
        "detail":          { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart", "index": false },
        "category_id":     { "type": "integer" },
        "category_name":   { "type": "keyword" },
        "merchant_id":     { "type": "integer" },
        "merchant_name":   { "type": "keyword" },
        "status":          { "type": "keyword" },
        "audit_status":    { "type": "keyword" },
        "price":           { "type": "double" },
        "min_price":       { "type": "double" },
        "max_price":       { "type": "double" },
        "total_stock":     { "type": "integer" },
        "is_activity":     { "type": "integer" },
        "main_image":      { "type": "keyword" },
        "images":          { "type": "keyword" },
        "skus": {
          "type": "nested",
          "properties": {
            "id":            { "type": "integer" },
            "sku_code":      { "type": "keyword" },
            "price":         { "type": "double" },
            "original_price":{ "type": "double" },
            "stock":         { "type": "integer" },
            "status":        { "type": "keyword" },
            "is_activity":   { "type": "integer" },
            "activity_id":   { "type": "integer" },
            "specs": {
              "type": "nested",
              "properties": {
                "spec_name":  { "type": "keyword" },
                "spec_value": { "type": "keyword" }
              }
            }
          }
        },
        "created_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "updated_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" }
      }
    },
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "analysis": {
        "analyzer": {
          "ik_max_word": { "type": "custom", "tokenizer": "ik_max_word" },
          "ik_smart":   { "type": "custom", "tokenizer": "ik_smart" }
        }
      }
    }
  }
}
```

**设计要点**：
- `name` 字段使用 `ik_max_word` 索引分词 + `ik_smart` 搜索分词，并添加 `keyword` 子字段支持精确匹配排序
- `detail` 字段设置 `index: false`，仅存储不建索引（富文本搜索价值低，节省空间）
- `main_image` 使用 `keyword` 类型（完整 URL，无搜索意义）
- `skus` 使用 `nested` 类型保持 SKU 内部字段关联性
- `min_price`/`max_price`/`total_stock` 为聚合字段，避免每次查询做 nested 聚合
- `category_name`/`merchant_name` 为冗余字段，避免搜索时回查 MySQL

#### 2.1.2 orders 索引

```json
{
  "index_patterns": ["orders"],
  "template": {
    "mappings": {
      "properties": {
        "id":              { "type": "integer" },
        "order_no":        { "type": "keyword" },
        "customer_id":     { "type": "integer" },
        "customer_name":   { "type": "keyword" },
        "merchant_id":     { "type": "integer" },
        "merchant_name":   { "type": "keyword" },
        "activity_id":     { "type": "integer" },
        "total_amount":    { "type": "double" },
        "status":          { "type": "keyword" },
        "payment_status":  { "type": "keyword" },
        "shipping_status": { "type": "keyword" },
        "address_id":      { "type": "integer" },
        "payment_method":  { "type": "keyword" },
        "items": {
          "type": "nested",
          "properties": {
            "product_id":    { "type": "integer" },
            "product_name":  { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
            "sku_id":        { "type": "integer" },
            "sku_attributes":{ "type": "keyword" },
            "price":         { "type": "double" },
            "quantity":      { "type": "integer" },
            "total_amount":  { "type": "double" }
          }
        },
        "paid_at":         { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "shipped_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "delivered_at":    { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "cancelled_at":    { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "created_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "updated_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" }
      }
    },
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
    }
  }
}
```

**设计要点**：
- `order_no` 使用 `keyword` 类型，支持精确匹配
- `items` 使用 `nested` 类型，支持按商品名搜索订单
- `customer_name`/`merchant_name` 为冗余字段，方便后台搜索
- `payment_method` 使用 `keyword`，枚举值无需分词

#### 2.1.3 users 索引（后台管理员）

```json
{
  "index_patterns": ["users"],
  "template": {
    "mappings": {
      "properties": {
        "id":              { "type": "integer" },
        "username":        { "type": "keyword" },
        "email":           { "type": "keyword" },
        "role_id":         { "type": "integer" },
        "role_name":       { "type": "keyword" },
        "status":          { "type": "keyword" },
        "created_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "updated_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" }
      }
    },
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
    }
  }
}
```

**设计要点**：
- `username`/`email` 使用 `keyword`，管理员数据量小，精确匹配即可
- `role_name` 冗余字段，方便按角色名搜索
- 不索引 `password` 字段，敏感数据不入 ES

#### 2.1.4 customers 索引（C端客户）

```json
{
  "index_patterns": ["customers"],
  "template": {
    "mappings": {
      "properties": {
        "id":              { "type": "integer" },
        "username":        { "type": "keyword" },
        "phone":           { "type": "keyword" },
        "email":           { "type": "keyword" },
        "nickname":        { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart", "fields": { "keyword": { "type": "keyword" } } },
        "status":          { "type": "keyword" },
        "avatar":          { "type": "keyword", "index": false },
        "last_login_at":   { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "created_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" },
        "updated_at":      { "type": "date", "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd'T'HH:mm:ss.SSSZ||epoch_millis" }
      }
    },
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
    }
  }
}
```

**设计要点**：
- `nickname` 支持中文分词搜索（用户可能搜昵称）
- `phone`/`email` 使用 `keyword`，精确匹配
- `avatar` 设置 `index: false`，仅存储不建索引
- 不索引 `password`/`last_login_ip` 等敏感字段

### 2.2 日志索引

#### 2.2.1 app-logs 索引模板

```json
{
  "index_patterns": ["app-logs-*"],
  "template": {
    "mappings": {
      "properties": {
        "@timestamp":      { "type": "date" },
        "log_level":       { "type": "keyword" },
        "message":         { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
        "service":         { "type": "keyword" },
        "caller":          { "type": "keyword" },
        "request_id":      { "type": "keyword" },
        "method":          { "type": "keyword" },
        "path":            { "type": "keyword" },
        "status_code":     { "type": "integer" },
        "latency_ms":      { "type": "long" },
        "client_ip":       { "type": "keyword" },
        "error_code":      { "type": "integer" },
        "error_message":   { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" }
      }
    },
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "lifecycle.name": "app-logs-policy",
      "lifecycle.rollover_alias": "app-logs"
    }
  }
}
```

#### 2.2.2 container-logs 索引模板

```json
{
  "index_patterns": ["container-logs-*"],
  "template": {
    "mappings": {
      "properties": {
        "@timestamp":      { "type": "date" },
        "container_name":  { "type": "keyword" },
        "message":         { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
        "stream":          { "type": "keyword" },
        "source":          { "type": "keyword" }
      }
    },
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "lifecycle.name": "app-logs-policy",
      "lifecycle.rollover_alias": "container-logs"
    }
  }
}
```

### 2.3 ILM 策略（7 天自动清理）

```json
{
  "policy": {
    "phases": {
      "hot": {
        "min_age": "0ms",
        "actions": {
          "rollover": {
            "max_age": "1d",
            "max_size": "500mb"
          }
        }
      },
      "delete": {
        "min_age": "7d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
```

---

## 3. Logstash 数据同步设计

### 3.1 Pipeline 配置

Logstash 使用多 pipeline 架构，每个业务表独立 pipeline，互不影响。

#### 3.1.1 products 同步 Pipeline

```ruby
# pipeline/products.conf
input {
  jdbc {
    jdbc_driver_library => "/usr/share/logstash/mysql-connector-j.jar"
    jdbc_driver_class => "com.mysql.cj.jdbc.Driver"
    jdbc_connection_string => "jdbc:mysql://mysql:3306/goshopadmin?useSSL=false&characterEncoding=UTF-8"
    jdbc_user => "root"
    jdbc_password => "password"
    schedule => "*/30 * * * * *"   # 每30秒同步一次
    statement => "
      SELECT 
        p.id, p.name, p.description, p.detail,
        p.price, p.stock, p.category_id, p.merchant_id,
        p.status, p.audit_status, p.is_activity,
        p.created_at, p.updated_at,
        pc.name AS category_name,
        m.name AS merchant_name,
        (SELECT pi.image_url FROM product_images pi WHERE pi.product_id = p.id AND pi.is_main = 1 LIMIT 1) AS main_image,
        (SELECT GROUP_CONCAT(pi2.image_url ORDER BY pi2.sort SEPARATOR '||') FROM product_images pi2 WHERE pi2.product_id = p.id) AS images,
        MIN(sku.price) AS min_price,
        MAX(sku.price) AS max_price,
        SUM(sku.stock) AS total_stock
      FROM products p
      LEFT JOIN product_skus sku ON p.id = sku.product_id AND sku.is_activity = 0
      LEFT JOIN product_categories pc ON p.category_id = pc.id
      LEFT JOIN merchants m ON p.merchant_id = m.id
      WHERE p.updated_at > :sql_last_value
      GROUP BY p.id
    "
    use_column_value => true
    tracking_column => "updated_at"
    tracking_column_type => "timestamp"
    last_run_metadata_path => "/usr/share/logstash/last_run/products_last_run"
  }
}

filter {
  ruby {
    code => '
      # 将 images 字符串拆分为数组
      if event.get("images")
        event.set("images", event.get("images").split("||"))
      end
    '
  }
  mutate {
    remove_field => ["@timestamp", "@version"]
  }
}

output {
  elasticsearch {
    hosts => ["http://elasticsearch:9200"]
    index => "products"
    document_id => "%{id}"
    action => "update"
    doc_as_upsert => true
  }
}
```

**SKU 嵌套数据同步**：由于 Logstash JDBC 对 nested 类型支持有限，SKU 明细数据通过 search-service 的定时全量补齐机制处理（见 4.4 节）。

#### 3.1.2 orders 同步 Pipeline

```ruby
# pipeline/orders.conf
input {
  jdbc {
    jdbc_driver_library => "/usr/share/logstash/mysql-connector-j.jar"
    jdbc_driver_class => "com.mysql.cj.jdbc.Driver"
    jdbc_connection_string => "jdbc:mysql://mysql:3306/goshopadmin?useSSL=false&characterEncoding=UTF-8"
    jdbc_user => "root"
    jdbc_password => "password"
    schedule => "*/15 * * * * *"   # 每15秒同步一次
    statement => "
      SELECT 
        o.id, o.order_no, o.customer_id, o.merchant_id,
        o.activity_id, o.total_amount, o.status,
        o.payment_status, o.shipping_status, o.address_id,
        o.payment_method, o.transaction_id,
        o.paid_at, o.shipped_at, o.delivered_at, o.cancelled_at,
        o.created_at, o.updated_at,
        c.username AS customer_name,
        m.name AS merchant_name
      FROM orders o
      LEFT JOIN customers c ON o.customer_id = c.id
      LEFT JOIN merchants m ON o.merchant_id = m.id
      WHERE o.updated_at > :sql_last_value
    "
    use_column_value => true
    tracking_column => "updated_at"
    tracking_column_type => "timestamp"
    last_run_metadata_path => "/usr/share/logstash/last_run/orders_last_run"
  }
}

filter {
  mutate {
    remove_field => ["@timestamp", "@version"]
  }
}

output {
  elasticsearch {
    hosts => ["http://elasticsearch:9200"]
    index => "orders"
    document_id => "%{id}"
    action => "update"
    doc_as_upsert => true
  }
}
```

**订单明细（items）同步**：通过 search-service 的定时补齐机制处理（见 4.4 节）。

#### 3.1.3 users 同步 Pipeline

```ruby
# pipeline/users.conf
input {
  jdbc {
    jdbc_driver_library => "/usr/share/logstash/mysql-connector-j.jar"
    jdbc_driver_class => "com.mysql.cj.jdbc.Driver"
    jdbc_connection_string => "jdbc:mysql://mysql:3306/goshopadmin?useSSL=false&characterEncoding=UTF-8"
    jdbc_user => "root"
    jdbc_password => "password"
    schedule => "*/60 * * * * *"   # 每60秒同步一次
    statement => "
      SELECT 
        u.id, u.username, u.email, u.role_id, u.status,
        u.created_at, u.updated_at,
        r.name AS role_name
      FROM users u
      LEFT JOIN roles r ON u.role_id = r.id
      WHERE u.updated_at > :sql_last_value
    "
    use_column_value => true
    tracking_column => "updated_at"
    tracking_column_type => "timestamp"
    last_run_metadata_path => "/usr/share/logstash/last_run/users_last_run"
  }
}

filter {
  mutate {
    remove_field => ["@timestamp", "@version"]
  }
}

output {
  elasticsearch {
    hosts => ["http://elasticsearch:9200"]
    index => "users"
    document_id => "%{id}"
    action => "update"
    doc_as_upsert => true
  }
}
```

#### 3.1.4 customers 同步 Pipeline

```ruby
# pipeline/customers.conf
input {
  jdbc {
    jdbc_driver_library => "/usr/share/logstash/mysql-connector-j.jar"
    jdbc_driver_class => "com.mysql.cj.jdbc.Driver"
    jdbc_connection_string => "jdbc:mysql://mysql:3306/goshopadmin?useSSL=false&characterEncoding=UTF-8"
    jdbc_user => "root"
    jdbc_password => "password"
    schedule => "*/60 * * * * *"   # 每60秒同步一次
    statement => "
      SELECT 
        id, username, phone, email, nickname,
        status, avatar, last_login_at, created_at, updated_at
      FROM customers
      WHERE updated_at > :sql_last_value
    "
    use_column_value => true
    tracking_column => "updated_at"
    tracking_column_type => "timestamp"
    last_run_metadata_path => "/usr/share/logstash/last_run/customers_last_run"
  }
}

filter {
  mutate {
    remove_field => ["@timestamp", "@version"]
  }
}

output {
  elasticsearch {
    hosts => ["http://elasticsearch:9200"]
    index => "customers"
    document_id => "%{id}"
    action => "update"
    doc_as_upsert => true
  }
}
```

### 3.2 同步策略汇总

| 同步任务 | 轮询间隔 | 同步方式 | 说明 |
|---------|---------|---------|------|
| products | 30s | 增量（updated_at） | 含 SKU 聚合字段，SKU 明细由 search-service 补齐 |
| orders | 15s | 增量（updated_at） | 订单变更频繁，间隔短；items 由 search-service 补齐 |
| users | 60s | 增量（updated_at） | 管理员变更频率低 |
| customers | 60s | 增量（updated_at） | 客户变更频率低 |

---

## 4. search-service 设计

### 4.1 项目结构

```
search-service/
├── config/
│   └── config.go           # 配置加载
├── controllers/
│   ├── base_controller.go  # 基础控制器（统一响应）
│   ├── product_controller.go  # 商品搜索
│   ├── order_controller.go    # 订单搜索
│   ├── user_controller.go     # 用户搜索
│   ├── customer_controller.go # 客户搜索
│   └── health_controller.go   # 健康检查（ES 连通性 + IK 插件 + 数据新鲜度）
├── errors/
│   └── code.go             # 错误码定义
├── middleware/
│   ├── cors.go             # CORS 中间件
│   ├── logger.go           # 请求日志中间件
│   └── rate_limit.go       # 请求限流中间件（防止 ES 过载）
├── models/
│   └── es_models.go        # ES 文档模型定义
├── routes/
│   └── routes.go           # 路由配置
├── services/
│   ├── es_client.go        # ES 客户端管理（含连接池 + 熔断器）
│   ├── product_service.go  # 商品搜索服务
│   ├── order_service.go    # 订单搜索服务
│   ├── user_service.go     # 用户搜索服务
│   ├── customer_service.go # 客户搜索服务
│   └── sync_service.go     # 数据补齐服务（SKU/OrderItems）
├── utils/
│   └── logger.go           # 日志工具
├── .env.example
├── go.mod
├── go.sum
└── main.go
```

### 4.2 API 设计

#### 4.2.1 商品搜索

```
GET /api/search/products

请求参数（Query String）：
  keyword     - 搜索关键词（商品名/描述）
  category_id - 分类ID筛选
  merchant_id - 商户ID筛选
  status      - 状态筛选（active/inactive）
  min_price   - 最低价格
  max_price   - 最高价格
  sort        - 排序字段（price/created_at/total_stock）
  order       - 排序方向（asc/desc）
  page        - 页码（默认1）
  page_size   - 每页数量（默认20）

响应：
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "name": "商品名称",
        "description": "商品描述",
        "category_id": 1,
        "category_name": "分类名",
        "merchant_id": 1,
        "merchant_name": "商户名",
        "status": "active",
        "price": 99.00,
        "min_price": 89.00,
        "max_price": 129.00,
        "total_stock": 500,
        "main_image": "http://xxx/1.jpg",
        "skus": [...],
        "created_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### 4.2.2 订单搜索

```
GET /api/search/orders

请求参数（Query String）：
  keyword         - 搜索关键词（订单号/商品名）
  customer_id     - 客户ID筛选
  merchant_id     - 商户ID筛选
  status          - 订单状态筛选
  payment_status  - 支付状态筛选
  shipping_status - 物流状态筛选
  start_date      - 开始日期
  end_date        - 结束日期
  min_amount      - 最小金额
  max_amount      - 最大金额
  sort            - 排序字段（created_at/total_amount）
  order           - 排序方向（asc/desc）
  page            - 页码
  page_size       - 每页数量

响应：
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "order_no": "ORD20260101001",
        "customer_id": 1,
        "customer_name": "客户名",
        "status": "paid",
        "total_amount": 299.00,
        "items": [...],
        "created_at": "2026-01-01T10:00:00Z"
      }
    ]
  }
}
```

#### 4.2.3 用户搜索（后台管理员）

```
GET /api/search/users

请求参数（Query String）：
  keyword  - 搜索关键词（用户名/邮箱）
  role_id  - 角色ID筛选
  status   - 状态筛选
  page     - 页码
  page_size - 每页数量

响应：
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 10,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "role_name": "超级管理员",
        "status": "active",
        "created_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### 4.2.4 客户搜索（C端）

```
GET /api/search/customers

请求参数（Query String）：
  keyword  - 搜索关键词（用户名/手机号/邮箱/昵称）
  status   - 状态筛选
  page     - 页码
  page_size - 每页数量

响应：
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 30,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "username": "customer001",
        "phone": "13800138000",
        "email": "c@example.com",
        "nickname": "小明",
        "status": "active",
        "created_at": "2026-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### 4.2.5 搜索建议

```
GET /api/search/suggest

请求参数（Query String）：
  prefix - 搜索前缀
  type   - 搜索类型（product/order/user/customer）

响应：
{
  "code": 0,
  "message": "success",
  "data": {
    "suggestions": ["手机壳", "手机膜", "手机充电器"]
  }
}
```

### 4.3 健康检查 API

search-service 提供多维度健康检查，供业务后端判断搜索服务可用性，实现降级决策。

```
GET /health

响应：
{
  "status": "healthy",              // healthy / degraded / unhealthy
  "elasticsearch": {
    "connected": true,              // ES 连接状态
    "cluster_status": "green",      // green / yellow / red
    "ik_plugin_installed": true     // IK 分词器是否已安装
  },
  "data_freshness": {
    "products_last_sync": "2026-06-02T15:03:30Z",  // products 最后同步时间
    "orders_last_sync": "2026-06-02T15:03:15Z",     // orders 最后同步时间
    "max_delay_seconds": 30                          // 最大允许延迟（秒）
  }
}
```

**健康状态判定逻辑**：

| 条件 | 状态 | 业务后端行为 |
|------|------|------------|
| ES 连接正常 + IK 已安装 + 同步延迟 < 30s | `healthy` | 正常使用 ES 搜索 |
| ES 连接正常但 IK 未安装，或同步延迟 > 30s | `degraded` | 使用 ES 搜索但提示功能受限 |
| ES 连接失败 | `unhealthy` | 降级到 MySQL 查询 |

### 4.4 ES 查询策略

#### 商品搜索查询构建

```go
// 多字段匹配 + 条件过滤 + 高亮
query := elastic.NewBoolQuery()
if keyword != "" {
    query.Must(elastic.NewMultiMatchQuery(keyword, "name", "description").
        Analyzer("ik_smart").
        Type("best_fields"))
}
if categoryID > 0 {
    query.Filter(elastic.NewTermQuery("category_id", categoryID))
}
if minPrice > 0 || maxPrice > 0 {
    rangeQ := elastic.NewRangeQuery("min_price")
    if minPrice > 0 { rangeQ.Gte(minPrice) }
    if maxPrice > 0 { rangeQ.Lte(maxPrice) }
    query.Filter(rangeQ)
}

highlight := elastic.NewHighlight().
    Field("name").
    Field("description").
    PreTags("<em>").
    PostTags("</em>")
```

### 4.5 ES 客户端熔断与限流

#### 4.5.1 熔断器（Circuit Breaker）

search-service 的 ES 客户端内置熔断器，防止 ES 故障时请求堆积导致级联崩溃。

```go
// es_client.go 熔断器配置
type CircuitBreakerConfig struct {
    MaxFailures    int           // 连续失败阈值（默认 5 次）
    ResetTimeout   time.Duration // 熔断恢复等待时间（默认 30s）
    HalfOpenMax    int           // 半开状态最大试探请求（默认 1 次）
}

// 熔断器状态机
// Closed（正常）→ 连续失败达 MaxFailures → Open（熔断，直接返回错误）
// Open → 等待 ResetTimeout → HalfOpen（试探）→ 成功则 Closed，失败则 Open
```

#### 4.5.2 请求限流

```go
// rate_limit.go 限流配置
type RateLimitConfig struct {
    MaxRequestsPerSecond float64 // 每秒最大请求数（默认 50）
    MaxConcurrentSearches int    // 最大并发搜索数（默认 20）
    SearchTimeout        time.Duration // 单次搜索超时（默认 3s）
}
```

**限流策略**：
- 超过 QPS 限制的请求返回 429 Too Many Requests
- 超过并发限制的请求排队等待，等待超时返回 503
- 搜索超时自动取消，返回部分结果或错误提示

### 4.6 数据补齐服务（sync_service）

Logstash JDBC 同步只能处理扁平数据，无法直接组装 nested 对象（SKU 明细、订单明细）。search-service 内置定时补齐服务，周期性从 MySQL 读取关联数据并更新到 ES。

**补齐逻辑**：

```
每 60 秒执行一次：
1. 查询 MySQL product_skus + product_sku_specs，按 product_id 分组
2. 批量更新 ES products 文档的 skus 字段
3. 查询 MySQL order_items，按 order_id 分组
4. 批量更新 ES orders 文档的 items 字段
```

**为什么不用 Logstash aggregate 插件**：
- aggregate 插件配置复杂，调试困难
- 需要维护 Logstash Ruby 代码，增加运维成本
- search-service 用 Go 实现，类型安全、可测试、可维护

### 4.7 与业务后端的集成方式

业务后端（backend/shop-backend）通过 HTTP REST 调用 search-service：

```
前端 → backend/shop-backend → search-service → Elasticsearch
```

#### 4.7.1 降级策略（Fallback to MySQL）

业务后端内置搜索降级机制，当 search-service 不可用时自动回退到 MySQL 查询。

```go
// backend/services/search_proxy_service.go
type SearchProxyService struct {
    searchServiceURL string
    httpClient       *http.Client
    circuitBreaker   *CircuitBreaker
    db               *gorm.DB
}

func (s *SearchProxyService) SearchProducts(ctx *gin.Context, params SearchParams) {
    if s.circuitBreaker.IsOpen() {
        // 熔断状态：直接走 MySQL
        s.searchFromMySQL(ctx, params)
        return
    }

    resp, err := s.callSearchService(params)
    if err != nil {
        // 调用失败：记录失败次数，降级到 MySQL
        s.circuitBreaker.RecordFailure()
        utils.Warn("search-service 调用失败，降级到 MySQL: %v", err)
        s.searchFromMySQL(ctx, params)
        return
    }

    s.circuitBreaker.RecordSuccess()
    // 转发 search-service 响应
    s.forwardResponse(ctx, resp)
}
```

**降级触发条件**：

| 条件 | 触发降级 | 恢复条件 |
|------|---------|---------|
| search-service 连续 3 次调用超时（>2s） | 是 | 熔断器等待 30s 后试探 |
| search-service 返回 5xx 错误 | 是 | 熔断器等待 30s 后试探 |
| search-service /health 返回 `unhealthy` | 是 | 下次健康检查通过 |
| search-service 容器停止 | 是 | 容器重启后健康检查通过 |

**MySQL 降级查询限制**：
- 降级查询仅支持基本关键词搜索（LIKE），不支持中文分词
- 降级查询增加 3s 超时限制，防止 MySQL 过载
- 降级时前端显示提示："搜索服务暂时不可用，结果可能不完整"

#### 4.7.2 调用示例（backend 代理搜索请求）

```go
// backend/controllers/search_controller.go
func (c *SearchController) SearchProducts(ctx *gin.Context) {
    params := SearchParams{
        Keyword:    ctx.Query("keyword"),
        CategoryID: ctx.Query("category_id"),
        // ...
    }
    c.searchProxyService.SearchProducts(ctx, params)
}
```

**为什么不让前端直接调用 search-service**：
- 统一认证鉴权：前端请求经过业务后端的 Auth 中间件
- 避免暴露内部服务：search-service 不对外暴露端口（仅 Docker 内网可访问）
- 降级保护：业务后端可以在 search-service 不可用时回退到 MySQL
- 请求聚合：业务后端可以组合搜索结果和业务逻辑

---

## 5. Filebeat 日志收集设计

### 5.1 应用日志收集

现有 `utils/logger.go` 的日志输出格式：

```
2026/06/02 15:04:05 [INFO] [logger.go:25] API请求: {"request_id":"xxx","method":"GET",...}
```

**改造方案**：在日志输出中增加 `service` 字段，使 Filebeat 能区分日志来源。

```go
// 改造后的日志格式（结构化 JSON 行）
{"timestamp":"2026-06-02T15:04:05Z","level":"INFO","service":"backend","caller":"logger.go:25","message":"API请求","request_id":"xxx","method":"GET",...}
```

**改造内容**：
- `backend/utils/logger.go`：增加 `service` 字段，输出纯 JSON 行格式
- `shop-backend/utils/logger.go`：同上，`service` 值为 `shop-backend`
- 保持 `utils.Info()`/`utils.Warn()`/`utils.Error()` 接口不变，仅修改内部格式化逻辑

### 5.2 Filebeat 配置

```yaml
# filebeat.yml
filebeat.config:
  modules:
    path: ${path.config}/modules.d/*.yml

filebeat.inputs:
  # 应用日志收集
  - type: log
    enabled: true
    paths:
      - /logs/backend/*.log
    fields:
      service: backend
      log_type: app
    fields_under_root: true
    json.keys_under_root: true
    json.add_error_key: true

  - type: log
    enabled: true
    paths:
      - /logs/shop-backend/*.log
    fields:
      service: shop-backend
      log_type: app
    fields_under_root: true
    json.keys_under_root: true
    json.add_error_key: true

  # Docker 容器日志收集
  - type: container
    enabled: true
    paths:
      - /var/lib/docker/containers/*/*.log
    processors:
      - add_docker_metadata:
          host: "unix:///var/run/docker.sock"
      - decode_json_fields:
          fields: ["message"]
          target: "container"
          overwrite_keys: true

processors:
  - add_host_metadata: {}
  - add_cloud_metadata: {}
  - add_docker_metadata: {}

output.elasticsearch:
  hosts: ["http://elasticsearch:9200"]
  indices:
    - index: "app-logs-%{+yyyy.MM.dd}"
      when.equals:
        log_type: "app"
    - index: "container-logs-%{+yyyy.MM.dd}"
      when.equals:
        log_type: "container"

logging.level: info
logging.to_files: true
logging.files:
  path: /var/log/filebeat
  name: filebeat
  keepfiles: 7
```

### 5.3 Docker 日志挂载

在 docker-compose.yml 中为 backend 和 shop-backend 添加日志卷挂载：

```yaml
backend:
  volumes:
    - ../backend:/app
    - backend-logs:/app/logs    # 新增：日志卷

shop-backend:
  volumes:
    - ../shop-backend:/app
    - shop-backend-logs:/app/logs  # 新增：日志卷
```

Filebeat 挂载这些日志卷：

```yaml
filebeat:
  volumes:
    - backend-logs:/logs/backend:ro
    - shop-backend-logs:/logs/shop-backend:ro
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

---

## 6. Kibana 仪表盘设计

### 6.1 仪表盘清单

| 仪表盘名称 | 索引模式 | 核心可视化 |
|-----------|---------|-----------|
| 订单统计 | orders | 订单趋势线图、状态饼图、金额柱状图、每日订单量 |
| 商品分析 | products | 商品状态分布、分类分布、价格区间分布 |
| 错误分析 | app-logs-* | 错误趋势线图、错误级别饼图、慢请求 Top10 |
| 容器监控 | container-logs-* | 各容器日志量、错误率对比、容器状态 |
| 搜索分析 | app-logs-* | 搜索关键词 Top10、搜索无结果率、搜索响应时间分布 |

### 6.2 Kibana 配置

```yaml
# kibana.yml
server.name: koshopadmin-kibana
server.host: "0.0.0.0"
elasticsearch.hosts: ["http://elasticsearch:9200"]
i18n.locale: "zh-CN"
```

---

## 7. Docker Compose 配置

### 7.1 新增服务定义

```yaml
# Elasticsearch
elasticsearch:
  image: elasticsearch:8.17.0
  container_name: goshopadmin-elasticsearch
  ports:
    - "9200:9200"
  environment:
    - discovery.type=single-node
    - xpack.security.enabled=false
    - xpack.security.http.ssl.enabled=false
    - ES_JAVA_OPTS=-Xms1g -Xmx1g
    - cluster.name=goshopadmin-es
    - bootstrap.memory_lock=true
  ulimits:
    memlock:
      soft: -1
      hard: -1
    nofile:
      soft: 65536
      hard: 65536
  volumes:
    - es-data:/usr/share/elasticsearch/data
    - ./elk/elasticsearch/plugins:/usr/share/elasticsearch/plugins
    - ./elk/elasticsearch/scripts/es-init.sh:/usr/share/elasticsearch/es-init.sh
  networks:
    - goshopadmin-network
  restart: always
  healthcheck:
    test: ["CMD-SHELL", "curl -sf http://localhost:9200/_cluster/health || exit 1"]
    interval: 30s
    timeout: 10s
    retries: 5
    start_period: 60s

# Logstash
logstash:
  image: logstash:8.17.0
  container_name: goshopadmin-logstash
  environment:
    - LS_JAVA_OPTS=-Xms512m -Xmx512m
  volumes:
    - ./elk/logstash/pipeline:/usr/share/logstash/pipeline
    - ./elk/logstash/config/logstash.yml:/usr/share/logstash/config/logstash.yml
    - ./elk/logstash/mysql-connector-j.jar:/usr/share/logstash/mysql-connector-j.jar
    - logstash-last-run:/usr/share/logstash/last_run
  networks:
    - goshopadmin-network
  restart: always
  depends_on:
    elasticsearch:
      condition: service_healthy
    mysql:
      condition: service_healthy

# Kibana
kibana:
  image: kibana:8.17.0
  container_name: goshopadmin-kibana
  ports:
    - "5601:5601"
  environment:
    - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    - I18N_LOCALE=zh-CN
  volumes:
    - ./elk/kibana/kibana.yml:/usr/share/kibana/config/kibana.yml
  networks:
    - goshopadmin-network
  restart: always
  depends_on:
    elasticsearch:
      condition: service_healthy

# Filebeat
filebeat:
  image: elastic/filebeat:8.17.0
  container_name: goshopadmin-filebeat
  user: root
  volumes:
    - ./elk/filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
    - backend-logs:/logs/backend:ro
    - shop-backend-logs:/logs/shop-backend:ro
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
    - /var/run/docker.sock:/var/run/docker.sock:ro
  networks:
    - goshopadmin-network
  restart: always
  depends_on:
    elasticsearch:
      condition: service_healthy
  healthcheck:
    test: ["CMD-SHELL", "filebeat test config -c /usr/share/filebeat/filebeat.yml -e || exit 1"]
    interval: 60s
    timeout: 10s
    retries: 3

# search-service
search-service:
  image: golang:1.24-alpine
  container_name: goshopadmin-search-service
  ports:
    - "8082:8082"
  volumes:
    - ../search-service:/app
  working_dir: /app
  environment:
    - SERVER_PORT=8082
    - ES_HOSTS=http://elasticsearch:9200
    - DB_HOST=mysql
    - DB_PORT=3306
    - DB_USER=root
    - DB_PASSWORD=password
    - DB_NAME=goshopadmin
    - GOPROXY=https://goproxy.cn,direct
  networks:
    - goshopadmin-network
  restart: always
  depends_on:
    elasticsearch:
      condition: service_healthy
    mysql:
      condition: service_healthy
  command: sh -c "go mod download && go run main.go"
  healthcheck:
    test: ["CMD", "wget", "--spider", "-q", "http://localhost:8082/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 30s
```

### 7.2 新增数据卷

```yaml
volumes:
  es-data:
    driver: local
    driver_opts:
      type: none
      device: ./data/elasticsearch
      o: bind
  logstash-last-run:
    driver: local
    driver_opts:
      type: none
      device: ./data/logstash/last_run
      o: bind
  backend-logs:
    driver: local
    driver_opts:
      type: none
      device: ../backend/logs
      o: bind
  shop-backend-logs:
    driver: local
    driver_opts:
      type: none
      device: ../shop-backend/logs
      o: bind
```

### 7.3 现有服务修改

为 backend 和 shop-backend 添加日志卷：

```yaml
backend:
  volumes:
    - ../backend:/app
    - backend-logs:/app/logs      # 新增

shop-backend:
  volumes:
    - ../shop-backend:/app
    - shop-backend-logs:/app/logs  # 新增
```

---

## 8. 前端集成设计

### 8.1 后台管理系统（frontend）

#### 8.1.1 新增路由

```javascript
// router/index.js
{
  path: '/search',
  name: 'Search',
  component: () => import('../views/search/Search.vue'),
  meta: { requiresAuth: true }
}
```

#### 8.1.2 新增页面

| 页面 | 路径 | 功能 |
|------|------|------|
| 综合搜索 | `views/search/Search.vue` | 商品/订单/用户/客户统一搜索入口 |
| 搜索结果 | `views/search/SearchResults.vue` | 搜索结果展示（Tab 切换不同类型） |

#### 8.1.3 API 调用

```javascript
// api/search.js
import api from './auth'

export const searchAPI = {
  searchProducts: (params) => api.get('/search/products', { params }),
  searchOrders: (params) => api.get('/search/orders', { params }),
  searchUsers: (params) => api.get('/search/users', { params }),
  searchCustomers: (params) => api.get('/search/customers', { params }),
  suggest: (params) => api.get('/search/suggest', { params })
}
```

注意：前端调用的是业务后端的代理接口（`/api/search/*`），由业务后端转发到 search-service。

### 8.2 C端商城（shop-frontend）

#### 8.2.1 新增功能

| 功能 | 位置 | 说明 |
|------|------|------|
| 商品搜索框 | 顶部导航栏 | 输入关键词搜索商品 |
| 搜索结果页 | 独立页面 | 商品列表 + 筛选条件 |
| 订单搜索 | 订单列表页 | 按订单号/商品名搜索 |

#### 8.2.2 API 调用

```javascript
// api/search.js
export const searchAPI = {
  searchProducts: (params) => api.get('/search/products', { params }),
  searchOrders: (params) => api.get('/search/orders', { params }),
  suggest: (params) => api.get('/search/suggest', { params })
}
```

---

## 9. 配置文件清单

### 9.1 新增配置文件

| 文件路径 | 说明 |
|---------|------|
| `docker/elk/elasticsearch/plugins/ik/` | IK 分词器插件目录（预下载） |
| `docker/elk/elasticsearch/scripts/es-init.sh` | ES 初始化脚本（检查 IK 插件、创建索引模板和 ILM 策略） |
| `docker/elk/logstash/pipeline/products.conf` | 商品同步 Pipeline |
| `docker/elk/logstash/pipeline/orders.conf` | 订单同步 Pipeline |
| `docker/elk/logstash/pipeline/users.conf` | 用户同步 Pipeline |
| `docker/elk/logstash/pipeline/customers.conf` | 客户同步 Pipeline |
| `docker/elk/logstash/config/logstash.yml` | Logstash 主配置 |
| `docker/elk/logstash/mysql-connector-j.jar` | MySQL JDBC 驱动 |
| `docker/elk/kibana/kibana.yml` | Kibana 配置 |
| `docker/elk/filebeat/filebeat.yml` | Filebeat 配置 |
| `docker/data/elasticsearch/` | ES 数据目录 |
| `docker/data/logstash/last_run/` | Logstash 同步位点目录 |
| `search-service/` | 搜索服务代码目录 |

---

## 10. 实施步骤

| 步骤 | 任务 | 依赖 | 产出 |
|------|------|------|------|
| 1 | Docker Compose 添加 ELK 服务 + IK 插件安装 | 无 | ES/Kibana/Logstash/Filebeat 容器可启动 |
| 2 | 创建 ES 索引模板和 ILM 策略 | 步骤1 | 4 个业务索引 + 2 个日志索引模板 |
| 3 | 配置 Logstash Pipeline（4 个同步任务） | 步骤2 | MySQL 数据自动同步到 ES |
| 4 | 开发 search-service（Go 项目） | 步骤2 | 搜索 API 服务可运行 |
| 5 | 配置 Filebeat 日志收集 | 步骤1 | 应用日志和容器日志自动收集 |
| 6 | 改造现有 logger.go（增加 service 字段，JSON 行格式） | 步骤5 | 日志格式兼容 Filebeat 解析 |
| 7 | Kibana 仪表盘配置 | 步骤3,5 | 订单统计/错误分析仪表盘 |
| 8 | 业务后端添加搜索代理接口 | 步骤4 | backend/shop-backend 转发搜索请求 |
| 9 | 后台管理前端集成搜索页面 | 步骤8 | frontend 搜索功能可用 |
| 10 | C端商城前端集成搜索功能 | 步骤8 | shop-frontend 搜索功能可用 |
| 11 | 端到端测试和性能优化 | 步骤9,10 | 全链路验证通过 |

---

## 11. 风险与应对

### 风险 1：ES 内存不足导致 OOM

| 维度 | 内容 |
|------|------|
| **影响** | 搜索服务不可用，严重时导致容器被 OOM Kill |
| **根因** | JVM 堆内存设置不合理、查询结果集过大、索引数据量增长超预期 |

**应对设计**：

1. **JVM 堆内存硬限制**：ES_JAVA_OPTS 设置 `-Xms1g -Xmx1g`，且 Docker memory limit 设为 2GB（堆内存的 2 倍，留给 OS cache 和 Lucene off-heap）
2. **search-service 请求限流**（见 4.5.2 节）：限制 QPS 50、并发搜索 20、超时 3s，防止突发流量打爆 ES
3. **ES 查询优化**：所有搜索 API 强制分页（page_size 上限 50），禁止 `match_all` 无条件查询，`detail` 字段设置 `index: false` 减少内存占用
4. **ILM 自动清理**（见 2.3 节）：日志索引 7 天自动删除，每天 rollover 限制单索引 500MB
5. **监控告警**：Kibana 仪表盘增加 ES 节点内存使用率监控，超过 85% 阈值告警

### 风险 2：Logstash JDBC 同步延迟

| 维度 | 内容 |
|------|------|
| **影响** | 搜索结果与数据库不一致，用户看到过期数据 |
| **根因** | Logstash 轮询间隔、SQL 查询慢、MySQL 负载高 |

**应对设计**：

1. **数据新鲜度健康检查**（见 4.3 节）：search-service `/health` 接口暴露每个索引的最后同步时间，业务后端据此判断是否降级
2. **分级同步频率**：orders 15s、products 30s、users/customers 60s，按业务变更频率分配资源
3. **业务后端降级策略**（见 4.7.1 节）：当同步延迟超过 30s 时，健康状态变为 `degraded`，前端提示"数据可能有延迟"
4. **关键查询走 MySQL**：订单详情、支付状态等强一致性场景，始终直接查询 MySQL，不依赖 ES
5. **Logstash last_run 持久化**：Logstash 同步位点写入 Docker volume，容器重启后从上次位置继续，避免全量重同步

### 风险 3：IK 分词器未安装

| 维度 | 内容 |
|------|------|
| **影响** | 中文搜索不可用，搜索结果质量极差 |
| **根因** | IK 插件未预装、ES 升级后插件不兼容、插件目录挂载失败 |

**应对设计**：

1. **预下载 IK 插件到项目目录**：`docker/elk/elasticsearch/plugins/ik/` 目录预置 IK 插件文件，Docker Compose 挂载到 ES 容器的 plugins 目录，避免运行时下载
2. **ES 初始化脚本**（`es-init.sh`）：ES 启动后自动执行以下检查：
   ```bash
   # 检查 IK 插件是否已安装
   curl -sf http://localhost:9200/_cat/plugins | grep ik || echo "WARNING: IK plugin not found"
   # 创建索引模板和 ILM 策略
   curl -X PUT http://localhost:9200/_index_template/products -d @/templates/products.json
   curl -X PUT http://localhost:9200/_ilm/policy/app-logs-policy -d @/templates/ilm.json
   ```
3. **search-service 健康检查**（见 4.3 节）：`/health` 接口检测 `ik_plugin_installed` 字段，IK 未安装时状态变为 `degraded`
4. **降级策略**：IK 未安装时，search-service 自动将中文分词查询降级为标准分词器（standard analyzer），搜索质量降低但不会完全失败

### 风险 4：Filebeat 权限不足无法读取 Docker 日志

| 维度 | 内容 |
|------|------|
| **影响** | 容器日志缺失，无法排查线上问题 |
| **根因** | Windows Docker 环境下 Docker socket 权限、容器日志路径差异 |

**应对设计**：

1. **Filebeat 以 root 用户运行**：Docker Compose 中设置 `user: root`
2. **Docker socket 挂载**：`/var/run/docker.sock:/var/run/docker.sock:ro`
3. **Filebeat 健康检查**：Docker Compose 中配置 Filebeat 配置文件验证 healthcheck
4. **应用日志双保险**：即使 Filebeat 无法收集容器日志，应用日志文件（`/app/logs/`）通过 volume 挂载仍可被 Filebeat 收集
5. **Windows 环境适配**：Windows Docker Desktop 下容器日志路径为 `C:\ProgramData\Docker\containers\`，Filebeat 配置中需根据操作系统调整路径（在 docker-compose.yml 中通过环境变量区分）

### 风险 5：search-service 不可用时搜索功能全部失效

| 维度 | 内容 |
|------|------|
| **影响** | 前端搜索功能完全不可用，用户体验极差 |
| **根因** | search-service 进程崩溃、ES 连接失败、内存溢出 |

**应对设计**：

1. **业务后端熔断器**（见 4.7.1 节）：backend/shop-backend 内置 CircuitBreaker，连续 3 次调用失败后自动熔断，30s 后试探恢复
2. **自动降级到 MySQL**（见 4.7.1 节）：熔断状态下搜索请求自动回退到 MySQL LIKE 查询，前端提示"搜索服务暂时不可用，结果可能不完整"
3. **search-service 自动重启**：Docker Compose 设置 `restart: always`，进程崩溃后自动重启
4. **search-service 健康检查**：Docker Compose 配置 `/health` 端点健康检查，start_period 30s 给予启动缓冲
5. **ES 客户端熔断器**（见 4.5.1 节）：search-service 内部对 ES 连接也做熔断保护，ES 短暂不可用时快速失败而非阻塞等待

---

*文档版本: 1.0 | 最后更新: 2026-06-02*
