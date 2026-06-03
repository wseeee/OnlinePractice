# ==================== 阶段一：编译 Go 后端 ====================
FROM golang:1.26-alpine AS builder

# 国内镜像加速（阿里云 GOPROXY）
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN go build -ldflags="-s -w" -o /app/server ./cmd/server

# ==================== 阶段二：最小运行镜像 ====================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /app/server .

# Swagger 文档目录
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./server"]
