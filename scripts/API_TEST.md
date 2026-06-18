# HTTP API 测试指南

## 快速测试方法

### 方法一：PowerShell 一键测试脚本（推荐）

```powershell
# 编译 + 启动服务 + 测试所有接口 + 自动停止
.\scripts\test_api.ps1

# 仅测试，不重新编译
.\scripts\test_api.ps1 -SkipBuild

# 测试指定容器ID
.\scripts\test_api.ps1 -ContainerId "abc123"

# 测试完成后保持服务运行
.\scripts\test_api.ps1 -KeepRunning
```

### 方法二：Go 单元测试（无需启动服务）

```bash
# 运行所有接口单元测试
go test ./internal/api/...

# 运行测试并显示详情
go test -v ./internal/api/...

# 运行测试并生成覆盖率报告
go test -v -cover ./internal/api/...

# 基准测试
go test -bench=. ./internal/api/...
```

### 方法三：curl 命令行测试

```bash
# 1. 健康检查
curl -s http://localhost:8080/health | jq .

# 2. 列出容器
curl -s http://localhost:8080/api/v1/docker/containers | jq .

# 3. 列出所有容器（含已停止）
curl -s "http://localhost:8080/api/v1/docker/containers?all=true" | jq .

# 4. 查看容器详情
curl -s http://localhost:8080/api/v1/docker/containers/CONTAINER_ID | jq .

# 5. 启动容器
curl -s -X POST http://localhost:8080/api/v1/docker/containers/CONTAINER_ID/start | jq .

# 6. 停止容器
curl -s -X POST http://localhost:8080/api/v1/docker/containers/CONTAINER_ID/stop | jq .

# 7. 停止容器（指定超时）
curl -s -X POST "http://localhost:8080/api/v1/docker/containers/CONTAINER_ID/stop?timeout=30" | jq .

# 8. 删除容器
curl -s -X DELETE http://localhost:8080/api/v1/docker/containers/CONTAINER_ID | jq .

# 9. 强制删除容器
curl -s -X DELETE "http://localhost:8080/api/v1/docker/containers/CONTAINER_ID?force=true" | jq .

# 10. 查看容器日志
curl -s http://localhost:8080/api/v1/docker/containers/CONTAINER_ID/logs | jq .

# 11. 查看容器日志（指定行数）
curl -s "http://localhost:8080/api/v1/docker/containers/CONTAINER_ID/logs?tail=50"

# 12. 收集系统指标
curl -s http://localhost:8080/api/v1/metrics/collect | jq .

# 13. 获取 Prometheus 格式指标
curl -s http://localhost:8080/api/v1/metrics/prometheus
```

### 方法四：IntelliJ IDEA / GoLand HTTP Client

1. 打开 `scripts/http-client.http` 文件
2. 点击左侧绿色运行按钮即可发送请求
3. 支持环境变量，可修改 `@baseUrl` 和 `@containerId`

### 方法五：Postman / Insomnia

导入以下 OpenAPI 风格定义：

```yaml
openapi: 3.0.0
info:
  title: Agent API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /health:
    get:
      summary: 健康检查
      responses:
        '200':
          description: OK
  /api/v1/docker/containers:
    get:
      summary: 列出容器
      parameters:
        - name: all
          in: query
          schema:
            type: boolean
      responses:
        '200':
          description: 容器列表
  /api/v1/docker/containers/{id}:
    get:
      summary: 查看容器详情
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: 容器详情
    delete:
      summary: 删除容器
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: force
          in: query
          schema:
            type: boolean
      responses:
        '200':
          description: 删除结果
  /api/v1/docker/containers/{id}/start:
    post:
      summary: 启动容器
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: 启动结果
  /api/v1/docker/containers/{id}/stop:
    post:
      summary: 停止容器
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: timeout
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: 停止结果
  /api/v1/docker/containers/{id}/logs:
    get:
      summary: 查看容器日志
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: tail
          in: query
          schema:
            type: string
            default: "100"
      responses:
        '200':
          description: 日志内容
  /api/v1/metrics/collect:
    get:
      summary: 收集系统指标
      responses:
        '200':
          description: 指标数据
  /api/v1/metrics/prometheus:
    get:
      summary: Prometheus 格式指标
      responses:
        '200':
          description: Prometheus 文本
```

## 测试文件说明

| 文件 | 用途 | 适用场景 |
|------|------|----------|
| `scripts/test_api.ps1` | PowerShell 集成测试 | Windows 环境一键测试 |
| `internal/api/router_test.go` | Go 单元测试 | CI/CD、开发时快速验证 |
| `scripts/http-client.http` | IDEA HTTP Client | IDE 内交互式测试 |
| `scripts/API_TEST.md` | 测试文档 | 查看 curl 命令和 Postman 配置 |
