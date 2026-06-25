#!/usr/bin/env bash
set -euo pipefail

# Build script for agent
#
# Usage:
#   ./build.sh build      # build for current platform
#   ./build.sh all        # cross-compile for all supported platforms
#   ./build.sh run        # run locally for development
#   ./build.sh clean      # remove dist/ directory
#   ./build.sh version    # print build metadata
#
# Override variables:
#   VERSION=v1.3.0 ./build.sh all
#   APP=my-agent OUTPUT_DIR=./bin ./build.sh build

# Project metadata
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
GITTIME="${GITTIME:-$(git log -1 --format=%cI --date=iso-strict 2>/dev/null || echo unknown)}"
BUILDDATE="${BUILDDATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)}"

# Build configuration
APP="${APP:-agent}"
CMD="${CMD:-./cmd/server}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"
GOFLAGS="${GOFLAGS:-}"
LDFLAGS="${LDFLAGS:- \
  -X agent/internal/version.Version=${VERSION} \
  -X agent/internal/version.GitCommit=${COMMIT} \
  -X agent/internal/version.GitTime=${GITTIME} \
  -X agent/internal/version.BuildDate=${BUILDDATE} \
}"

# Platforms to cross-compile: os/arch
PLATFORMS=(
  linux/amd64
  linux/arm64
  windows/amd64
  darwin/amd64
  darwin/arm64
)

build() {
  mkdir -p "${OUTPUT_DIR}"
  go build ${GOFLAGS} -ldflags "${LDFLAGS}" -o "${OUTPUT_DIR}/${APP}" "${CMD}"
  echo "Built: ${OUTPUT_DIR}/${APP}"
}

run() {
  go run ${GOFLAGS} -ldflags "${LDFLAGS}" "${CMD}"
}

build_all() {
  clean
  mkdir -p "${OUTPUT_DIR}"
  for p in "${PLATFORMS[@]}"; do
    os="${p%%/*}"
    arch="${p##*/}"
    suffix=""
    if [ "${os}" = "windows" ]; then
      suffix=".exe"
    fi
    out="${OUTPUT_DIR}/${APP}-${VERSION}-${os}-${arch}${suffix}"
    echo "Building ${out} ..."
    GOOS="${os}" GOARCH="${arch}" CGO_ENABLED=0 go build ${GOFLAGS} \
      -ldflags "${LDFLAGS}" \
      -o "${out}" "${CMD}"
  done
  echo "All binaries built in ${OUTPUT_DIR}/"
}

clean() {
  rm -rf "${OUTPUT_DIR}"
  echo "Cleaned ${OUTPUT_DIR}/"
}

version() {
  echo "Version:    ${VERSION}"
  echo "Git commit: ${COMMIT}"
  echo "Git time:   ${GITTIME}"
  echo "Build date: ${BUILDDATE}"
}

usage() {
  echo "Usage: $0 {build|all|run|clean|version}"
  exit 1
}

main() {
  local cmd="${1:-build}"
  case "${cmd}" in
    build)
      build
      ;;
    all)
      build_all
      ;;
    run)
      run
      ;;
    clean)
      clean
      ;;
    version)
      version
      ;;
    *)
      usage
      ;;
  esac
}

main "$@"
