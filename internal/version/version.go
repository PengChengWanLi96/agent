package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// 这些变量可在编译时通过 -ldflags 注入，例如：
//
//	go build -ldflags "-X agent/internal/version.Version=v1.2.0 \
//	    -X agent/internal/version.GitCommit=$(git rev-parse --short HEAD) \
//	    -X agent/internal/version.GitTime=$(git log -1 --format=%cI --date=iso-strict) \
//	    -X agent/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//	    -o agent ./cmd/server
//
// 当使用裸的 go build 时，若模块位于 Git 仓库中，Go 会自动把 VCS 信息写入
// 二进制（要求 Go 1.18+）。本包统一将 Git 时间与构建时间以 UTC 格式输出。
var (
	Version   = "dev"
	GitCommit = "unknown"
	GitTime   = "unknown"
	BuildDate = "unknown"
)

func normalizeUTC(s string) string {
	if s == "" || s == "unknown" {
		return s
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format(time.RFC3339)
}

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		var (
			revision string
			vcsTime  string
		)
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.time":
				vcsTime = s.Value
			}
		}
		if GitCommit == "unknown" && revision != "" {
			if len(revision) > 12 {
				GitCommit = revision[:12]
			} else {
				GitCommit = revision
			}
		}
		if vcsTime != "" {
			t, err := time.Parse(time.RFC3339, vcsTime)
			if err == nil {
				if GitTime == "unknown" {
					GitTime = t.UTC().Format(time.RFC3339)
				}
				if BuildDate == "unknown" {
					BuildDate = t.UTC().Format(time.RFC3339)
				}
			} else {
				if GitTime == "unknown" {
					GitTime = vcsTime
				}
				if BuildDate == "unknown" {
					BuildDate = vcsTime
				}
			}
		}
	}

	GitTime = normalizeUTC(GitTime)
	BuildDate = normalizeUTC(BuildDate)
}

// Info 包含构建与运行时的版本信息。
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitTime   string `json:"git_time"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetInfo 返回结构化版本信息。
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitTime:   GitTime,
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
		"agent %s\n  git commit: %s\n  git time:   %s\n  build date: %s\n  go version: %s\n  platform:   %s/%s",
		Version, GitCommit, GitTime, BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}
