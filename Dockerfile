# 构建阶段
FROM golang:alpine AS builder

# 设置国内 Go 模块代理，解决网络超时问题
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 复制依赖文件和第三方模块
COPY go.mod go.sum ./
COPY third_party/ third_party/

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o myNetdisk main.go

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/myNetdisk .

# 复制模板和静态文件
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# 创建上传目录并设置权限
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app/uploads

# 暴露端口
EXPOSE 8080

# 切换到非 root 用户
USER appuser

# 启动应用
CMD ["./myNetdisk"]