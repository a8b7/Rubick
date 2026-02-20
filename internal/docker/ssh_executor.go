package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

// SSHExecutor SSH 远程命令执行器
type SSHExecutor struct {
	config    *ConnectionConfig
	sshClient *ssh.Client
	tempDir   string
	mu        sync.Mutex
}

// NewSSHExecutor 创建 SSH 执行器
func NewSSHExecutor(config *ConnectionConfig) (*SSHExecutor, error) {
	if config.SSHPort == 0 {
		config.SSHPort = 22
	}

	return &SSHExecutor{
		config:  config,
		tempDir: "/tmp/rubick-compose",
	}, nil
}

// connect 建立 SSH 连接
func (e *SSHExecutor) connect(ctx context.Context) error {
	if e.sshClient != nil {
		return nil
	}

	sshConfig := &ssh.ClientConfig{
		User:            e.config.SSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	switch e.config.SSHAuthType {
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(e.config.SSHPrivateKey))
		if err != nil {
			return fmt.Errorf("解析 SSH 私钥失败: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	case "password":
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(e.config.SSHPassword)}
	default:
		return fmt.Errorf("不支持的 SSH 认证类型: %s", e.config.SSHAuthType)
	}

	sshAddr := fmt.Sprintf("%s:%d", e.config.Host, e.config.SSHPort)
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}

	e.sshClient = sshClient
	return nil
}

// Execute 执行命令并返回输出
func (e *SSHExecutor) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	if err := e.connect(ctx); err != nil {
		return nil, err
	}

	// 构建完整命令
	fullCmd := cmd
	for _, arg := range args {
		// 简单的参数转义
		if strings.Contains(arg, " ") || strings.Contains(arg, "\"") {
			fullCmd += fmt.Sprintf(" \"%s\"", strings.ReplaceAll(arg, "\"", "\\\""))
		} else {
			fullCmd += fmt.Sprintf(" %s", arg)
		}
	}

	session, err := e.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建 SSH session 失败: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(fullCmd)
	if err != nil {
		return nil, fmt.Errorf("命令执行失败: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ExecuteStream 执行命令并返回流式输出
func (e *SSHExecutor) ExecuteStream(ctx context.Context, cmd string, args ...string) (io.ReadCloser, error) {
	if err := e.connect(ctx); err != nil {
		return nil, err
	}

	// 构建完整命令
	fullCmd := cmd
	for _, arg := range args {
		if strings.Contains(arg, " ") || strings.Contains(arg, "\"") {
			fullCmd += fmt.Sprintf(" \"%s\"", strings.ReplaceAll(arg, "\"", "\\\""))
		} else {
			fullCmd += fmt.Sprintf(" %s", arg)
		}
	}

	session, err := e.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建 SSH session 失败: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("创建 stdout pipe 失败: %w", err)
	}
	session.Stderr = session.Stdout

	if err := session.Start(fullCmd); err != nil {
		session.Close()
		return nil, fmt.Errorf("启动命令失败: %w", err)
	}

	return &sshStreamReader{
		reader:  stdout,
		session: session,
	}, nil
}

// WriteFile 写入文件内容到远程临时位置
func (e *SSHExecutor) WriteFile(ctx context.Context, content string, filename string) (string, error) {
	if err := e.connect(ctx); err != nil {
		return "", err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 为每个请求创建独立的子目录
	requestDir := filepath.Join(e.tempDir, uuid.New().String())

	// 创建目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", requestDir)
	if _, err := e.Execute(ctx, "sh", "-c", mkdirCmd); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	remotePath := filepath.Join(requestDir, filename)

	// 使用 cat 写入文件
	session, err := e.sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH session 失败: %w", err)
	}
	defer session.Close()

	session.Stdin = bytes.NewBufferString(content)
	cmd := fmt.Sprintf("cat > %s", remotePath)
	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("写入远程文件失败: %w", err)
	}

	return remotePath, nil
}

// RemoveFile 删除远程文件（及其父目录）
func (e *SSHExecutor) RemoveFile(ctx context.Context, filepath string) error {
	// 删除包含文件的整个目录
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if dir == "" {
		_, err := e.Execute(ctx, "rm", "-f", filepath)
		return err
	}
	_, err := e.Execute(ctx, "rm", "-rf", dir)
	return err
}

// Close 关闭执行器
func (e *SSHExecutor) Close() error {
	if e.sshClient != nil {
		return e.sshClient.Close()
	}
	return nil
}

