# 构建阶段
FROM golang:1.21-alpine AS builder

# go-sqlite3 需要 CGO，安装 gcc 和 sqlite-dev
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

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
