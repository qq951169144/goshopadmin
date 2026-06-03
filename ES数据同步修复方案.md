# ES 数据同步修复方案

## 问题总览

| # | 问题 | 影响 | 严重程度 |
|:--|:-----|:-----|:---------|
| 1 | docker-compose.yml 中 search-service 环境变量名 `ES_URL` 与代码读取的 `ES_HOSTS` 不匹配 | search-service 连不上 ES，数据同步完全失效 | 严重 |
| 2 | products.conf SQL 引用不存在的字段（`audit_status`, `main_image`, `images`, `deleted_at`） | 商品数据无法同步到 ES，products 索引不存在 | 严重 |
| 3 | product_skus SQL 引用不存在的 `deleted_at` 字段 | SKU 数据同步失败 | 严重 |
| 4 | backend 没有 godotenv，.env 文件不会被自动读取 | 本地开发不便 | 中等 |
| 5 | ES 索引模板 products.json 包含数据库中不存在的字段映射 | 索引模板与实际数据不一致 | 中等 |
| 6 | search-service 模型 es_models.go 包含数据库中不存在的字段 | 模型与实际数据不一致 | 低 |

---

## 修改1：docker-compose.yml 环境变量名修复

**文件**: `docker/docker-compose.yml` 第429行

```diff
- ES_URL=http://elasticsearch:9200
+ ES_HOSTS=http://elasticsearch:9200
```

**说明**: search-service 代码读取的是 `ES_HOSTS` 环境变量（见 `search-service/config/config.go` 第43行），但 docker-compose.yml 注入的是 `ES_URL`，导致 Docker 环境下 ES 地址回退到默认值 `http://localhost:9200`，无法连接 ES 容器。

---

## 修改2：products.conf SQL 修复

**文件**: `docker/elk/logstash/pipeline/products.conf`

### 2.1 商品基本信息 SQL（第一个 jdbc input）

**当前问题**:
- `p.audit_status` — products 表无此字段
- `p.main_image` — products 表无此字段（图片在 product_images 表）
- `p.images` — products 表无此字段
- `p.deleted_at IS NULL` — products 表无 deleted_at 字段
- `ps.deleted_at IS NULL` — product_skus 表无 deleted_at 字段

**修改后的 SQL**:

```sql
SELECT
  p.id,
  p.name,
  p.description,
  p.detail,
  p.category_id,
  pc.name AS category_name,
  p.merchant_id,
  m.name AS merchant_name,
  p.status,
  p.price,
  IFNULL(MIN(ps.price), p.price) AS min_price,
  IFNULL(MAX(ps.price), p.price) AS max_price,
  IFNULL(SUM(ps.stock), 0) AS total_stock,
  p.is_activity,
  COALESCE(
    (SELECT pi.image_url FROM product_images pi
     WHERE pi.product_id = p.id AND pi.is_main = 1
     LIMIT 1),
    (SELECT pi.image_url FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.sort ASC
     LIMIT 1)
  ) AS main_image,
  p.created_at,
  p.updated_at
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN merchants m ON p.merchant_id = m.id
LEFT JOIN product_skus ps ON p.id = ps.product_id
WHERE p.updated_at > :sql_last_value
GROUP BY p.id
ORDER BY p.updated_at ASC
```

**变更点**:
1. 移除 `p.audit_status`（表中无此字段）
2. 移除 `p.images`（表中无此字段）
3. 移除 `p.main_image`（改为子查询从 product_images 表获取）
4. 添加 `COALESCE` 子查询获取 main_image：优先取 is_main=1 的图片，否则取 sort 最小的图片
5. 移除 `ps.deleted_at IS NULL`（product_skus 表无 deleted_at 字段）
6. 移除 `p.deleted_at IS NULL`（products 表无 deleted_at 字段）

### 2.2 商品 SKU SQL（第二个 jdbc input）

**当前问题**: `ps.deleted_at IS NULL` — product_skus 表无 deleted_at 字段

**修改后的 SQL**:

```sql
SELECT
  ps.id,
  ps.product_id,
  ps.sku_code,
  ps.price,
  ps.original_price,
  ps.stock,
  ps.status,
  ps.is_activity,
  ps.activity_id,
  ps.updated_at
FROM product_skus ps
WHERE ps.updated_at > :sql_last_value
ORDER BY ps.updated_at ASC
```

