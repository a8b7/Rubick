# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 项目概述

Rubick 是一个 Docker 容器管理平台，后端使用 Go，前端使用 Vue 3。提供 Web UI 来管理 Docker 主机、容器、镜像、卷、网络和 Docker Compose 项目。

## 项目架构

```
Rubick/
├── cmd/server/main.go           # 应用程序入口
├── internal/
│   ├── config/                  # 配置加载 (Viper)
│   ├── crypto/                  # 敏感字段 AES 加密
│   ├── database/                # SQLite 初始化 (GORM)
│   ├── docker/                  # Docker 宥户端管理
│   │   ├── client.go            # ClientManager 单例，管理 Docker 连接
│   │   ├── connection.go        # 连接接口
│   │   ├── local.go             # 本地 Docker socket 连接
│   │   ├── tcp.go               # TCP 连接 Docker API
│   │   ├── ssh.go               # SSH 隧道 Docker 连接
│   │   ├── container_service.go # 容器操作
│   │   ├── image_service.go     # 镜像操作
│   │   ├── compose_service.go   # Docker Compose 操作
│   │   ├── executor.go          # 命令执行器接口
│   │   ├── local_executor.go    # 本地命令执行
│   │   └── ssh_executor.go      # SSH 远程命令执行
│   ├── handler/                 # HTTP 处理器 (Gin)
│   │   ├── router.go            # 路由定义
│   │   ├── host_handler.go      # Docker 主机 CRUD
│   │   ├── container_handler.go # 容器操作
│   │   ├── image_handler.go     # 镜像管理
│   │   ├── volume_handler.go    # 卷管理
│   │   ├── network_handler.go   # 网络管理
│   │   ├── compose_handler.go   # Docker Compose 项目管理
│   │   ├── audit_handler.go     # 审计日志查询
│   │   ├── audit_middleware.go  # 请求日志中间件
│   │   ├── middleware.go        # 通用中间件
│   │   ├── response.go          # 统一 API 响应工具
│   │   └── websocket.go         # 实时日志流
│   ├── model/                   # 数据模型 (GORM)
│   │   ├── host.go              # Host, Certificate, ComposeProject
│   │   └── audit_log.go         # AuditLog 模型
│   ├── repository/              # 数据库操作
│   │   ├── host_repository.go
│   │   ├── compose_repository.go
│   │   └── audit_repository.go
│   └── static/                  # 内嵌前端构建文件
├── web/                         # Vue 3 前端
│   ├── src/
│   │   ├── api/                 # Axios API 客户端
│   │   │   ├── request.ts       # Axios 实例，配置基础 URL
│   │   │   ├── host.ts          # 主机 API
│   │   │   ├── container.ts     # 容器 API
│   │   │   ├── image.ts         # 镜像 API
│   │   │   ├── compose.ts       # Compose API
│   │   │   └── index.ts         # API 导出
│   │   ├── router/index.ts      # Vue Router 配置
│   │   ├── stores/              # Pinia 状态管理
│   │   │   ├── host.ts          # 当前 Docker 主机状态
│   │   │   └── index.ts         # Store 导出
│   │   ├── views/               # 页面组件
│   │   │   ├── hosts/Index.vue  # 主机管理页
│   │   │   ├── containers/      # 容器页面
│   │   │   │   ├── Index.vue    # 容器列表
│   │   │   │   └── Detail.vue   # 容器详情
│   │   │   ├── images/Index.vue # 镜像列表页
│   │   │   └── compose/         # Compose 项目页面
│   │   │       ├── Index.vue    # 项目列表
│   │   │       ├── Form.vue     # 创建/编辑表单
│   │   │       └── Detail.vue   # 项目详情
│   │   ├── components/          # 可复用组件
│   │   │   ├── layout/
│   │   │   │   └── AppLayout.vue
│   │   │   ├── Confirm.vue      # 确认对话框
│   │   │   ├── DirBrowser.vue   # 目录浏览器
│   │   │   ├── DirectoryUpload.vue # 目录上传
│   │   │   ├── LogViewer.vue    # 实时日志查看器
│   │   │   └── Toast.vue        # Toast 通知
│   │   └── utils/toast.ts       # Toast 辅助函数
│   └── vite.config.ts
└── configs/config.yaml          # 服务器配置
```

