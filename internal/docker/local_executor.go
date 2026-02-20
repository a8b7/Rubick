package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// LocalExecutor 本地命令执行器
type LocalExecutor struct {
	tempDir string
	mu      sync.Mutex
}

// NewLocalExecutor 创建本地执行器
func NewLocalExecutor() (*LocalExecutor, error) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "rubick-compose")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	return &LocalExecutor{
		tempDir: tempDir,
	}, nil
}

// Execute 执行命令并返回输出
func (e *LocalExecutor) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, cmd, args...)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("命令执行失败: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ExecuteStream 执行命令并返回流式输出
func (e *LocalExecutor) ExecuteStream(ctx context.Context, cmd string, args ...string) (io.ReadCloser, error) {
	command := exec.CommandContext(ctx, cmd, args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdout pipe 失败: %w", err)
	}
	command.Stderr = command.Stdout // 合并 stderr 到 stdout

	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("启动命令失败: %w", err)
	}

	// 返回包装的 ReadCloser，在读取完成后等待命令结束
	return &streamReader{
		ReadCloser: stdout,
		cmd:        command,
	}, nil
}

// WriteFile 写入文件内容到临时位置
func (e *LocalExecutor) WriteFile(ctx context.Context, content string, filename string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 为每个请求创建独立的子目录
	requestDir := filepath.Join(e.tempDir, uuid.New().String())
	if err := os.MkdirAll(requestDir, 0755); err != nil {
		return "", fmt.Errorf("创建请求目录失败: %w", err)
	}

	filePath := filepath.Join(requestDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return filePath, nil
}

// RemoveFile 删除文件（及其父目录）
func (e *LocalExecutor) RemoveFile(ctx context.Context, filepath string) error {
	// 删除包含文件的整个目录
	dir := filepath[:strings.LastIndex(filepath, string(os.PathSeparator))]
	if dir == "" {
		return os.Remove(filepath)
	}
	return os.RemoveAll(dir)
}

// Close 关闭执行器
func (e *LocalExecutor) Close() error {
	// 清理所有临时文件
	return os.RemoveAll(e.tempDir)
}

// ListDir 列出目录内容
func (e *LocalExecutor) ListDir(ctx context.Context, dirPath string) ([]FileInfo, error) {
	// 验证路径安全性
	if err := validatePath(dirPath); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法读取的文件
		}

		files = append(files, FileInfo{
			Name:  entry.Name(),
			Path:  filepath.Join(dirPath, entry.Name()),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	return files, nil
}

// ReadFile 读取文件内容
func (e *LocalExecutor) ReadFile(ctx context.Context, filepath string) (string, error) {
	// 验证路径安全性
	if err := validatePath(filepath); err != nil {
		return "", err
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	return string(content), nil
}

// MkdirAll 创建目录（包括父目录）
func (e *LocalExecutor) MkdirAll(ctx context.Context, dirPath string) error {
	// 验证路径安全性
	if err := validatePath(dirPath); err != nil {
		return err
	}

	return os.MkdirAll(dirPath, 0755)
}

// WriteFileToPath 写入文件到指定路径
func (e *LocalExecutor) WriteFileToPath(ctx context.Context, content string, filepath string) error {
	// 验证路径安全性
	if err := validatePath(filepath); err != nil {
		return err
	}

	// 确保父目录存在
	dir := filepath[:strings.LastIndex(filepath, string(os.PathSeparator))]
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建父目录失败: %w", err)
		}
	}

	return os.WriteFile(filepath, []byte(content), 0644)
}

// RemoveDir 删除目录及其内容
func (e *LocalExecutor) RemoveDir(ctx context.Context, dirPath string) error {
	// 验证路径安全性
	if err := validatePath(dirPath); err != nil {
		return err
	}

	return os.RemoveAll(dirPath)
}

// FileExists 检查文件是否存在
func (e *LocalExecutor) FileExists(ctx context.Context, filepath string) (bool, error) {
	// 验证路径安全性
	if err := validatePath(filepath); err != nil {
		return false, err
	}

	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// validatePath 验证路径安全性（防止路径遍历攻击）
func validatePath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("路径不能包含 '..'")
	}

	return nil
}

// streamReader 包装 ReadCloser，在关闭时等待命令结束
type streamReader struct {
	io.ReadCloser
	cmd     *exec.Cmd
	closed  bool
	closeMu sync.Mutex
}

func (r *streamReader) Close() error {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil
	}

	err := r.ReadCloser.Close()
	r.cmd.Wait()
	r.closed = true
	return err
}
