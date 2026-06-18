# 系统架构文档

## 概述

本项目是一个基于 Go-Gin 框架的服务管理 Agent，提供 RESTful API 接口，用于：
- Docker 容器管理（集成 docker-client）
- Linux 系统指标采集（直接集成 node-exporter v1.5.0 作为 Go Module 依赖）

## 架构设计

采用分层架构，便于后续扩展更多功能模块：

```
cmd/server/
    main.go              # 应用入口，负责初始化各组件

internal/
    api/                 # HTTP 接口层（Handler + Router）
        router.go        # 路由注册
        docker_handler.go    # Docker API 处理器
        metrics_handler.go   # Metrics API 处理器

    service/             # 业务逻辑层
        service.go       # Docker/Metrics 业务服务

    client/              # 外部客户端/采集器层
        docker/          # Docker HTTP API 客户端
        metrics/         # Node-Exporter 采集器（Go Module 依赖集成）
            collector_linux.go   # Linux 实现（使用 node_exporter collector）
            collector_stub.go    # 非 Linux 平台 stub

    config/              # 配置管理
    model/               # 数据模型定义

docs/                    # 文档目录
```

## 分层职责

### 1. API 层（internal/api）
- 负责 HTTP 请求路由和参数解析
- 调用 Service 层处理业务逻辑
- 统一响应格式封装

### 2. Service 层（internal/service）
- 封装业务逻辑
- 协调多个客户端操作
- 为 API 层提供清晰的接口

### 3. Client 层（internal/client）
- 封装外部系统的调用细节或系统采集逻辑
- **Docker Client**：通过 HTTP API 连接 Docker 守护进程，支持 Unix Socket 和 TCP/TLS
- **Node-Exporter Collector**：直接集成 `github.com/prometheus/node_exporter` 作为 Go Module 依赖，通过其 collector 包采集 Linux 系统指标

### 4. Config 层（internal/config）
- 集中管理环境变量配置
- 支持 Docker 连接参数配置（host、TLS、API 版本）

### 5. Model 层（internal/model）
- 定义跨层使用的数据结构
- 包含请求/响应 DTO

## Node-Exporter 集成说明

Node-Exporter 以 Go Module 形式集成，版本在 `go.mod` 中显式指定：

```go
require (
    github.com/prometheus/node_exporter v1.5.0
)
```

### 版本升级

后续升级 node-exporter 版本时，只需修改 `go.mod` 中的版本号：

```bash
# 方式一：直接修改 go.mod
# 将 github.com/prometheus/node_exporter v1.5.0 改为目标版本

# 方式二：使用 go get
go get github.com/prometheus/node_exporter@v1.6.0

# 然后更新依赖
go mod tidy
```

### 跨平台编译

由于 node_exporter 的 collector 仅在 Linux 上可用，项目使用 Go Build Tags 实现跨平台支持：
- `collector_linux.go`：Linux 平台使用真正的 node_exporter collector
- `collector_stub.go`：其他平台返回错误提示

在 Linux 服务器上编译和运行：
```bash
GOOS=linux GOARCH=amd64 go build -o agent cmd/server/main.go
```

#### 在 Windows 上交叉编译 Linux 版本

> 说明：`GOARCH=amd64` 对应 Linux x86_64（x86），`GOARCH=arm64` 对应 Linux ARM64。

**Windows DOS（cmd.exe）**

```bat
:: 编译 Linux x86（amd64）
set GOOS=linux
set GOARCH=amd64
go build -o agent-linux-amd64 cmd/server/main.go

:: 编译 Linux arm64
set GOOS=linux
set GOARCH=arm64
go build -o agent-linux-arm64 cmd/server/main.go
```

> 注意：`set` 设置的环境变量在当前命令行窗口内会一直生效，编译完成后建议执行 `set GOOS=` 与 `set GOARCH=` 清除，避免影响后续本地编译。

**Windows PowerShell**

```powershell
# 编译 Linux x86（amd64）
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o agent-linux-amd64 cmd/server/main.go

# 编译 Linux arm64
$env:GOOS="linux"; $env:GOARCH="arm64"; go build -o agent-linux-arm64 cmd/server/main.go
```

**Windows Git Bash**

```bash
# 编译 Linux x86（amd64）
GOOS=linux GOARCH=amd64 go build -o agent-linux-amd64 cmd/server/main.go

# 编译 Linux arm64
GOOS=linux GOARCH=arm64 go build -o agent-linux-arm64 cmd/server/main.go
```

## Docker 配置说明

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| SERVER_ADDR | :8080 | HTTP 服务监听地址 |
| DOCKER_HOST | unix:///var/run/docker.sock | Docker 守护进程地址 |
| DOCKER_API_VERSION | v1.43 | Docker API 版本 |
| DOCKER_TLS_VERIFY | （空） | 是否启用 TLS 验证（非空即启用） |
| DOCKER_CERT_PATH | （空） | TLS 证书目录路径（需包含 cert.pem, key.pem, ca.pem） |

