# Nginx 高并发优化完整方案

## 一、负载均衡方案

### 1. 架构设计

通过 Nginx 作为反向代理，将请求分发到多个 Go 后端服务实例，实现负载均衡。

### 2. 具体配置

#### 2.1 修改 docker-compose.yml

```yaml
# 后端服务（多个实例）
backend1:
  image: golang:1.24-alpine
  container_name: goshopadmin-backend-1
  # 不暴露端口到主机，只在内部网络通信
  volumes:
    - ../backend:/app
  working_dir: /app
  environment:
    - SERVER_PORT=8080
    - DB_HOST=mysql
    - DB_PORT=3306
    - DB_USER=root
    - DB_PASSWORD=password
    - DB_NAME=goshopadmin
    - REDIS_HOST=redis
    - REDIS_PORT=6379
    - JWT_SECRET=your-secret-key
  networks:
    - goshopadmin-network
  restart: always
  depends_on:
    - mysql
    - redis
  command: sh -c "go mod download && go run main.go"

backend2:
  image: golang:1.24-alpine
  container_name: goshopadmin-backend-2
  # 与 backend1 配置相同
  volumes:
    - ../backend:/app
  working_dir: /app
  environment:
    - SERVER_PORT=8080
    - DB_HOST=mysql
    - DB_PORT=3306
    - DB_USER=root
    - DB_PASSWORD=password
    - DB_NAME=goshopadmin
    - REDIS_HOST=redis
    - REDIS_PORT=6379
    - JWT_SECRET=your-secret-key
  networks:
    - goshopadmin-network
  restart: always
  depends_on:
    - mysql
    - redis
  command: sh -c "go mod download && go run main.go"
```

#### 2.2 修改 Nginx 配置文件

```nginx
http {
    # 定义后端服务器组
    upstream backend_servers {
        # 轮询策略
        server goshopadmin-backend-1:8080 max_fails=3 fail_timeout=30s;
        server goshopadmin-backend-2:8080 max_fails=3 fail_timeout=30s;
        
        # 可选：使用权重策略
        # server goshopadmin-backend-1:8080 weight=5;
        # server goshopadmin-backend-2:8080 weight=5;
        
        # 可选：使用 IP 哈希策略（保持会话一致性）
        # ip_hash;
    }

    server {
        listen 80;
        server_name localhost;

        location / {
            root /usr/share/nginx/html;
            index index.html;
            try_files $uri $uri/ /index.html;
        }

        # API 路径转发到后端服务
        location /api/ {
            proxy_pass http://backend_servers/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # 健康检查
            proxy_connect_timeout 30s;
            proxy_read_timeout 30s;
            proxy_send_timeout 30s;
        }
    }
}
```

#### 2.3 修改前端 API 基础 URL

```javascript
// frontend/src/api/auth.js
// 修改为相对路径
const API_BASE_URL = '/api';

// 或使用 Nginx 地址
// const API_BASE_URL = 'http://localhost/api';
```

### 3. 部署方式

1. 复制上述配置到对应文件
2. 执行 `docker-compose up -d` 启动服务
3. 访问 `http://localhost` 测试负载均衡效果

## 二、Nginx 应对高并发的方案

### 1. 配置优化

#### 1.1 调整工作进程数

```nginx
worker_processes auto;  # 自动设置为 CPU 核心数
```

#### 1.2 调整连接数

```nginx
events {
    worker_connections 10240;  # 每个工作进程的最大连接数
    use epoll;  # 使用 epoll 事件模型
    multi_accept on;  # 允许一个工作进程同时接受多个连接
}
```

#### 1.3 启用 keepalive

```nginx
http {
    keepalive_timeout 65;  # 保持连接的超时时间
    keepalive_requests 100;  # 每个连接最多处理的请求数
}
```

#### 1.4 启用 Gzip 压缩

```nginx
http {
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
    gzip_comp_level 5;
    gzip_min_length 256;
    gzip_buffers 16 8k;
    gzip_http_version 1.1;
    gzip_vary on;
}
```

### 2. 缓存策略

#### 2.1 静态文件缓存

```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 30d;
    add_header Cache-Control "public, max-age=2592000";
    proxy_cache_valid 200 30d;
    proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=static_cache:10m max_size=10g inactive=60m use_temp_path=off;
}
```

#### 2.2 API 响应缓存

```nginx
proxy_cache_path /var/cache/nginx/api_cache levels=1:2 keys_zone=api_cache:10m max_size=10g inactive=10m use_temp_path=off;

location /api/ {
    proxy_cache api_cache;
    proxy_cache_valid 200 5m;
    proxy_cache_key "$scheme$request_method$host$request_uri";
    proxy_pass http://backend_servers/;
}
```

### 3. 限流策略

#### 3.1 基于 IP 的限流

```nginx
limit_req_zone $binary_remote_addr zone=ip_limit:10m rate=10r/s;

location /api/ {
    limit_req zone=ip_limit burst=20 nodelay;
    proxy_pass http://backend_servers/;
}
```

#### 3.2 基于请求路径的限流

```nginx
limit_req_zone $request_uri zone=path_limit:10m rate=20r/s;

location /api/ {
    limit_req zone=path_limit burst=30 nodelay;
    proxy_pass http://backend_servers/;
}
```

### 4. 健康检查

#### 4.1 后端服务健康检查

```nginx
upstream backend_servers {
    server goshopadmin-backend-1:8080 max_fails=3 fail_timeout=30s;
    server goshopadmin-backend-2:8080 max_fails=3 fail_timeout=30s;
    check interval=3000 rise=2 fall=3 timeout=1000 type=http;
    check_http_send "GET /health HTTP/1.0\r\nHost: localhost\r\n\r\n";
    check_http_expect_alive http_2xx http_3xx;
}
```

