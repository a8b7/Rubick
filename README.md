# Rubick

一个现代化的 Docker 容器管理平台，提供直观的 Web UI 来管理 Docker 主机、容器、镜像、卷、网络和 Docker Compose 项目。

## 功能特性

- **多主机管理** - 支持本地 Docker socket、TCP 连接和 SSH 隧道连接远程 Docker 主机
- **容器管理** - 完整的容器生命周期管理，包括创建、启动、停止、重启、删除等操作
- **镜像管理** - 镜像拉取、删除、标签管理、镜像搜索
- **卷管理** - Docker 卷的创建、删除和浏览
- **网络管理** - Docker 网络的创建和删除
- **Docker Compose** - Compose 项目的创建、部署、启动、停止和日志查看
- **实时日志** - 通过 WebSocket 实时查看容器和 Compose 项目日志
- **终端执行** - 支持在容器内执行命令
- **审计日志** - 记录所有 API 操作
- **安全加密** - SSH 密码和私钥使用 AES 加密存储

## 技术栈

### 后端
- **Go 1.24** - 核心运行时
- **Gin** - HTTP Web 框架
- **GORM** - ORM 库
- **SQLite** - 嵌入式数据库
- **Docker SDK** - Docker 客户端库
- **Viper** - 配置管理
- **Gorilla WebSocket** - WebSocket 支持

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全
- **Vite 7** - 构建工具
- **Tailwind CSS 4** - 原子化 CSS 框架
- **DaisyUI 5** - UI 组件库
- **Pinia** - 状态管理
- **Vue Router 4** - 路由管理
- **Axios** - HTTP 客户端
- **xterm.js** - 终端模拟器
- **Iconify** - 图标库

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+ & pnpm
- Docker
- CGO (SQLite 依赖)

### 安装与运行

#### 开发模式

1. 克隆项目
```bash
git clone <repository-url>
cd Rubick
```

2. 启动后端
```bash
# 下载依赖
make deps

# 运行后端
make run
```

3. 启动前端开发服务器
```bash
cd web
pnpm install
pnpm run dev
```

4. 访问应用
- 前端开发服务器: http://localhost:3000
- 后端 API: http://localhost:8080/api/v1

#### 生产构建

```bash
# 完整构建（前端 + 后端）
make build-full

# 仅构建后端
make build

# 运行
./bin/rubick
```

访问 http://localhost:8080

## 项目结构

```
Rubick/
├── cmd/server/main.go           # 应用程序入口
├── internal/
│   ├── config/                  # 配置加载
│   ├── crypto/                  # AES 加密工具
│   ├── database/                # 数据库初始化
│   ├── docker/                  # Docker 客户端管理
│   │   ├── client.go            # 客户端管理器
│   │   ├── connection.go        # 连接接口
│   │   ├── local.go             # 本地连接
│   │   ├── tcp.go               # TCP 连接
│   │   ├── ssh.go               # SSH 隧道连接
│   │   ├── container_service.go # 容器服务
│   │   ├── image_service.go     # 镜像服务
│   │   ├── compose_service.go   # Compose 服务
│   │   └── executor.go          # 命令执行器
│   ├── handler/                 # HTTP 处理器
│   │   ├── router.go            # 路由定义
│   │   ├── host_handler.go      # 主机管理
│   │   ├── container_handler.go # 容器操作
│   │   ├── image_handler.go     # 镜像管理
│   │   ├── volume_handler.go    # 卷管理
│   │   ├── network_handler.go   # 网络管理
│   │   ├── compose_handler.go   # Compose 管理
│   │   ├── audit_handler.go     # 审计日志
│   │   └── websocket.go         # WebSocket 日志
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据库操作
│   └── static/                  # 嵌入的前端文件
├── web/                         # Vue 3 前端
│   ├── src/
│   │   ├── api/                 # API 客户端
│   │   ├── router/              # 路由配置
│   │   ├── stores/              # Pinia 状态
│   │   ├── views/               # 页面组件
│   │   └── components/          # 可复用组件
│   └── vite.config.ts
├── configs/config.yaml          # 默认配置
├── Makefile                     # 构建脚本
└── go.mod                       # Go 依赖
```

## API 路由

| 路由 | 方法 | 描述 |
|------|------|------|
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/hosts` | GET/POST | 主机列表/创建 |
| `/api/v1/hosts/:id` | GET/PUT/DELETE | 主机详情/更新/删除 |
| `/api/v1/hosts/:id/test` | POST | 测试主机连接 |
| `/api/v1/containers` | GET | 容器列表 |
| `/api/v1/containers/:id/*` | * | 容器操作 |
| `/api/v1/images` | GET | 镜像列表 |
| `/api/v1/images/*` | * | 镜像操作 |
| `/api/v1/volumes` | GET/POST | 卷列表/创建 |
| `/api/v1/networks` | GET/POST | 网络列表/创建 |
| `/api/v1/compose/projects` | GET/POST | Compose 项目列表/创建 |
| `/api/v1/compose/*` | * | Compose 相关操作 |
| `/api/v1/audit/logs` | GET | 审计日志查询 |
| `/ws/containers/:id/logs` | WS | 容器实时日志 |
| `/ws/compose/:id/logs` | WS | Compose 实时日志 |

## 配置

配置文件位于 `configs/config.yaml`，也支持环境变量覆盖（前缀 `RUBICK_`）。

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/rubick.db"
```

### 环境变量

| 变量 | 描述 |
|------|------|
| `RUBICK_SERVER_HOST` | 服务器监听地址 |
| `RUBICK_SERVER_PORT` | 服务器监听端口 |
| `RUBICK_DATABASE_PATH` | 数据库文件路径 |
| `RUBICK_ENCRYPTION_KEY` | 敏感数据加密密钥（32字节） |

## 常用命令

```bash
# 构建后端
make build

# 运行开发服务器
make run

# 运行测试
make test

# 下载依赖
make deps

# 清理构建产物
make clean

# 完整构建（前端 + 后端）
make build-full
```

## 前端开发

```bash
cd web

# 安装依赖
pnpm install

# 开发服务器
pnpm run dev

# 生产构建
pnpm run build

# 预览构建
pnpm run preview
```

## 安全说明

- SSH 密码和私钥使用 AES-256-GCM 加密存储
- 加密密钥通过环境变量 `RUBICK_ENCRYPTION_KEY` 配置
- 请勿将 `.env` 文件提交到版本控制

## License

MIT
