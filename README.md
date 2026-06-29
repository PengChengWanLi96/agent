# Agent

一个基于 Go 1.24 + Gin 的服务管理 Agent，提供 RESTful API 与内嵌 Web 控制台，用于 Docker 容器/镜像/网络/卷管理、Linux 系统指标采集、SSH 远程文件管理与命令执行。

## 功能特性

- **Docker 管理**
  - 容器：列出、查看详情、创建、启动、停止、重启、暂停、恢复、重命名、强制停止、删除、执行命令、查看日志、清理
  - 镜像：列出、查看详情、拉取、删除、清理
  - 网络：列出、创建、删除、连接/断开容器、清理
  - 卷：列出、创建、删除、清理
  - 支持本地 Unix Socket、远程 TCP 以及 TLS 连接 Docker 守护进程

- **Linux 系统指标采集**
  - 直接集成 Prometheus Node-Exporter（Go Module 依赖），无需独立进程
  - `/api/v1/metrics/collect`：返回结构化 JSON 指标（CPU、内存、磁盘、网络、负载等）
  - `/api/v1/metrics/prometheus`：返回原始 Prometheus 文本，可直接被 Prometheus Server 抓取
  - 仅 Linux 平台可用，其他平台返回 500

- **SSH 远程管理**
  - 建立 SSH 会话，支持密码或私钥认证
  - 远程文件浏览、上传、下载、新建目录、重命名、删除
  - 远程命令执行
  - Web 端提供交互式终端与文件管理器

- **文件上传**
  - `/api/v1/upload/files`：多文件上传到服务器本地目录
  - `/api/v1/upload/files`（GET）：列出上传目录文件

- **内嵌 Web 控制台**
  - 访问 `http://<addr>/` 即可打开
  - 原生 HTML/CSS/JS，无外部 CDN 依赖
  - 集成 xterm.js 终端界面

- **版本信息**
  - 支持 `--version` / `-v` 查看版本、Git Commit、构建时间、Go 版本、平台
  - 构建时自动通过 `-ldflags` 注入版本信息

## 架构

```
cmd/server/
    main.go              # 应用入口

internal/
    api/                 # Gin 路由与 Handler
        router.go
        docker_handler.go
        metrics_handler.go
        ssh_handler.go
        upload_handler.go
    service/             # 业务逻辑
        docker_service.go
        metrics_service.go
        ssh_service.go
        upload_service.go
        terminal_linux.go / terminal_stub.go
    client/              # 外部客户端/采集器
        docker/client.go
        metrics/collector_linux.go / collector_stub.go
        ssh/client.go
    config/              # 环境变量配置
    model/               # 数据模型
    version/             # 版本信息

web/                     # 内嵌前端静态资源（go:embed）

docs/                    # 详细文档
scripts/                 # 测试脚本与 curl 示例
```

## 快速开始

### 前置条件

- Go 1.24+
- Git
- Docker 环境（可选，用于 Docker 管理功能）

### 构建

```bash
# 为当前平台构建
./build.sh build

# 或直接使用 go build
 go build -o agent ./cmd/server

# 交叉编译 Linux 版本
 GOOS=linux GOARCH=amd64 go build -o agent ./cmd/server
```

> 推荐使用包路径 `./cmd/server` 构建，Go 会自动把 Git VCS 信息写入二进制。

### 运行

```bash
# 默认监听 :8080，连接本地 Docker
./agent

# 指定端口
./agent --port 9090

# 连接远程 Docker
./agent --docker-host tcp://192.168.1.100:2375

# 开发调试
./build.sh run
```

浏览器访问 `http://localhost:8080` 打开 Web 控制台。

