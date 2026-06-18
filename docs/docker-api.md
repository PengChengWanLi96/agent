# Docker 管理功能说明

## 功能概述

集成 Docker HTTP API，提供容器生命周期管理能力。Agent 通过用户显式配置的 Docker 连接信息连接到 Docker 守护进程。

## 技术实现

- **通信方式**: Docker HTTP API
- **连接方式**: 支持 Unix Socket 和 TCP（含 TLS）
- **配置方式**: 通过环境变量显式配置

## Docker 连接配置

用户需要通过环境变量显式配置 Docker 连接信息：

### 本地 Docker（默认）
```bash
export DOCKER_HOST=unix:///var/run/docker.sock
```

### 远程 TCP Docker
```bash
export DOCKER_HOST=tcp://192.168.1.100:2375
```

### 远程 TLS Docker
```bash
export DOCKER_HOST=tcp://192.168.1.100:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/path/to/certs
```

证书目录需要包含：
- `cert.pem` - 客户端证书
- `key.pem` - 客户端私钥
- `ca.pem` - CA 证书

## 支持的操作

| 操作 | 方法 | 说明 |
|-----|------|------|
| 列出容器 | GET /containers/json | 支持筛选运行中/全部容器 |
| 查看详情 | GET /containers/{id}/json | 获取容器配置、状态、网络等信息 |
| 启动容器 | POST /containers/{id}/start | 启动已停止的容器 |
| 停止容器 | POST /containers/{id}/stop | 支持自定义超时时间 |
| 删除容器 | DELETE /containers/{id} | 支持强制删除 |
| 查看日志 | GET /containers/{id}/logs | 支持 stdout + stderr，自定义行数 |

## 数据模型

### Container（容器概要）

```go
type Container struct {
    ID      string   // 容器短 ID（12位）
    Names   []string // 容器名称列表
    Image   string   // 镜像名称
    State   string   // 运行状态（running/exited/...）
    Status  string   // 状态描述
    Created int64    // 创建时间戳
}
```

### ContainerDetail（容器详情）

```go
type ContainerDetail struct {
    ID      string         // 完整容器 ID
    Name    string         // 容器名称
    Image   string         // 镜像名称
    State   ContainerState // 运行状态详情
    Config  map[string]interface{} // 配置信息
    Network map[string]interface{} // 网络设置
}
```

## 扩展建议

后续可增加以下 Docker 管理功能：
- 镜像管理（拉取、删除、列出镜像）
- 网络管理（创建、删除、查看网络）
- 卷管理（创建、删除、挂载卷）
- 容器创建和运行（支持 Dockerfile 或 Compose）
- 容器资源限制和监控
