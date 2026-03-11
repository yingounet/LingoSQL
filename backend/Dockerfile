# 阶段 1：构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@latest --activate
COPY frontend/ ./frontend/
WORKDIR /app/frontend
RUN pnpm install --frozen-lockfile
RUN mkdir -p ../backend && pnpm run build

# 阶段 2：构建后端
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
# 复制前端构建产物
COPY --from=frontend-builder /app/backend/web ./web
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

# 阶段 3：运行镜像
FROM alpine:3.23

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/config.yaml .
COPY --from=backend-builder /app/migrations ./migrations
COPY --from=backend-builder /app/web ./web

RUN mkdir -p /data

EXPOSE 3366

CMD ["./server"]
