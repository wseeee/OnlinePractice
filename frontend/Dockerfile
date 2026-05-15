# ==================== 阶段一：编译 Vue 前端 ====================
FROM node:20-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存层
COPY package.json package-lock.json ./
RUN npm ci --registry=https://registry.npmmirror.com

# 复制源码并构建
COPY . .
RUN npm run build

# ==================== 阶段二：Nginx 部署 ====================
FROM nginx:1.27-alpine

# 删除默认配置
RUN rm -f /etc/nginx/conf.d/default.conf

# 自定义 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/oj.conf

# 从构建阶段复制静态文件
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
