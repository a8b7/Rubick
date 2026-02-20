package docker

import (
	"context"
	"io"
)

// FileInfo 文件信息
type FileInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// CommandExecutor 命令执行器接口
// 用于执行 docker compose 命令
type CommandExecutor interface {
	// Execute 执行命令并返回输出
	Execute(ctx context.Context, cmd string, args ...string) ([]byte, error)

	// ExecuteStream 执行命令并返回流式输出
	ExecuteStream(ctx context.Context, cmd string, args ...string) (io.ReadCloser, error)

	// WriteFile 写入文件内容到临时位置，返回文件路径
	WriteFile(ctx context.Context, content string, filename string) (string, error)

	// RemoveFile 删除文件
	RemoveFile(ctx context.Context, filepath string) error

	// Close 关闭执行器（清理资源）
	Close() error

	// ListDir 列出目录内容
	ListDir(ctx context.Context, dirPath string) ([]FileInfo, error)

	// ReadFile 读取文件内容
	ReadFile(ctx context.Context, filepath string) (string, error)

	// MkdirAll 创建目录（包括父目录）
	MkdirAll(ctx context.Context, dirPath string) error

	// WriteFileToPath 写入文件到指定路径
	WriteFileToPath(ctx context.Context, content string, filepath string) error

	// RemoveDir 删除目录及其内容
	RemoveDir(ctx context.Context, dirPath string) error

	// FileExists 检查文件是否存在
	FileExists(ctx context.Context, filepath string) (bool, error)
}
