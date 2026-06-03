#!/bin/bash
# ============================================================================
# Elasticsearch 初始化脚本
# ============================================================================
# 本脚本在 Elasticsearch 启动后自动执行，完成以下初始化工作：
# 1. 检查 IK 中文分词插件是否安装
# 2. 创建索引生命周期管理（ILM）策略
# 3. 创建业务索引模板（商品、订单、用户、客户、日志等）
# 4. 创建初始日志索引（应用日志、容器日志）
#
# 为什么需要这个脚本？
# - Elasticsearch 启动后是"空"的，需要预先配置好索引模板和策略
# - 索引模板定义了字段类型和分词器，确保数据写入时格式正确
# - ILM 策略自动管理日志的生命周期，避免磁盘被日志占满
# ============================================================================

# Elasticsearch 的连接地址（在 Docker 网络中，ES 的主机名是 elasticsearch）
ES_URL="http://elasticsearch:9200"

# 注意：ES 健康检查已由 es-entrypoint.sh 完成，此处不再重复等待
# es-entrypoint.sh 在调用本脚本前已确认 ES 处于 green/yellow 状态

# ============================================================================
# 第一步：检查 IK 中文分词插件
# ============================================================================
# IK 分词器是中文搜索的核心组件，它能把中文句子拆分成有意义的词语
# 例如："华为手机旗舰款" → "华为"、"手机"、"旗舰"、"款"
# 如果没有 IK 分词器，ES 只能按单个汉字拆分，搜索效果很差
# ============================================================================
echo ""
echo "=========================================="
echo "第一步：检查 IK 中文分词插件"
echo "=========================================="
IK_INSTALLED=$(curl -s "$ES_URL/_cat/plugins" | grep "ik" || true)
if [ -z "$IK_INSTALLED" ]; then
  echo "⚠️  警告：IK 分词插件未安装！中文搜索将无法正常工作。"
  echo "请在 Dockerfile 或启动命令中安装 IK 插件："
  echo "  elasticsearch-plugin install https://get.infini.cloud/elasticsearch/analysis-ik/8.17.0"
else
  echo "✅ IK 分词插件已安装"
  echo "$IK_INSTALLED"
fi

# ============================================================================
# 第二步：创建索引生命周期管理（ILM）策略
# ============================================================================
# ILM（Index Lifecycle Management）是 ES 的自动索引管理功能
# 它可以根据时间或大小自动滚动（rollover）和删除索引
# 为什么需要 ILM？
# - 日志数据量会持续增长，如果不管理会占满磁盘
# - 7天前的日志通常不再需要查看，可以自动删除
# - rollover 机制让索引不会无限增大，保持查询性能
# ============================================================================
echo ""
echo "=========================================="
echo "第二步：创建 ILM 策略"
echo "=========================================="
curl -s -X PUT "$ES_URL/_ilm/policy/app-logs-policy" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/ilm-policy.json
echo ""
echo "✅ ILM 策略创建完成"

# ============================================================================
# 第三步：创建业务索引模板
# ============================================================================
# 索引模板（Index Template）是 ES 的预定义配置
# 当写入数据到某个索引时，ES 会自动匹配对应的模板，应用模板中的配置
# 为什么需要索引模板？
# - 预定义字段类型：确保日期、数字、文本等字段被正确解析
# - 配置分词器：指定哪些字段需要中文分词
# - 设置 nested 类型：确保嵌套对象（如商品的 SKU 列表）能独立查询
# - 绑定 ILM 策略：日志索引自动应用生命周期管理
# ============================================================================

# ---------- 商品索引模板 ----------
# 商品索引是最复杂的模板，包含：
# - IK 分词器：商品名称和描述需要中文搜索
# - nested 类型的 SKU：每个商品有多个 SKU，需要独立查询
# - keyword 子字段：名称既需要分词搜索，也需要精确匹配排序
echo "创建商品索引模板..."
curl -s -X PUT "$ES_URL/_index_template/products" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/products.json
echo ""

# ---------- 订单索引模板 ----------
# 订单索引包含：
# - nested 类型的 items：订单中的商品列表需要独立查询
# - 多个日期字段：支付时间、发货时间等需要精确到秒
echo "创建订单索引模板..."
curl -s -X PUT "$ES_URL/_index_template/orders" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/orders.json
echo ""

# ---------- 用户索引模板（后台管理员） ----------
# 用户索引较简单，主要是 keyword 类型用于精确匹配
# username、email 等字段不需要分词，直接精确匹配即可
echo "创建用户索引模板..."
curl -s -X PUT "$ES_URL/_index_template/users" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/users.json
echo ""

# ---------- 客户索引模板（C端消费者） ----------
# 客户索引的 nickname 字段支持中文分词搜索
# 例如搜索"小明"可以匹配到昵称为"快乐的小明"的客户
echo "创建客户索引模板..."
curl -s -X PUT "$ES_URL/_index_template/customers" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/customers.json
echo ""

# ---------- 应用日志索引模板 ----------
# 应用日志索引绑定 ILM 策略，自动管理生命周期
# 日志字段包括：日志级别、服务名、请求路径、响应时间等
echo "创建应用日志索引模板..."
curl -s -X PUT "$ES_URL/_index_template/app-logs" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/app-logs.json
echo ""

# ---------- 容器日志索引模板 ----------
# 容器日志索引也绑定 ILM 策略
# 记录 Docker 容器的标准输出和标准错误日志
echo "创建容器日志索引模板..."
curl -s -X PUT "$ES_URL/_index_template/container-logs" \
  -H 'Content-Type: application/json' \
  -d @/usr/share/elasticsearch/templates/container-logs.json
echo ""

echo "✅ 所有索引模板创建完成"

# ============================================================================
# 第四步：创建初始日志索引
# ============================================================================
# 日志索引使用"滚动索引"模式，格式为 <别名>-000001
# 当索引满足 ILM 策略的 rollover 条件时，ES 会自动创建下一个索引
# 例如：app-logs-000001 满足条件后，自动创建 app-logs-000002
# 查询时使用别名 app-logs，ES 会自动路由到最新的索引
# ============================================================================
echo ""
echo "=========================================="
echo "第四步：创建初始日志索引"
echo "=========================================="

# 创建应用日志初始索引
# is_write_index: true 表示写入数据时使用这个别名会写入此索引
echo "创建应用日志初始索引..."
curl -s -X PUT "$ES_URL/app-logs-000001" \
  -H 'Content-Type: application/json' \
  -d '{
    "aliases": {
      "app-logs": {
        "is_write_index": true
      }
    }
  }'
echo ""

# 创建容器日志初始索引
echo "创建容器日志初始索引..."
curl -s -X PUT "$ES_URL/container-logs-000001" \
  -H 'Content-Type: application/json' \
  -d '{
    "aliases": {
      "container-logs": {
        "is_write_index": true
      }
    }
  }'
echo ""

echo "✅ 初始日志索引创建完成"

# ============================================================================
# 初始化完成
# ============================================================================
echo ""
echo "=========================================="
echo "🎉 Elasticsearch 初始化全部完成！"
echo "=========================================="
echo "已创建的内容："
echo "  - ILM 策略：app-logs-policy（7天自动清理）"
echo "  - 索引模板：products, orders, users, customers, app-logs, container-logs"
echo "  - 初始索引：app-logs-000001, container-logs-000001"
echo ""
echo "接下来 Logstash 会自动从 MySQL 同步业务数据到 Elasticsearch"
echo "Filebeat 会自动收集应用日志和容器日志"
echo "Kibana 可以通过 http://localhost:5601 访问"
echo "=========================================="
