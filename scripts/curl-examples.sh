#!/usr/bin/env bash
# curl 命令示例
# 用法：source scripts/curl-examples.sh 或复制其中命令执行

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
CONTAINER_ID="${CONTAINER_ID:-test-container-id}"
IMAGE_ID="${IMAGE_ID:-test-image-id}"
UPLOAD_DIR="${UPLOAD_DIR:-/tmp/uploads}"

echo "当前 BASE_URL: $BASE_URL"
echo ""

# 1. 健康检查
curl -s "$BASE_URL/health" | python3 -m json.tool 2>/dev/null || true

# 2. 查看版本信息
curl -s "$BASE_URL/api/v1/version" | python3 -m json.tool 2>/dev/null || true

# 3. 列出运行中容器
curl -s "$BASE_URL/api/v1/docker/containers" | python3 -m json.tool 2>/dev/null || true

# 4. 列出所有容器（包含已停止）
curl -s "$BASE_URL/api/v1/docker/containers?all=true" | python3 -m json.tool 2>/dev/null || true

# 5. 查看容器详情
curl -s "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID" | python3 -m json.tool 2>/dev/null || true

# 6. 启动容器
curl -s -X POST "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/start" | python3 -m json.tool 2>/dev/null || true

# 7. 停止容器
curl -s -X POST "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/stop" | python3 -m json.tool 2>/dev/null || true

# 8. 停止容器（指定超时时间）
curl -s -X POST "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/stop?timeout=30" | python3 -m json.tool 2>/dev/null || true

# 9. 删除容器
curl -s -X DELETE "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID" | python3 -m json.tool 2>/dev/null || true

# 10. 强制删除容器
curl -s -X DELETE "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID?force=true" | python3 -m json.tool 2>/dev/null || true

# 11. 查看容器日志
curl -s "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/logs" | python3 -m json.tool 2>/dev/null || true

# 12. 查看容器日志（指定行数）
curl -s "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/logs?tail=50" | python3 -m json.tool 2>/dev/null || true

# 13. 列出镜像
curl -s "$BASE_URL/api/v1/docker/images" | python3 -m json.tool 2>/dev/null || true

# 14. 查看镜像详情
curl -s "$BASE_URL/api/v1/docker/images/$IMAGE_ID" | python3 -m json.tool 2>/dev/null || true

# 15. 收集系统指标
curl -s "$BASE_URL/api/v1/metrics/collect" | python3 -m json.tool 2>/dev/null || true

# 16. 获取 Prometheus 格式指标
curl -s "$BASE_URL/api/v1/metrics/prometheus"

# 17. 上传文件（需要本地存在 /tmp/test.txt 或替换路径）
if [ -f /tmp/test.txt ]; then
  curl -s -X POST "$BASE_URL/api/v1/upload/files" \
    -F "dest_dir=$UPLOAD_DIR" \
    -F "files=@/tmp/test.txt" | python3 -m json.tool 2>/dev/null || true
else
  echo "跳过上传示例：/tmp/test.txt 不存在，可执行 echo 'hello' > /tmp/test.txt 后重试"
fi

# 18. 查看上传目录列表
curl -s "$BASE_URL/api/v1/upload/files?dest_dir=$UPLOAD_DIR" | python3 -m json.tool 2>/dev/null || true
