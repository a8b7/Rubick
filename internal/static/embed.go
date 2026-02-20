package static

import (
	"embed"
	"io/fs"
)

//go:embed dist logo.png logo.svg docker-icon.png
var staticFS embed.FS

// GetDistFS 获取 dist 文件系统
func GetDistFS() (fs.FS, error) {
	return fs.Sub(staticFS, "dist")
}

// GetIndexHTML 获取 index.html 内容
func GetIndexHTML() ([]byte, error) {
	return staticFS.ReadFile("dist/index.html")
}