### 5. SSL 优化

#### 5.1 启用 HTTP/2

```nginx
server {
    listen 443 ssl http2;
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256';
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
}
```

### 6. 架构优化

#### 6.1 分层架构

- **接入层**：Nginx 处理 HTTPS、负载均衡、限流
- **应用层**：多个 Go 后端服务实例
- **数据层**：MySQL、Redis 集群

#### 6.2 水平扩展

- 使用 Docker Swarm 或 Kubernetes 管理容器
- 自动扩缩容根据流量调整实例数量
- 使用服务发现机制动态更新后端服务器列表

#### 6.3 CDN 加速

- 静态资源使用 CDN 加速
- API 响应缓存到 CDN
- 全球多区域部署，就近访问

### 7. 高级优化方案

#### 7.1 配置 Nginx 工作模式

```nginx
# 使用事件驱动模型
events {
    use epoll;  # Linux 平台推荐使用 epoll
    # use kqueue;  # FreeBSD/Mac OS X 平台推荐使用 kqueue
    # use select;  # 通用事件模型
}
```

#### 7.2 调整系统内核参数

```bash
# 增加文件描述符限制
ulimit -n 65536

# 修改系统内核参数
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_max_tw_buckets = 5000
net.ipv4.ip_local_port_range = 1024 65000
```

#### 7.3 启用 Nginx 线程池

```nginx
# 在 http 块中添加
thread_pool default threads=32 max_queue=65536;

# 在 location 块中使用
location / {
    aio threads=default;
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
}
```

#### 7.4 优化静态文件处理

```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    root /usr/share/nginx/html;
    expires 30d;
    add_header Cache-Control "public, max-age=2592000";
    add_header ETag "$entity_tag";
    add_header Last-Modified "$date_gmt";
    if_modified_since exact;
    etag on;
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    open_file_cache max=10000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;
    open_file_cache_errors on;
}
```

#### 7.5 配置连接池

```nginx
http {
    # 配置连接池
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    
    # 配置上游服务器连接池
    upstream backend_servers {
        server goshopadmin-backend-1:8080;
        server goshopadmin-backend-2:8080;
        keepalive 32;  # 保持 32 个空闲连接
    }
    
    # 在 location 中使用
    location /api/ {
        proxy_pass http://backend_servers;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

#### 7.6 启用 Nginx 缓存模块

```nginx
# 配置缓存路径
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=10g inactive=60m use_temp_path=off;

# 配置缓存键
proxy_cache_key "$scheme$request_method$host$request_uri$args";

# 在 location 中使用
location /api/ {
    proxy_cache my_cache;
    proxy_cache_valid 200 302 10m;
    proxy_cache_valid 404 1m;
    proxy_cache_bypass $http_cache_control;
    add_header X-Cache-Status $upstream_cache_status;
    proxy_pass http://backend_servers;
}
```

#### 7.7 配置限流和防 DDoS

```nginx
# 基于 IP 的限流
limit_req_zone $binary_remote_addr zone=ip_limit:10m rate=10r/s;

# 基于连接数的限流
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

# 限制请求体大小
client_max_body_size 10m;

# 配置限流规则
location /api/ {
    limit_req zone=ip_limit burst=20 nodelay;
    limit_conn conn_limit 10;
    proxy_pass http://backend_servers;
}
```

#### 7.8 启用 Nginx 模块优化

- **ngx_http_gzip_static_module**：预压缩静态文件
- **ngx_http_ssl_module**：SSL 优化
- **ngx_http_v2_module**：HTTP/2 支持
- **ngx_http_realip_module**：获取真实客户端 IP
- **ngx_http_headers_module**：自定义响应头
- **ngx_http_sub_module**：内容替换

#### 7.9 监控和调优

- **启用 Nginx 状态监控**：

```nginx
location /nginx_status {
    stub_status on;
    access_log off;
    allow 127.0.0.1;
    deny all;
}
```

- **使用 Prometheus 和 Grafana 监控**：
  - 安装 nginx-prometheus-exporter
  - 配置 Prometheus 采集 Nginx 指标
  - 使用 Grafana 可视化监控数据

- **定期分析 Nginx 日志**：
  - 使用 ELK 或 Graylog 分析访问日志
  - 识别异常流量和性能瓶颈
  - 优化热点路径和资源

## 三、性能监控

### 1. Nginx 监控

- 启用 nginx-status 模块
- 使用 Prometheus + Grafana 监控 Nginx 性能指标
- 配置告警机制，当性能指标异常时及时通知

### 2. 后端服务监控

- 每个后端服务暴露健康检查接口
- 监控服务响应时间、错误率、QPS 等指标
- 使用分布式追踪系统追踪请求链路

## 四、总结

通过以上方案，可以显著提升系统的并发处理能力和稳定性：

1. **负载均衡**：将请求分发到多个后端实例，提高系统整体处理能力
2. **配置优化**：调整 Nginx 配置参数，充分利用服务器资源
3. **缓存策略**：减少重复请求，减轻后端压力
4. **限流策略**：防止系统被突发流量击垮
5. **健康检查**：自动剔除故障节点，提高系统可用性
6. **架构优化**：通过分层架构和水平扩展，提高系统可扩展性
7. **高级优化**：通过系统内核调优、线程池配置等进一步提升性能
8. **监控分析**：实时监控系统状态，及时发现和解决问题

这些方案可以根据实际业务需求和服务器资源进行调整，以达到最佳的性能表现。