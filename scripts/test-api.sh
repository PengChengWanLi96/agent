#!/bin/bash

# ============================================================
# HTTP API 测试脚本 (curl 版本)
# 根据 scripts/http-client.http 生成
# ============================================================

# 配置
BASE_URL="${BASE_URL:-http://localhost:8080}"
CONTAINER_ID="${CONTAINER_ID:-test-container-id}"

echo "========================================"
echo "HTTP API 测试脚本"
echo "========================================"
echo "BASE_URL: $BASE_URL"
echo "CONTAINER_ID: $CONTAINER_ID"
echo "========================================"
echo ""

# ============================================================
# 1. 健康检查
echo "[1/14] 健康检查"
curl -s -X GET "$BASE_URL/health" | head -c 500
echo -e "\n"

# ============================================================
# 2. 列出所有容器
echo "[2/14] 列出所有容器"
curl -s -X GET "$BASE_URL/api/v1/docker/containers" | head -c 500
echo -e "\n"

# ============================================================
# 3. 列出所有容器（包含已停止）
echo "[3/14] 列出所有容器（包含已停止）"
curl -s -X GET "$BASE_URL/api/v1/docker/containers?all=true" | head -c 500
echo -e "\n"

# ============================================================
# 4. 查看容器详情
echo "[4/14] 查看容器详情"
curl -s -X GET "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID" | head -c 500
echo -e "\n"

# ============================================================
# 5. 启动容器
echo "[5/14] 启动容器"
curl -s -X POST "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/start" | head -c 500
echo -e "\n"

# ============================================================
# 6. 停止容器
echo "[6/14] 停止容器"
curl -s -X POST "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/stop" | head -c 500
echo -e "\n"

# ============================================================
# 7. 停止容器（指定超时时间）
echo "[7/14] 停止容器（指定超时时间）"
curl -s -X POST "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/stop?timeout=30" | head -c 500
echo -e "\n"

# ============================================================
# 8. 删除容器
echo "[8/14] 删除容器"
curl -s -X DELETE "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID" | head -c 500
echo -e "\n"

# ============================================================
# 9. 强制删除容器
echo "[9/14] 强制删除容器"
curl -s -X DELETE "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID?force=true" | head -c 500
echo -e "\n"

# ============================================================
# 10. 查看容器日志
echo "[10/14] 查看容器日志"
curl -s -X GET "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/logs" | head -c 1000
echo -e "\n"

# ============================================================
# 11. 查看容器日志（指定行数）
echo "[11/14] 查看容器日志（指定行数）"
curl -s -X GET "$BASE_URL/api/v1/docker/containers/$CONTAINER_ID/logs?tail=50" | head -c 1000
echo -e "\n"

# ============================================================
# 12. 收集系统指标
echo "[12/14] 收集系统指标"
curl -s -X GET "$BASE_URL/api/v1/metrics/collect" | head -c 500
echo -e "\n"

# ============================================================
# 13. 获取 Prometheus 格式指标
echo "[13/14] 获取 Prometheus 格式指标"
curl -s -X GET "$BASE_URL/api/v1/metrics/prometheus" | head -c 500
echo -e "\n"

# ============================================================
echo "========================================"
echo "测试完成"
echo "========================================"
