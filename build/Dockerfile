FROM golang:1.22-alpine AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制源代码 (需要在go mod download之前，因为有local replace)
COPY . .

# 下载依赖
RUN go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/netdisk

# 最终镜像
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建应用用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 复制静态文件和模板
COPY --from=builder /app/web ./web

# 创建运行时目录
RUN mkdir -p /app/data /app/storage /app/logs && \
    chown -R appuser:appgroup /app/data /app/storage /app/logs

# 切换到应用用户
USER appuser

# 设置环境变量
ENV APP_ENV=production

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./main"]
