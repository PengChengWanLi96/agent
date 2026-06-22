# API 接口文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`
- **响应格式**:

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 接口列表

### 健康检查

```
GET /health
```

**响应示例**:
```json
{
  "code": 0,
  "message": "ok"
}
```

---

## Docker 管理 API

### 1. 获取容器列表

```
GET /api/v1/docker/containers?all={true|false}
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| all | bool | 否 | 是否包含已停止的容器，默认 false |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "a1b2c3d4e5f6",
      "names": ["/my-container"],
      "image": "nginx:latest",
      "state": "running",
      "status": "Up 2 hours",
      "created": 1715769600
    }
  ]
}
```

### 2. 查看容器详情

```
GET /api/v1/docker/containers/:id
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "a1b2c3d4e5f6",
    "name": "/my-container",
    "image": "nginx:latest",
    "state": {
      "status": "running",
      "running": true,
      "paused": false,
      "restarting": false
    },
    "config": {
      "hostname": "my-container",
      "env": ["PATH=/usr/local/sbin:/usr/local/bin"],
      "cmd": ["nginx", "-g", "daemon off;"]
    }
  }
}
```

### 3. 启动容器

```
POST /api/v1/docker/containers/:id/start
```

**响应示例**:
```json
{
  "code": 0,
  "message": "container started"
}
```

### 4. 停止容器

```
POST /api/v1/docker/containers/:id/stop?timeout={seconds}
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| timeout | int | 否 | 停止超时时间（秒），默认 10 |

**响应示例**:
```json
{
  "code": 0,
  "message": "container stopped"
}
```

### 5. 删除容器

```
DELETE /api/v1/docker/containers/:id?force={true|false}
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| force | bool | 否 | 是否强制删除，默认 false |

**响应示例**:
```json
{
  "code": 0,
  "message": "container removed"
}
```

### 6. 查看容器日志

```
GET /api/v1/docker/containers/:id/logs?tail={number}
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| tail | string | 否 | 返回日志行数，默认 100 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": "192.168.1.1 - - [20/May/2026:10:00:00 +0000] GET / HTTP/1.1 200 ..."
}
```

---

## 系统指标 API

### 1. 采集结构化指标

```
GET /api/v1/metrics/collect
```

从 node-exporter 采集器获取关键指标，返回 JSON 格式。

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cpu": {
      "usage_percent": 0,
      "user_seconds": 12345.67,
      "system_seconds": 5678.90,
      "idle_seconds": 98765.43,
      "cores": 4
    },
    "memory": {
      "total_bytes": 16777216000,
      "available_bytes": 8388608000,
      "used_bytes": 8388608000,
      "free_bytes": 0,
      "usage_percent": 50.0
    },
    "disk": [
      {
        "filesystem": "/",
        "total_bytes": 107374182400,
        "used_bytes": 53687091200,
        "free_bytes": 53687091200,
        "usage_percent": 50.0
      }
    ],
    "network": [
      {
        "interface": "eth0",
        "receive_bytes": 123456789,
        "transmit_bytes": 98765432,
        "receive_packets": 12345,
        "transmit_packets": 9876
      }
    ],
    "load": {
      "load1": 0.5,
      "load5": 0.3,
      "load15": 0.2
    },
    "uptime_seconds": 0,
    "timestamp": "0001-01-01T00:00:00Z"
  }
}
```

**注意**: 此接口仅在 Linux 平台可用。其他平台返回错误。

### 2. 采集 Prometheus 原始指标

```
GET /api/v1/metrics/prometheus
```

返回 node-exporter 采集的原始 Prometheus 格式指标文本，可直接被 Prometheus Server 抓取。

**响应示例** (Content-Type: text/plain):
```
# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.
# TYPE node_cpu_seconds_total counter
node_cpu_seconds_total{cpu="0",mode="idle"} 12345.67
node_cpu_seconds_total{cpu="0",mode="user"} 5678.90
...
# HELP node_memory_MemTotal_bytes Total memory in bytes
# TYPE node_memory_MemTotal_bytes gauge
node_memory_MemTotal_bytes 1.6777216e+10
...
```

**注意**: 此接口仅在 Linux 平台可用。其他平台返回错误。

---

## SSH 文件管理 API

### 1. 建立 SSH 连接

```
POST /api/v1/ssh/connect
```

**请求体**:
```json
{
  "host": "192.168.1.100",
  "port": 22,
  "user": "root",
  "password": "your-password",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "a1b2c3d4...",
    "host": "192.168.1.100",
    "user": "root",
    "created_at": 1715769600
  }
}
```

### 2. 列出远程目录

```
GET /api/v1/ssh/sessions/:id/files?path=/root
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    { "name": "logs", "path": "/root/logs", "size": 0, "is_dir": true, "mode": "drwxr-xr-x", "mod_time": 1715769600 },
    { "name": "app.tar.gz", "path": "/root/app.tar.gz", "size": 2048, "is_dir": false, "mode": "-rw-r--r--", "mod_time": 1715769500 }
  ]
}
```

### 3. 下载文件

```
GET /api/v1/ssh/sessions/:id/download?path=/root/app.tar.gz
```

响应头 `Content-Disposition: attachment`，直接返回文件流。

### 4. 上传文件

```
POST /api/v1/ssh/sessions/:id/upload
Content-Type: multipart/form-data
```

**参数**:
| 参数 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| path | string | 是 | 远程目标路径，如 `/root/app.tar.gz` |
| file | file | 是 | 要上传的文件 |

### 5. 删除文件或目录

```
DELETE /api/v1/ssh/sessions/:id/files?path=/root/old.log
```

### 6. 创建目录

```
POST /api/v1/ssh/sessions/:id/mkdir?path=/root/newdir
```

### 7. 重命名

```
POST /api/v1/ssh/sessions/:id/rename
```

**请求体**:
```json
{
  "old_path": "/root/oldname",
  "new_path": "/root/newname"
}
```

### 8. 执行命令

```
POST /api/v1/ssh/sessions/:id/exec
GET  /api/v1/ssh/sessions/:id/exec?command=ls%20-la
```

**请求体 (POST)**:
```json
{
  "command": "ls -la"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "output": "total 32\ndrwxr-xr-x 3 root root 4096 ...",
    "exit_code": 0
  }
}
```

### 9. 关闭会话

```
DELETE /api/v1/ssh/sessions/:id
```

---

## 错误响应

所有接口在出错时返回如下格式：

```json
{
  "code": 500,
  "message": "error description"
}
```

常见错误码：
- `500` - 服务器内部错误（如 Docker 连接失败、采集器初始化失败）
- `404` - 资源不存在（Gin 默认返回）
