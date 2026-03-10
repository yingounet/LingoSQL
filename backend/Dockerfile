# 构建阶段（modernc.org/sqlite 纯 Go 实现，无需 CGO）
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用（CGO_ENABLED=0 纯静态二进制）
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

# 运行阶段（含 mysqldump、pg_dump 用于备份）
FROM alpine:latest

RUN apk --no-cache add ca-certificates mysql-client postgresql-client

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/migrations ./migrations

# 创建数据目录
RUN mkdir -p /data

# 暴露端口
EXPOSE 3366

# 运行应用
CMD ["./server"]
