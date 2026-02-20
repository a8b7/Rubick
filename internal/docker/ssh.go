package docker

import (
	"bufio"
	"context"
	"fmt"
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

	// 创建自定义 HTTP 客户端，通过 SSH 隧道连接
	httpClient := &http.Client{
		Transport: &sshTransport{
			sshClient:  sshClient,
			dockerPort: c.config.Port,
		},
		Timeout: 120 * time.Second,
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost("tcp://localhost"),
		client.WithAPIVersionNegotiation(),
		client.WithHTTPClient(httpClient),
	)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("创建 SSH Docker 客户端失败: %w", err)
	}

	c.client = cli
	return c.client, nil
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

// RoundTrip 实现 http.RoundTripper 接口
func (t *sshTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 通过 SSH 隧道建立 TCP 连接
	dockerAddr := fmt.Sprintf("127.0.0.1:%d", t.dockerPort)
	conn, err := t.sshClient.Dial("tcp", dockerAddr)
	if err != nil {
		return nil, fmt.Errorf("通过 SSH 隧道连接 Docker 失败: %w", err)
	}
	defer conn.Close()

	// 发送 HTTP 请求
	if err := req.Write(conn); err != nil {
		return nil, fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}

	// 读取 HTTP 响应
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, fmt.Errorf("读取 HTTP 响应失败: %w", err)
	}

	return resp, nil
}

// CloseIdleConnections 关闭空闲连接
func (t *sshTransport) CloseIdleConnections() {
	// SSH 连接由 SSHConnection 管理
}
