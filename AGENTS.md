# AGENTS.md

## Project

Go 1.24 service agent exposing REST APIs for Docker container management and Linux system metrics collection.

## Entrypoint

- `cmd/server/main.go` — single binary, initializes Docker client, metrics collector, Gin router, and HTTP server

## Build & Run

```bash
# Build
 go build -o agent cmd/server/main.go

# Run (requires Docker socket or remote Docker host)
 go run cmd/server/main.go

# Cross-compile for Linux (metrics only work on Linux)
 GOOS=linux GOARCH=amd64 go build -o agent cmd/server/main.go
```

## Test

```bash
# Run all tests
 go test ./...

# Run a single test
 go test ./internal/api -run TestHealthEndpoint

# Run benchmarks
 go test ./internal/api -bench=.
```

Tests that call Docker or metrics APIs expect 500 when Docker is unavailable or on non-Linux platforms. This is expected.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDR` | `:8080` | HTTP listen address |
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon address (also supports `tcp://host:port`) |
| `DOCKER_API_VERSION` | `v1.43` | Docker API version |
| `DOCKER_TLS_VERIFY` | *(empty)* | Enable TLS if set to any non-empty value |
| `DOCKER_CERT_PATH` | *(empty)* | Path to TLS certs (cert.pem, key.pem, ca.pem) |

## Architecture

```
cmd/server/main.go      → entrypoint
internal/api/           → Gin handlers + router
internal/service/       → business logic (thin wrappers)
internal/client/docker/ → raw Docker HTTP API client
internal/client/metrics/→ node-exporter collector
internal/config/        → env-based config
internal/model/         → shared structs
```

## Platform-specific Code

- `internal/client/metrics/collector_linux.go` — real node-exporter collector (build tag `linux`)
- `internal/client/metrics/collector_stub.go` — no-op stub returning error (build tag `!linux`)

Metrics endpoints (`/api/v1/metrics/collect`, `/api/v1/metrics/prometheus`) return 500 on non-Linux platforms.

## Adding a New Feature

1. Add client in `internal/client/<module>/`
2. Add service in `internal/service/<module>_service.go`
3. Add handler + routes in `internal/api/<module>_handler.go`
4. Wire in `cmd/server/main.go`

## API Quick Reference

- `GET /health` — health check
- `GET /api/v1/docker/containers?all=true` — list containers
- `GET /api/v1/docker/containers/:id` — inspect container
- `POST /api/v1/docker/containers/:id/start` — start container
- `POST /api/v1/docker/containers/:id/stop?timeout=10` — stop container
- `DELETE /api/v1/docker/containers/:id?force=true` — remove container
- `GET /api/v1/docker/containers/:id/logs?tail=100` — container logs
- `GET /api/v1/metrics/collect` — structured system metrics (Linux only)
- `GET /api/v1/metrics/prometheus` — raw Prometheus text (Linux only)