## 命令行参数

程序支持以下启动参数，**命令行选项优先级高于环境变量**：

| 参数 | 简写 | 说明 |
|------|------|------|
| `--version` | `-v` | 打印版本信息（版本号、git commit、构建日期、Go 版本、平台）并退出 |
| `--help` | `-h` | 打印帮助信息并退出 |
| `--port` | `-p` | HTTP 服务监听端口（覆盖 `SERVER_ADDR` 中的端口），范围 1-65535 |
| `--addr` | | HTTP 服务监听地址，如 `:8080` 或 `0.0.0.0:8080`（覆盖 `SERVER_ADDR`） |
| `--docker-host` | | Docker 守护进程地址 |
| `--docker-api-version` | | Docker API 版本 |
| `--docker-tls-verify` | | 启用 Docker TLS 验证 |
| `--docker-cert-path` | | Docker TLS 证书目录路径 |

示例：

```bash
# 查看版本信息
./agent --version

# 查看帮助
./agent --help

# 指定端口启动
./agent --port 9090

# 指定监听地址和远程 Docker
./agent --addr 0.0.0.0:8080 --docker-host tcp://192.168.1.100:2375
```

### 注入版本信息（编译时）

`--version` 输出的版本信息可在编译时通过 `-ldflags` 注入。推荐在构建产品包时使用以下命令，把**构建时间**、**git commit ID**、**git 提交时间**写入二进制。**Git 时间与构建时间统一使用 UTC 表示，符合 ISO 8601 国际标准：**

```bash
# 本地当前平台
go build -ldflags "\
  -X agent/internal/version.Version=v1.2.0 \
  -X agent/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X agent/internal/version.GitTime=$(git log -1 --format=%cI --date=iso-strict) \
  -X agent/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o agent ./cmd/server

# Linux 交叉编译
GOOS=linux GOARCH=amd64 go build -ldflags "\
  -X agent/internal/version.Version=v1.2.0 \
  -X agent/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X agent/internal/version.GitTime=$(git log -1 --format=%cI --date=iso-strict) \
  -X agent/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o agent-linux-amd64 ./cmd/server
```

Windows PowerShell 示例：

```powershell
$ver = "v1.2.0"
$commit = git rev-parse --short HEAD
$gitTime = git log -1 --format=%cI --date=iso-strict
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
go build -ldflags "-X agent/internal/version.Version=$ver -X agent/internal/version.GitCommit=$commit -X agent/internal/version.GitTime=$gitTime -X agent/internal/version.BuildDate=$buildDate" -o agent.exe ./cmd/server
```

`--version` 输出示例：

```
agent v1.2.0
  git commit: a1b2c3d
  git time:   2026-06-18T11:40:32Z
  build date: 2026-06-18T11:45:00Z
  go version: go1.24.3
  platform:   linux/amd64
```

| 字段 | 变量名 | 说明 |
|------|--------|------|
| 版本号 | `Version` | 产品版本号，如 `v1.2.0` |
| Git Commit | `GitCommit` | 当前代码的短 commit ID |
| Git 提交时间 | `GitTime` | 最近一条 commit 的提交时间（UTC） |
| 构建时间 | `BuildDate` | 本次二进制构建的 UTC 时间 |

未注入时，若项目位于 Git 仓库中且使用包路径构建，Go 会自动从 VCS 信息回退填充 `GitCommit`、`GitTime` 与 `BuildDate`，VCS 时间会被转换为 UTC 输出。

## 运行方式

### 首次编译（需要网络环境下载依赖）

```bash
# 下载依赖（需要能访问 Go Module 代理或 GitHub）
go mod tidy

# 编译
go build -o agent cmd/server/main.go
```

### 直接运行

```bash
# 本地 Docker
go run cmd/server/main.go

# 远程 Docker TCP
DOCKER_HOST=tcp://192.168.1.100:2375 go run cmd/server/main.go

# 远程 Docker TLS
DOCKER_HOST=tcp://192.168.1.100:2376 DOCKER_TLS_VERIFY=1 DOCKER_CERT_PATH=/certs go run cmd/server/main.go
```

### Docker 编译（推荐用于 Linux 服务器部署）

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o agent cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent .
CMD ["./agent"]
```

## 扩展性设计

新增功能时，只需在对应层添加代码：
1. **Client 层**：添加新的外部客户端封装或系统采集逻辑
2. **Service 层**：添加对应的业务逻辑
3. **API 层**：注册新的路由和处理器

例如新增 Kubernetes 管理：
- `internal/client/k8s/client.go`
- `internal/service/k8s_service.go`
- `internal/api/k8s_handler.go`
