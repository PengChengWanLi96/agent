package version

import (
	"fmt"
	"runtime"
)

// 这些变量可在编译时通过 -ldflags 注入，例如：
//
//	go build -ldflags "-X agent/internal/version.Version=v1.2.0 \
//	    -X agent/internal/version.GitCommit=$(git rev-parse --short HEAD) \
//	    -X agent/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//	    -o agent cmd/server/main.go
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// Info 包含构建与运行时的版本信息。
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetInfo 返回结构化版本信息。
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// Short 返回简短版本号。
func Short() string {
	return Version
}

// String 返回包含构建信息的多行版本描述。
func String() string {
	return fmt.Sprintf(
		"agent %s\n  git commit: %s\n  build date: %s\n  go version: %s\n  platform:   %s/%s",
		Version, GitCommit, BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}