// ListDir 列出目录内容
func (e *SSHExecutor) ListDir(ctx context.Context, dirPath string) ([]FileInfo, error) {
	if err := e.connect(ctx); err != nil {
		return nil, err
	}

	// 验证路径安全性
	if err := validateSSHPath(dirPath); err != nil {
		return nil, err
	}

	// 使用 ls -la 命令获取目录内容
	cmd := fmt.Sprintf("ls -la %s 2>/dev/null || echo 'DIR_NOT_FOUND'", dirPath)
	output, err := e.Execute(ctx, "sh", "-c", cmd)
	if err != nil {
		return nil, fmt.Errorf("列出目录失败: %w", err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "DIR_NOT_FOUND") {
		return nil, fmt.Errorf("目录不存在")
	}

	// 解析 ls -la 输出
	var files []FileInfo
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total ") {
			continue
		}

		// 解析 ls -la 格式: drwxr-xr-x  2 user group 4096 Jan 01 12:00 dirname
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		perms := fields[0]
		isDir := strings.HasPrefix(perms, "d")
		size := int64(0)
		if len(fields) >= 5 {
			fmt.Sscanf(fields[4], "%d", &size)
		}

		// 文件名可能包含空格，从第 8 个字段开始
		name := strings.Join(fields[8:], " ")

		files = append(files, FileInfo{
			Name:  name,
			Path:  filepath.Join(dirPath, name),
			IsDir: isDir,
			Size:  size,
		})
	}

	return files, nil
}

// ReadFile 读取文件内容
func (e *SSHExecutor) ReadFile(ctx context.Context, filepath string) (string, error) {
	if err := e.connect(ctx); err != nil {
		return "", err
	}

	// 验证路径安全性
	if err := validateSSHPath(filepath); err != nil {
		return "", err
	}

	output, err := e.Execute(ctx, "cat", filepath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	return string(output), nil
}

// MkdirAll 创建目录（包括父目录）
func (e *SSHExecutor) MkdirAll(ctx context.Context, dirPath string) error {
	if err := e.connect(ctx); err != nil {
		return err
	}

	// 验证路径安全性
	if err := validateSSHPath(dirPath); err != nil {
		return err
	}

	_, err := e.Execute(ctx, "mkdir", "-p", dirPath)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return nil
}

// WriteFileToPath 写入文件到指定路径
func (e *SSHExecutor) WriteFileToPath(ctx context.Context, content string, filepath string) error {
	if err := e.connect(ctx); err != nil {
		return err
	}

	// 验证路径安全性
	if err := validateSSHPath(filepath); err != nil {
		return err
	}

	// 确保父目录存在
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if dir != "" {
		if _, err := e.Execute(ctx, "mkdir", "-p", dir); err != nil {
			return fmt.Errorf("创建父目录失败: %w", err)
		}
	}

	session, err := e.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH session 失败: %w", err)
	}
	defer session.Close()

	session.Stdin = bytes.NewBufferString(content)
	cmd := fmt.Sprintf("cat > %s", filepath)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// RemoveDir 删除目录及其内容
func (e *SSHExecutor) RemoveDir(ctx context.Context, dirPath string) error {
	if err := e.connect(ctx); err != nil {
		return err
	}

	// 验证路径安全性
	if err := validateSSHPath(dirPath); err != nil {
		return err
	}

	_, err := e.Execute(ctx, "rm", "-rf", dirPath)
	if err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}

	return nil
}

// FileExists 检查文件是否存在
func (e *SSHExecutor) FileExists(ctx context.Context, filepath string) (bool, error) {
	if err := e.connect(ctx); err != nil {
		return false, err
	}

	// 验证路径安全性
	if err := validateSSHPath(filepath); err != nil {
		return false, err
	}

	// 使用 test -e 命令检查文件是否存在
	_, err := e.Execute(ctx, "test", "-e", filepath)
	if err == nil {
		return true, nil
	}
	// test 命令失败意味着文件不存在
	return false, nil
}

// validateSSHPath 验证远程路径安全性（防止路径遍历攻击）
func validateSSHPath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("路径不能包含 '..'")
	}

	return nil
}

// sshStreamReader 包装 SSH 流读取器
type sshStreamReader struct {
	reader  io.Reader
	session *ssh.Session
	closed  bool
	closeMu sync.Mutex
}

func (r *sshStreamReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *sshStreamReader) Close() error {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil
	}

	r.session.Close()
	r.closed = true
	return nil
}
