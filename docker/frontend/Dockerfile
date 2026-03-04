# 开发阶段
FROM node:20-alpine AS dev

# 设置工作目录
WORKDIR /app

# 复制package.json和package-lock.json文件
COPY ../../frontend/package*.json ./

# 下载依赖
RUN npm install && npm install terser

# 复制源代码
COPY ../../frontend/ ./

# 暴露端口
EXPOSE 5173

# 运行开发服务器
CMD ["npm", "run", "dev", "--", "--host"]

# 生产构建阶段
FROM node:20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制package.json和package-lock.json文件
COPY ../../frontend/package*.json ./

# 下载依赖
RUN npm install && npm install terser

# 复制源代码
COPY ../../frontend/ ./

# 构建应用
RUN npm run build

# 运行阶段
FROM nginx:1.20-alpine

# 复制构建产物到Nginx静态目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 前端开发环境使用Vite开发服务器，不需要Nginx配置

# 暴露端口
EXPOSE 80

# 运行Nginx
CMD ["nginx", "-g", "daemon off;"]
