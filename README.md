# LingoSQL - 数据库管理工具

LingoSQL 是一个现代化的 Web 数据库管理工具，支持 MySQL 和 PostgreSQL 数据库的管理。

## 功能特性

- ✅ 多数据库支持（MySQL、PostgreSQL）
- ✅ 用户认证与授权系统
- ✅ 数据库连接管理
- ✅ SQL 查询执行与结果展示
- ✅ 查询历史记录
- 🔄 查询收藏功能（开发中）
- 🔄 数据库/表创建与管理（开发中）
- 🔄 AI 功能：自然语言生成 SQL（计划中）

## 技术栈

### 后端
- Golang 1.21+
- Gin Web 框架
- SQLite（服务端数据库）
- MySQL/PostgreSQL 驱动

### 前端
- Vue 3 + TypeScript
- Element Plus UI 组件库
- Vite 构建工具
- Pinia 状态管理

## 快速开始

### 开发环境

#### 后端

1. 安装依赖：
```bash
cd backend
go mod download
```

2. 配置环境变量（可选）：
```bash
export CONFIG_PATH=./config.yaml
export DB_PATH=./data/lingosql.db
```

3. 运行服务器：
```bash
go run cmd/server/main.go
```

服务器将在 `http://localhost:8080` 启动。

#### 前端

1. 安装依赖：
```bash
cd frontend
npm install
```

2. 启动开发服务器：
```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动。

### 生产部署

#### 使用 Docker

1. 构建镜像：
```bash
docker build -t lingosql:latest -f backend/Dockerfile .
```

2. 运行容器：
```bash
docker run -d \
  --name lingosql \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e DB_PATH=/data/lingosql.db \
  -e JWT_SECRET=your-secret-key \
  lingosql:latest
```

#### 使用 Docker Compose

```bash
docker-compose up -d
```

## 配置说明

配置文件位于 `backend/config.yaml`：

```yaml
server:
  port: 8080
  mode: debug # debug, release

database:
  path: ./data/lingosql.db

jwt:
  secret: your-secret-key-change-in-production
  expires_in: 720h # 30 days

encryption:
  key: your-encryption-key-32-chars-long!! # 32 characters for AES-256

log:
  level: info # debug, info, warn, error
  format: json # json, text
```

**重要提示：** 在生产环境中，请务必修改 `jwt.secret` 和 `encryption.key` 的值！

## API 文档

API 文档请参考 [docs/04-API设计.md](./docs/04-API设计.md)

## 项目结构

```
LingoSQL/
├── backend/          # 后端代码
│   ├── cmd/         # 应用入口
│   ├── internal/    # 内部包
│   ├── migrations/  # 数据库迁移
│   └── pkg/         # 可复用包
├── frontend/        # 前端代码
│   └── src/         # 源代码
├── docs/            # 文档
└── data/            # 数据目录（自动创建）
```

## 开发计划

详细开发计划请参考 [docs/09-开发计划.md](./docs/09-开发计划.md)

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
