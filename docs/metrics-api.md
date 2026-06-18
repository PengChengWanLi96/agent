# Node-Exporter 集成说明

## 功能概述

直接集成 Prometheus Node-Exporter 作为 Go Module 依赖，版本在 `go.mod` 中显式指定。通过 node-exporter 的 collector 包采集 Linux 系统指标，无需部署独立的 node-exporter 进程。

## 技术实现

- **集成方式**: Go Module 依赖 (`github.com/prometheus/node_exporter`)
- **指定版本**: `v1.5.0`
- **采集方式**: 调用 node_exporter collector 的 `Update()` 方法
- **输出格式**: 支持 JSON 结构化数据和 Prometheus 原始文本

## 版本管理

### 当前版本

在 `go.mod` 中指定：

```go
require (
    github.com/prometheus/node_exporter v1.5.0
)
```

### 版本升级

后续升级 node-exporter 版本时，修改 `go.mod` 中的版本号并更新依赖：

```bash
# 方式一：使用 go get 指定新版本
go get github.com/prometheus/node_exporter@v1.6.0

# 方式二：手动修改 go.mod 后执行
go mod tidy

# 验证升级
go list -m github.com/prometheus/node_exporter
```

## 跨平台支持

由于 node_exporter 的 collector 仅在 Linux 上可用，项目使用 Go Build Tags 实现跨平台编译：

| 文件 | Build Tag | 说明 |
|-----|-----------|------|
| `collector_linux.go` | `linux` | 使用 node_exporter 真正的 collector |
| `collector_stub.go` | `!linux` | 返回错误提示（非 Linux 平台） |

### 编译命令

```bash
# Linux AMD64
go build -o agent-linux-amd64 cmd/server/main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o agent-linux-arm64 cmd/server/main.go
```

## 采集指标

node_exporter v1.5.0 默认启用以下 collector：

| Collector | 说明 |
|-----------|------|
| cpu | CPU 使用时间 |
| cpufreq | CPU 频率 |
| loadavg | 系统负载 |
| meminfo | 内存信息 |
| filesystem | 文件系统使用情况 |
| netdev | 网络设备统计 |
| stat | 内核/系统统计 |
| time | 系统时间 |
| uname | 系统信息 |
| ... | 更多 collector 详见 node_exporter 文档 |

## API 接口

### 1. 结构化指标

```
GET /api/v1/metrics/collect
```

从 node-exporter 采集的 Prometheus 指标中解析关键指标，返回 JSON 格式。

**解析的关键指标**:
- CPU（user/system/idle 时间、核心数）
- 内存（total/available/used/使用百分比）
- 磁盘（各挂载点容量信息）
- 网络（各网卡收发字节/数据包）
- 负载（1/5/15 分钟平均负载）

### 2. Prometheus 原始指标

```
GET /api/v1/metrics/prometheus
```

返回 node-exporter 采集的完整 Prometheus 格式指标文本，可直接被 Prometheus Server 抓取。

**示例输出**:
```
# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.
# TYPE node_cpu_seconds_total counter
node_cpu_seconds_total{cpu="0",mode="idle"} 12345.67
...
# HELP node_memory_MemTotal_bytes Total memory in bytes.
# TYPE node_memory_MemTotal_bytes gauge
node_memory_MemTotal_bytes 1.6777216e+10
...
```

## 自定义 Collector

如需禁用某些 collector 或启用额外的 collector，可修改 `collector_linux.go` 中的 `NewNodeCollector` 调用：

```go
// 只启用特定的 collector
nc, err := collector.NewNodeCollector(logger, "cpu", "meminfo", "filesystem")
```

更多 collector 名称可参考 node_exporter 源码中的 `registerCollector` 调用。

## 扩展建议

后续可增加以下功能：
- 指标缓存和定时采集（减少频繁调用 collector）
- 指标告警规则配置
- 历史指标存储（对接时序数据库如 InfluxDB）
- 自定义 collector 插件机制