**变更点**: 移除 `ps.deleted_at IS NULL` 条件

### 2.3 filter 部分无需修改

当前 filter 只处理 `images` 的 `||` 分隔拆分，但既然不需要 images 字段了，移除 images 的 split 处理：

```diff
  if "products" in [tags] {
-   mutate {
-     split => { "images" => "||" }
-   }
-   mutate {
-     remove_field => ["@version"]
-   }
+   mutate {
+     remove_field => ["@version"]
+   }
  }
```

---

## 修改3：ES 索引模板 products.json 修复

**文件**: `docker/elk/elasticsearch/templates/products.json`

移除数据库中不存在的字段映射：

1. 移除 `audit_status` 字段定义（第67-70行）
2. 移除 `images` 字段定义（第95-98行）
3. 保留 `main_image` 字段定义（因为现在通过子查询获取）

---

## 修改4：search-service 模型 es_models.go 修复

**文件**: `search-service/models/es_models.go`

ProductDoc 结构体当前包含 `MainImage` 字段，这个字段现在通过子查询可以获取到了，所以**保留**。

但需要移除 `Sales` 字段（数据库 products 表没有 sales 字段）：

```diff
  // MainImage 主图URL，用于搜索结果展示
  MainImage string `json:"main_image"`

- // Sales 销量，用于排序
- Sales int `json:"sales"`

  // Skus SKU 列表，嵌套文档，包含商品的规格和价格信息
```

---

## 修改5：backend 添加 godotenv 支持

**说明**: backend 是唯一没有 godotenv 的 Go 项目，创建的 .env 文件不会被自动读取。

### 5.1 修改 backend/go.mod

添加依赖：

```
require github.com/joho/godotenv v1.5.1
```

### 5.2 修改 backend/config/config.go

在 `LoadConfig()` 函数开头添加 godotenv 加载：

```diff
+ import "github.com/joho/godotenv"

  func LoadConfig() (*Config, error) {
+   // 加载 .env 文件（如果存在），不会覆盖已存在的系统环境变量
+   _ = godotenv.Load()
+
    // 服务器配置
    serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
```

---

## 修改6：重启容器并重建索引

修复代码后需要执行以下操作：

```powershell
# 1. 删除旧的 products 索引（数据结构已变更）
docker exec goshopadmin-elasticsearch curl -s -X DELETE "http://localhost:9200/products"

# 2. 清除 Logstash 同步位置记录（让商品数据重新全量同步）
docker exec goshopadmin-logstash rm -f /usr/share/logstash/last_run/products_last_run
docker exec goshopadmin-logstash rm -f /usr/share/logstash/last_run/product_skus_last_run

# 3. 重启相关容器
docker restart goshopadmin-logstash
docker restart goshopadmin-search-service
docker restart goshopadmin-backend
```

---

## 修改汇总表

| # | 文件 | 操作 | 说明 |
|:--|:-----|:-----|:-----|
| 1 | `docker/docker-compose.yml` | 修改 | `ES_URL` -> `ES_HOSTS` |
| 2 | `docker/elk/logstash/pipeline/products.conf` | 修改 | 修复 SQL 字段 + 添加 main_image 子查询 + 移除 images split |
| 3 | `docker/elk/elasticsearch/templates/products.json` | 修改 | 移除 audit_status、images 字段映射 |
| 4 | `search-service/models/es_models.go` | 修改 | 移除 Sales 字段 |
| 5 | `backend/go.mod` | 修改 | 添加 godotenv 依赖 |
| 6 | `backend/config/config.go` | 修改 | 添加 godotenv.Load() |

---

## 不需要修改的文件

| 文件 | 原因 |
|:-----|:-----|
| `backend/.env` | 已存在，内容正确 |
| `shop-backend/.env` | 已存在，内容正确 |
| `search-service/.env` | 已存在，内容正确（本地开发用 localhost，Docker 环境由 docker-compose.yml 覆盖） |
| `frontend/.env` | 已存在，内容正确 |
| `shop-frontend/.env` | 已存在，内容正确 |
| `docker/elk/logstash/pipeline/orders.conf` | SQL 字段与数据库匹配，无需修改 |
| `docker/elk/logstash/pipeline/users.conf` | SQL 字段与数据库匹配，无需修改 |
| `docker/elk/logstash/pipeline/customers.conf` | SQL 字段与数据库匹配，无需修改 |
