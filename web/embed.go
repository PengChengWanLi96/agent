package web

import "embed"

// Content 包含前端静态资源，编译时会被打包到二进制中。
//
//go:embed index.html
var Content embed.FS
