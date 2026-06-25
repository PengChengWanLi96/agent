# 构建指南

本文档介绍如何使用项目根目录下的 `build.sh` 构建 Agent 二进制文件。

## 前置条件

- 安装 [Go 1.24+](https://go.dev/dl/)
- 安装 Git，并将项目克隆到本地
- Linux / macOS / Windows Git Bash：直接运行 `./build.sh`
- Windows PowerShell / cmd：建议安装 Git Bash，或参考文末手动命令

## 可用构建目标

| 目标 | 说明 |
|------|------|
| `./build.sh build` | 为当前平台构建单个二进制文件到 `dist/agent`（Windows 为 `dist/agent.exe`） |
| `./build.sh all` | 交叉编译所有支持的平台，输出到 `dist/` |
| `./build.sh run` | 本地直接运行（用于开发调试） |
| `./build.sh clean` | 删除 `dist/` 目录 |
| `./build.sh version` | 打印当前构建版本信息 |

默认目标为 `build`，直接执行 `./build.sh` 等价于 `./build.sh build`。

## 快速开始

```bash
# 为当前平台构建
cd agent
./build.sh build

# 查看版本信息
./dist/agent --version

# 本地运行（开发调试）
./build.sh run
```

## 交叉编译

执行 `./build.sh all` 会构建以下平台产物：

| 操作系统 | 架构 | 输出文件名 |
|----------|------|------------|
| Linux | amd64 | `dist/agent-{VERSION}-linux-amd64` |
| Linux | arm64 | `dist/agent-{VERSION}-linux-arm64` |
| Windows | amd64 | `dist/agent-{VERSION}-windows-amd64.exe` |
| macOS | amd64 | `dist/agent-{VERSION}-darwin-amd64` |
| macOS | arm64 | `dist/agent-{VERSION}-darwin-arm64` |

示例：

```bash
./build.sh all

# 指定版本号
VERSION=v1.2.0 ./build.sh all
```

> 说明：`linux/amd64` 对应常见的 Linux x86_64 服务器；`linux/arm64` 对应 ARM 服务器或云实例。

## 自定义构建变量

可在命令行覆盖以下变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VERSION` | Git 标签或 `dev` | 产品版本号 |
| `COMMIT` | 当前 Git 短 commit ID | Git 提交 ID |
| `GITTIME` | 最近 commit 的提交时间 | Git 提交时间 |
| `BUILDDATE` | 当前 UTC 时间 | 二进制构建时间 |
| `APP` | `agent` | 输出二进制名称 |
| `OUTPUT_DIR` | `dist` | 输出目录 |
| `GOFLAGS` | （空） | 额外传递给 `go build` 的 flags |

示例：

```bash
# 指定版本号和输出目录
VERSION=v1.2.0 OUTPUT_DIR=./release ./build.sh build

# 添加额外的 go build 参数
GOFLAGS="-trimpath" ./build.sh all
```

## 版本信息注入

构建时会自动通过 `-ldflags` 将版本信息注入到 `agent/internal/version` 包中：

- `Version`：产品版本号
- `GitCommit`：Git 短 commit ID
- `GitTime`：Git 提交时间（UTC）
- `BuildDate`：构建时间（UTC）

构建完成后可通过以下命令验证：

```bash
./dist/agent --version
```

输出示例：

```
agent v1.2.0
  git commit: a1b2c3d
  git time:   2026-06-18T11:40:32Z
  build date: 2026-06-18T11:45:00Z
  go version: go1.24.3
  platform:   linux/amd64
```

## 使用 Makefile（可选）

如果系统已安装 `make`，Makefile 仍是 `build.sh` 的薄封装，命令保持不变：

```bash
make build
make all
make run
make clean
make version
```

> 注意：Windows 默认不自带 `make`，推荐使用 `./build.sh` 直接构建。

## Windows 环境注意事项

### 使用 Git Bash

Git Bash 中可以直接使用与 Linux 相同的命令：

```bash
./build.sh build
./build.sh all
```

### 使用 PowerShell

如果系统没有 Git Bash，可以直接使用等价的 PowerShell 命令：

```powershell
$ver = "v1.2.0"
$commit = git rev-parse --short HEAD
$gitTime = git log -1 --format=%cI --date=iso-strict
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

# 当前平台
$env:GOFLAGS=""
go build -ldflags "-X agent/internal/version.Version=$ver -X agent/internal/version.GitCommit=$commit -X agent/internal/version.GitTime=$gitTime -X agent/internal/version.BuildDate=$buildDate" -o dist/agent.exe ./cmd/server

# 交叉编译 Linux x86_64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags "-X agent/internal/version.Version=$ver -X agent/internal/version.GitCommit=$commit -X agent/internal/version.GitTime=$gitTime -X agent/internal/version.BuildDate=$buildDate" -o dist/agent-linux-amd64 ./cmd/server
```

### 使用 cmd.exe

```bat
set VERSION=v1.2.0
for /f "tokens=*" %%a in ('git rev-parse --short HEAD') do set COMMIT=%%a
for /f "tokens=*" %%a in ('git log -1 --format=^%cI') do set GITTIME=%%a
for /f "tokens=*" %%a in ('powershell -Command "Get-Date.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')"') do set BUILDDATE=%%a

go build -ldflags "-X agent/internal/version.Version=%VERSION% -X agent/internal/version.GitCommit=%COMMIT% -X agent/internal/version.GitTime=%GITTIME% -X agent/internal/version.BuildDate=%BUILDDATE%" -o dist\agent.exe ./cmd/server
```

## 手动 go build（不推荐用于产品包）

如果不方便使用 `build.sh`，也可以直接执行 `go build`：

```bash
# 当前平台
go build -ldflags "\
  -X agent/internal/version.Version=v1.2.0 \
  -X agent/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X agent/internal/version.GitTime=$(git log -1 --format=%cI --date=iso-strict) \
  -X agent/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o agent ./cmd/server
```

> 注意：直接通过文件路径构建（如 `go build -o agent cmd/server/main.go`）不会携带 Git VCS 信息；推荐使用包路径 `./cmd/server`。
