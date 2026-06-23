#!/bin/bash

# ============================================================
# 上传项目目录下的 agent 二进制到目标服务器
# 用法：
#   ./scripts/upload-agent.sh
#   SERVER_ADDR=192.168.1.10:8080 UPLOAD_DIR=/opt/myagent ./scripts/upload-agent.sh
# ============================================================

set -e

# 配置变量（可通过环境变量覆盖）
SERVER_ADDR="${SERVER_ADDR:-localhost:8080}"
UPLOAD_DIR="${UPLOAD_DIR:-/opt/myagent}"
FILE_NAME="${FILE_NAME:-agent}"
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
FILE_PATH="${PROJECT_DIR}/${FILE_NAME}"
BASE_URL="http://${SERVER_ADDR}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查本地文件是否存在
if [ ! -f "$FILE_PATH" ]; then
    echo -e "${RED}错误：文件不存在 ${FILE_PATH}${NC}"
    echo "请先构建 agent 二进制，例如："
    echo "  GOOS=linux GOARCH=amd64 go build -o agent ./cmd/server"
    exit 1
fi

echo "========================================"
echo "上传 agent 到目标服务器"
echo "========================================"
echo "服务器: ${SERVER_ADDR}"
echo "目标目录: ${UPLOAD_DIR}"
echo "本地文件: ${FILE_PATH}"
echo "文件大小: $(du -h "${FILE_PATH}" | cut -f1)"
echo "========================================"
echo ""

# 执行上传
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST \
    -F "dest_dir=${UPLOAD_DIR}" \
    -F "files=@${FILE_PATH};filename=${FILE_NAME}" \
    "${BASE_URL}/api/v1/upload/files")

http_code=$(echo "$response" | tail -n1 | sed 's/HTTP_CODE://')
body=$(echo "$response" | sed '$d')

if [ "$http_code" != "200" ]; then
    echo -e "${RED}上传失败，HTTP 状态码: ${http_code}${NC}"
    echo "$body"
    exit 1
fi

echo -e "${GREEN}上传成功${NC}"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