## 配置

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_ADDR` | `:8080` | HTTP 服务监听地址 |
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker 守护进程地址 |
| `DOCKER_API_VERSION` | `v1.43` | Docker API 版本 |
| `DOCKER_TLS_VERIFY` | （空） | 非空时启用 TLS 验证 |
| `DOCKER_CERT_PATH` | （空） | TLS 证书目录（需含 cert.pem、key.pem、ca.pem） |

### 命令行参数

| 参数 | 简写 | 说明 |
|------|------|------|
| `--version` | `-v` | 打印版本信息并退出 |
| `--port` | `-p` | 覆盖 `SERVER_ADDR` 端口 |
| `--addr` | | 覆盖 `SERVER_ADDR` |
| `--docker-host` | | Docker 守护进程地址 |
| `--docker-api-version` | | Docker API 版本 |
| `--docker-tls-verify` | | 启用 Docker TLS 验证 |
| `--docker-cert-path` | | Docker TLS 证书目录 |

命令行参数优先级高于环境变量。

## API 概览

| 功能 | 方法 | 路径 |
|------|------|------|
| 健康检查 | GET | `/health` |
| 版本信息 | GET | `/api/v1/version` |
| 列出/创建容器 | GET / POST | `/api/v1/docker/containers` |
| 容器详情/生命周期/日志 | - | `/api/v1/docker/containers/:id/*` |
| 镜像管理 | - | `/api/v1/docker/images/*` |
| 网络管理 | - | `/api/v1/docker/networks/*` |
| 卷管理 | - | `/api/v1/docker/volumes/*` |
| 结构化指标 | GET | `/api/v1/metrics/collect` |
| Prometheus 指标 | GET | `/api/v1/metrics/prometheus` |
| SSH 连接 | POST | `/api/v1/ssh/connect` |
| SSH 会话/文件/命令/终端 | - | `/api/v1/ssh/sessions/:id/*` |
| 本地上传文件 | POST | `/api/v1/upload/files` |
| 列出本地上传目录 | GET | `/api/v1/upload/files` |

完整 API 文档见 [docs/api.md](docs/api.md)。

## 平台说明

- Docker 管理功能在所有支持 Docker 的平台上可用；Docker 不可用时相关接口返回 500。
- 系统指标采集依赖 Node-Exporter 的 Linux collector，仅在 Linux 上可用；非 Linux 平台返回 500。
- SSH 文件管理与命令执行在所有平台上可用。

## 测试

```bash
# 运行全部测试
go test ./...

# 运行单个包测试
go test ./internal/api -run TestHealthEndpoint

# 运行基准测试
go test ./internal/api -bench=.
```

## 交叉编译

```bash
# 构建所有支持平台
./build.sh all

# 指定版本号
VERSION=v1.2.0 ./build.sh all
```

输出文件：

| 平台 | 文件名 |
|------|--------|
| Linux amd64 | `dist/agent-{VERSION}-linux-amd64` |
| Linux arm64 | `dist/agent-{VERSION}-linux-arm64` |
| Windows amd64 | `dist/agent-{VERSION}-windows-amd64.exe` |
| macOS amd64 | `dist/agent-{VERSION}-darwin-amd64` |
| macOS arm64 | `dist/agent-{VERSION}-darwin-arm64` |

## 文档

- [API 文档](docs/api.md)
- [架构说明](docs/architecture.md)
- [构建指南](docs/build.md)
- [Docker 管理](docs/docker-api.md)
- [指标采集](docs/metrics-api.md)
- [SSH 文件管理](docs/ssh-file-management.md)

## 依赖

主要依赖：

- [gin-gonic/gin](https://github.com/gin-gonic/gin) — Web 框架
- [docker/docker](https://github.com/docker/docker) — Docker API 客户端
- [prometheus/node_exporter](https://github.com/prometheus/node_exporter) — 系统指标采集
- [pkg/sftp](https://github.com/pkg/sftp) — SFTP 文件操作
- [golang.org/x/crypto/ssh](https://golang.org/x/crypto/ssh) — SSH 客户端
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket（已预留）

完整依赖见 [go.mod](go.mod)。

## License

MIT
