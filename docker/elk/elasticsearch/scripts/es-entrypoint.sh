#!/bin/bash
# ============================================================================
# Elasticsearch 启动与初始化脚本
# ============================================================================
# 功能：
# 1. 使用官方 entrypoint 启动 Elasticsearch（后台运行）
# 2. 等待 Elasticsearch 就绪
# 3. 执行初始化脚本（创建索引模板、ILM策略等）
# 4. 捕获 SIGTERM 信号并转发给 ES 进程，实现优雅关闭
# 5. 保持 Elasticsearch 前台运行
# ============================================================================

# 以后台方式启动 Elasticsearch（使用官方 entrypoint）
/bin/tini -- /usr/local/bin/docker-entrypoint.sh eswrapper &

ES_PID=$!

# 捕获 SIGTERM/SIGINT 信号并转发给 ES 进程
# Docker 执行 docker stop 时会向 PID 1 发送 SIGTERM
# 如果不转发信号，ES 进程无法优雅关闭，会被 SIGKILL 强杀导致数据损坏
trap "kill -TERM $ES_PID 2>/dev/null" TERM INT

# 等待 Elasticsearch 就绪
echo "=========================================="
echo "等待 Elasticsearch 启动..."
echo "=========================================="
until curl -s http://localhost:9200/_cluster/health | grep -q '"status":"green\|yellow"'; do
  echo "Elasticsearch 尚未就绪，5秒后重试..."
  sleep 5
done
echo "Elasticsearch 已就绪！"

# 执行初始化脚本
if [ -f /usr/share/elasticsearch/scripts/es-init.sh ]; then
  echo "执行 Elasticsearch 初始化脚本..."
  bash /usr/share/elasticsearch/scripts/es-init.sh
fi

echo "初始化完成，Elasticsearch 正在运行..."

# 等待 Elasticsearch 进程（保持容器运行）
# 当 ES 进程退出时，wait 返回其退出码
wait $ES_PID
EXIT_CODE=$?

# 退出时清理，确保容器以 ES 的退出码退出
exit $EXIT_CODE