## 常用命令

### 后端 (Go)
```bash
# 仅构建后端
make build

# 开发模式运行
make run

# 运行测试
make test

# 下载依赖
make deps

# 清理构建产物
make clean
```

### 前端 (Vue)
```bash
cd web

# 安装依赖
pnpm install

# 开发服务器 (代理 /api 到 :8080)
pnpm run dev

# 生产构建
pnpm run build

# 预览生产构建
pnpm run preview
```

### 完整构建
```bash
# 构建前端 + 后端 (复制 web/dist 到 internal/static)
make build-full
```

### 运行应用
```bash
# 启动后端 (同时提供 API 和前端)
./bin/rubick
# 或
make run

# 访问地址 http://localhost:8080
# API 地址 http://localhost:8080/api/v1/
```

## 核心模式

### 后端
- **Handler → Repository → Model** 分层架构
- Docker 客户端连接由 `docker.ClientManager` 单例管理，支持本地 socket、TCP 和 SSH 隧道
- 命令执行器 (`local_executor.go`, `ssh_executor.go`) 用于运行 Docker Compose 命令
- 敏感字段（SSH 密码、私钥）使用 `internal/crypto` 进行静态加密
- WebSocket 端点用于实时容器/Compose 日志
- GORM 钩子 (`BeforeCreate`, `AfterFind`) 透明处理加密/解密
- 审计中间件记录所有 API 请求

### 前端
- **Vue 3 + TypeScript + Vite 7**
- **Tailwind CSS v4 + DaisyUI v5** 样式框架（不是 Element Plus）
- **Pinia** 状态管理（`useHostStore` 管理当前 Docker 主机）
- **Vue Router 4** 懒加载路由
- **Axios** API 调用，基础 URL `/api/v1`
- **@iconify/vue** 图标
- API 响应格式 `{ code, message, data }`

### 前端路由
| 路径 | 组件 | 描述 |
|------|------|------|
| `/hosts` | `views/hosts/Index.vue` | Docker 主机管理 |
| `/containers` | `views/containers/Index.vue` | 容器列表 |
| `/containers/:id` | `views/containers/Detail.vue` | 容器详情 |
| `/images` | `views/images/Index.vue` | 镜像列表 |
| `/compose` | `views/compose/Index.vue` | Compose 项目列表 |
| `/compose/create` | `views/compose/Form.vue` | 创建项目 |
| `/compose/:id` | `views/compose/Detail.vue` | 项目详情 |
| `/compose/:id/edit` | `views/compose/Form.vue` | 编辑项目 |

### API 路由 (前缀 `/api/v1`)
| 路由 | 处理器 | 描述 |
|------|--------|------|
| `GET /health` | - | 健康检查 |
| `GET /audit/logs` | ListAuditLogs | 审计日志查询 |
| `/hosts` | host_handler | Docker 主机 CRUD + 连接测试 |
| `/containers` | container_handler | 容器操作 + exec + 日志 |
| `/images` | image_handler | 镜像拉取/删除/标签/搜索 |
| `/volumes` | volume_handler | 卷 CRUD |
| `/networks` | network_handler | 网络 CRUD |
| `/compose/projects` | compose_handler | Compose 项目 CRUD |
| `/compose/browse` | compose_handler | 目录浏览 |
| `/compose/scan` | compose_handler | 扫描 compose 文件 |
| `/compose/upload` | compose_handler | 目录上传 |
| `/ws/containers/:id/logs` | websocket | 实时容器日志 |
| `/ws/compose/:id/logs` | websocket | 实时 Compose 日志 |

## 配置

默认配置在 `configs/config.yaml`:
- 服务器: `0.0.0.0:8080`
- 数据库: SQLite 位于 `./data/rubick.db`
- 环境变量前缀: `RUBICK_`

## 注意事项

- SQLite 需要 CGO_ENABLED=1（使用 CGO sqlite 驱动）
- 前端开发服务器运行在端口 3000，代理 API 到后端 8080
- 生产构建通过 `internal/static` 将前端嵌入 Go 二进制文件
