FROM docker.elastic.co/elasticsearch/elasticsearch:8.17.0

# 安装 IK 中文分词插件
# IK 插件已迁移到 InfiniLabs，新下载地址：https://get.infini.cloud/elasticsearch/analysis-ik/{version}
# 旧地址 https://github.com/medcl/elasticsearch-analysis-ik/releases/ 已失效
RUN elasticsearch-plugin install --batch https://get.infini.cloud/elasticsearch/analysis-ik/8.17.0

# 复制自定义 entrypoint 脚本（启动ES后自动执行初始化）
# 使用 --chmod=0755 在 COPY 时直接设置可执行权限（需要 Docker BuildKit）
# ES 官方镜像以 elasticsearch 用户运行，RUN chmod 会因权限不足失败
COPY --chmod=0755 scripts/es-entrypoint.sh /usr/share/elasticsearch/es-entrypoint.sh

# 注意：es-init.sh 和 templates/ 不再 COPY 进镜像
# 改为通过 docker-compose.yml 挂载，修改后无需重新构建镜像
# 挂载点：
#   ./elk/elasticsearch/scripts -> /usr/share/elasticsearch/scripts
#   ./elk/elasticsearch/templates -> /usr/share/elasticsearch/templates

# 覆盖默认 entrypoint，先启动ES再执行初始化
ENTRYPOINT ["/usr/share/elasticsearch/es-entrypoint.sh"]
