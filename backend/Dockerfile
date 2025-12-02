# 构建阶段：使用golang:alpine（包含Go）
FROM golang:1.24-alpine AS builder

# 安装SQLite开发库（CGO需要）
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 设置工作目录
WORKDIR /app

# 复制所有文件（包括vendor目录）
COPY . .

# 设置使用vendor模式（完全离线构建）
ENV GO111MODULE=on
ENV GOPROXY=direct
ENV GOSUMDB=off

# 构建应用（启用CGO，动态链接）
RUN CGO_ENABLED=1 GOOS=linux go build -mod=vendor -ldflags="-s -w" -o vaultseed-backend ./cmd/main.go

# 运行阶段：使用alpine（包含必要的运行时库）
FROM alpine:3.19

# 安装SQLite运行时库
RUN apk add --no-cache sqlite libc6-compat ca-certificates

# 创建非root用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/vaultseed-backend .

# 复制数据库文件（如果需要）
COPY --from=builder /app/vaultseed.db ./vaultseed.db

# 设置权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./vaultseed-backend"]
