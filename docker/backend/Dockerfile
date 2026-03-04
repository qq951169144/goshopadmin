# 开发阶段
FROM golang:1.24-alpine AS dev

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY ../../backend/go.mod ../../backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY ../../backend/ ./

# 暴露端口
EXPOSE 8080

# 运行开发服务器
CMD ["go", "run", "main.go"]

# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY ../../backend/go.mod ../../backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY ../../backend/ ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o goshopadmin-server main.go

# 运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/goshopadmin-server ./

# 复制环境变量示例文件
COPY ../../backend/.env.example ./.env

# 设置环境变量
ENV SERVER_PORT=8080
ENV DB_HOST=mysql
ENV DB_PORT=3306
ENV DB_USER=root
ENV DB_PASSWORD=password
ENV DB_NAME=goshopadmin
ENV REDIS_HOST=redis
ENV REDIS_PORT=6379
ENV REDIS_PASSWORD=
ENV REDIS_DB=0
ENV JWT_SECRET=your-secret-key
ENV JWT_EXPIRE_HOUR=24
ENV MQ_HOST=rabbitmq
ENV MQ_PORT=5672
ENV MQ_USER=guest
ENV MQ_PASSWORD=guest
ENV MQ_VHOST=/

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./goshopadmin-server"]
