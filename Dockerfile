FROM golang:1.20-alpine AS builder

# 安装依赖
RUN apk add --no-cache gcc musl-dev

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -o infiniteimages .

# 使用更小的基础镜像
FROM alpine:3.17

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

# 创建必要的目录
RUN mkdir -p /app/static/images /app/logs
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/infiniteimages .

# 复制静态文件
COPY --from=builder /app/static ./static

# 暴露端口
EXPOSE 6868

# 设置环境变量
ENV SERVER_ADDR=0.0.0.0:6868
ENV STORAGE_TYPE=local
ENV LOCAL_STORAGE_PATH=static/images
ENV DEBUG_MODE=false

# 运行应用
CMD ["./infiniteimages"]