# SSH 文件管理功能开发记录

## 需求概述

在 Agent 管理控制台中新增 **SSH 文件管理** 功能：点击工具栏按钮后，弹出 SSH 连接面板；连接成功后，页面分为左右两部分：

- **左侧**：基于 SFTP 的远程文件查看与管理（浏览、上传、下载、新建目录、重命名、删除）。
- **右侧**：基于 SSH 的命令行终端，可输入命令并实时查看输出。

## 技术方案

### 后端

| 层级 | 文件 | 职责 |
|------|------|------|
| Client | `internal/client/ssh/client.go` | 封装 `golang.org/x/crypto/ssh` + `github.com/pkg/sftp`，提供连接、SFTP 文件操作、命令执行 |
| Service | `internal/service/ssh_service.go` | 管理 SSH 会话生命周期（内存会话表），转发操作到 Client |
| Handler | `internal/api/ssh_handler.go` | Gin HTTP 接口：连接、文件列表、上传/下载、目录操作、命令执行、断开连接 |
| Model | `internal/model/model.go` | SSH 相关的请求/响应结构体 |
| Router | `internal/api/router.go` | 注册 `/api/v1/ssh/*` 路由 |
| Main | `cmd/server/main.go` | 初始化 `SSHService` 并注入 Router |

### 前端

- 在 `web/index.html` 中新增：
  - 工具栏按钮 **SSH 文件管理**
  - SSH 连接表单（主机、端口、用户名、密码/私钥）
  - 左右分栏工作区：左侧 SFTP 文件管理器、右侧命令行终端
  - 模态框：新建目录、重命名
- 纯原生 HTML/CSS/JS，无外部 CDN 依赖，随二进制一起嵌入。

## 关键 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/ssh/connect` | 建立 SSH 连接，返回 session id |
| GET | `/api/v1/ssh/sessions` | 列出当前会话 |
| DELETE | `/api/v1/ssh/sessions/:id` | 关闭指定会话 |
| GET | `/api/v1/ssh/sessions/:id/files?path=/` | 列出远程目录 |
| GET | `/api/v1/ssh/sessions/:id/download?path=/file` | 下载远程文件 |
| POST | `/api/v1/ssh/sessions/:id/upload` | 上传文件（form: `path`, `file`） |
| DELETE | `/api/v1/ssh/sessions/:id/files?path=/file` | 删除文件或目录 |
| POST | `/api/v1/ssh/sessions/:id/mkdir?path=/dir` | 创建目录 |
| POST | `/api/v1/ssh/sessions/:id/rename` | 重命名（body: `old_path`, `new_path`） |
| POST / GET | `/api/v1/ssh/sessions/:id/exec` | 执行命令 |

## 开发步骤

1. **依赖安装**
   ```bash
   go get github.com/pkg/sftp@v1.13.6
   go get github.com/gorilla/websocket@v1.5.3
   go get golang.org/x/crypto@v0.32.0
   ```

2. **新增 SSH Client**
   - 实现 `NewClient`，支持密码或私钥认证。
   - 同时创建 `ssh.Client` 和 `sftp.Client`。
   - 实现 `ListDir`、`Download`、`Upload`、`Remove`、`Mkdir`、`Rename`、`Exec`。

3. **新增 SSH Service**
   - 使用内存 `map[string]*SSHSession` 管理会话。
   - 使用 `crypto/rand` 生成会话 ID，避免引入额外 UUID 库。
   - 提供 `Connect`、`CloseSession`、`GetSession`、`ListSessions` 及各类文件/命令操作代理。

4. **新增 Handler 与 Model**
   - 在 `model.go` 增加 `SSHConnectRequest`、`SSHSessionResponse`、`SSHFileInfo`、`SSHExecRequest`、`SSHExecResponse`、`SSHRenameRequest`。
   - 在 `ssh_handler.go` 实现 REST 接口，统一返回 `model.Response`。

5. **路由与入口集成**
   - `NewRouter` 增加 `sshSvc *service.SSHService` 参数。
   - 在 `/api/v1/ssh` 下注册所有路由。
   - `cmd/server/main.go` 创建 `service.NewSSHService()` 并传入 Router。
   - 更新 `router_test.go` 以适配新的 Router 签名，并补充 SSH 相关基础测试。

6. **前端实现**
   - 在工具栏添加 **SSH 文件管理** 按钮。
   - 添加连接表单，验证必填项。
   - 连接成功后隐藏表单，显示左右分栏工作区。
   - 左侧文件管理器：面包屑导航、文件列表、返回上级、新建目录、上传、下载、重命名、删除。
   - 右侧终端：命令历史滚动显示、命令输入框、通过 GET `/exec?command=...` 执行命令。

7. **测试与验证**
   - `go test ./...` 全部通过。
   - `go build -o agent ./cmd/server` 编译成功。

## 运行效果

1. 启动服务：
   ```bash
   go run ./cmd/server
   ```
2. 浏览器访问 `http://localhost:8080`。
3. 点击 **SSH 文件管理**，输入主机、用户名、密码或私钥，点击连接。
4. 左侧浏览远程文件系统，右侧输入命令执行。

## 注意事项

- 当前实现使用 `ssh.InsecureIgnoreHostKey()` 忽略主机密钥验证，适合内部可信环境；生产环境建议改为固定主机密钥验证。
- 会话保存在内存中，服务重启后失效。
- 命令行终端基于 HTTP 轮询（每次回车发送一次请求），非交互式长连接；适用于查看输出、执行管理命令。
- 文件上传/下载通过 HTTP 直接传输，大文件请留意超时与内存占用。

## 后续可优化项

- 增加 WebSocket 终端，支持交互式命令与实时输出。
- 增加会话过期清理与并发数限制。
- 支持主机密钥指纹校验与 KnownHosts 配置。
- 支持拖拽上传、文件内容预览。
