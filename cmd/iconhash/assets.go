package main

import (
	"embed"
	"io/fs"
)

//go:embed assets
var embeddedFiles embed.FS

// GetAssetsFS 返回嵌入式文件系统的子集
func GetAssetsFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "assets")
}
