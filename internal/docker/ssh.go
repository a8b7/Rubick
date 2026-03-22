package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"golang.org/x/crypto/ssh"
)

// SSHConnection SSH 隧道 Docker 连接
type SSHConnection struct {
	config    *ConnectionConfig
	client    *client.Client
	sshClient *ssh.Client
}

// NewSSHConnection 创建 SSH 连接
func NewSSHConnection(config *ConnectionConfig) *SSHConnection {
	if config.Port == 0 {
		config.Port = 2375 // 远程 Docker 端口
	}
	if config.SSHPort == 0 {
		config.SSHPort = 22
	}
	return &SSHConnection{
		config: config,
	}
}

// dockerVersionResponse Docker 版本响应
type dockerVersionResponse struct {
	APIVersion    string `json:"ApiVersion"`
	MinAPIVersion string `json:"MinAPIVersion"`
}

// Connect 建立连接
func (c *SSHConnection) Connect(ctx context.Context) (*client.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	// 建立 SSH 连接
	sshConfig := &ssh.ClientConfig{
		User:            c.config.SSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: 使用已知主机密钥
		Timeout:         30 * time.Second,
	}

	// 设置认证方式
	switch c.config.SSHAuthType {
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(c.config.SSHPrivateKey))
		if err != nil {
			return nil, fmt.Errorf("解析 SSH 私钥失败: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	case "password":
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(c.config.SSHPassword)}
	default:
		return nil, fmt.Errorf("不支持的 SSH 认证类型: %s", c.config.SSHAuthType)
	}

	// 连接 SSH 服务器
	sshAddr := fmt.Sprintf("%s:%d", c.config.Host, c.config.SSHPort)
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}
	c.sshClient = sshClient

	// 创建临时 transport 用于版本协商
	tempTransport := &sshTransport{
		sshClient:  sshClient,
		dockerPort: c.config.Port,
	}

	// 获取 Docker API 版本
	apiVersion, err := c.negotiateAPIVersion(tempTransport)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("协商 API 版本失败: %w", err)
	}

	// 创建自定义 HTTP 客户端，通过 SSH 隧道连接
	httpClient := &http.Client{
		Transport: tempTransport,
		Timeout:   120 * time.Second,
	}

	// 使用协商后的版本创建 client
	cli, err := client.NewClientWithOpts(
		client.WithHost("tcp://localhost"),
		client.WithVersion(apiVersion),
		client.WithHTTPClient(httpClient),
	)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("创建 SSH Docker 客户端失败: %w", err)
	}

	c.client = cli
	return c.client, nil
}

// negotiateAPIVersion 获取 Docker 服务器的 API 版本
func (c *SSHConnection) negotiateAPIVersion(transport *sshTransport) (string, error) {
	// 创建临时 HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// 发送 /version 请求
	req, err := http.NewRequest("GET", "http://localhost/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取版本失败，状态码: %d", resp.StatusCode)
	}

	var versionInfo dockerVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return "", fmt.Errorf("解析版本信息失败: %w", err)
	}

	// 如果获取不到版本，使用默认版本
	if versionInfo.APIVersion == "" {
		return "1.45", nil
	}

	return versionInfo.APIVersion, nil
}

// Close 关闭连接
func (c *SSHConnection) Close() error {
	var errs []error

	if c.client != nil {
		if err := c.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
	}
	return nil
}

// Type 返回连接类型
func (c *SSHConnection) Type() ConnectionType {
	return ConnectionTypeSSH
}

// Info 返回连接信息
func (c *SSHConnection) Info() map[string]string {
	return map[string]string{
		"type":      "ssh",
		"host":      c.config.Host,
		"ssh_port":  fmt.Sprintf("%d", c.config.SSHPort),
		"ssh_user":  c.config.SSHUser,
		"auth_type": c.config.SSHAuthType,
		"desc":      "SSH 隧道连接",
	}
}

// Test 测试连接是否可用
func (c *SSHConnection) Test(ctx context.Context) error {
	cli, err := c.Connect(ctx)
	if err != nil {
		return err
	}

	_, err = cli.Ping(ctx)
	return err
}

// sshTransport 自定义 HTTP Transport，使用 SSH 隧道
type sshTransport struct {
	sshClient  *ssh.Client
	dockerPort int
}

// sshConnReadCloser 包装 SSH 连接，确保 body 读取完成后关闭连接
type sshConnReadCloser struct {
	conn   net.Conn
	reader io.Reader
}

func (r *sshConnReadCloser) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *sshConnReadCloser) Close() error {
	// 先关闭读取端，再关闭连接
	if closer, ok := r.reader.(io.Closer); ok {
		closer.Close()
	}
	return r.conn.Close()
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *sshTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 通过 SSH 隧道建立 TCP 连接
	dockerAddr := fmt.Sprintf("127.0.0.1:%d", t.dockerPort)
	conn, err := t.sshClient.Dial("tcp", dockerAddr)
	if err != nil {
		return nil, fmt.Errorf("通过 SSH 隧道连接 Docker 失败: %w", err)
	}

	// 发送 HTTP 请求
	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}

	// 读取 HTTP 响应
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("读取 HTTP 响应失败: %w", err)
	}

	// 关键修复：包装 resp.Body，让它在关闭时同时关闭 SSH 连接
	// 这样响应体读取完成后，连接才会被释放
	resp.Body = &sshConnReadCloser{
		conn:   conn,
		reader: resp.Body,
	}

	return resp, nil
}

// CloseIdleConnections 关闭空闲连接
func (t *sshTransport) CloseIdleConnections() {
	// SSH 连接由 SSHConnection 管理
}
