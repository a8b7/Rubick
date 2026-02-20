package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// TCPConnection TCP+TLS Docker 连接
type TCPConnection struct {
	config *ConnectionConfig
	client *client.Client
}

// NewTCPConnection 创建 TCP 连接
func NewTCPConnection(config *ConnectionConfig) *TCPConnection {
	if config.Port == 0 {
		config.Port = 2376 // 默认 TLS 端口
	}
	return &TCPConnection{
		config: config,
	}
}

// Connect 建立连接
func (c *TCPConnection) Connect(ctx context.Context) (*client.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	hostURL := fmt.Sprintf("tcp://%s:%d", c.config.Host, c.config.Port)

	opts := []client.Opt{
		client.WithHost(hostURL),
		client.WithAPIVersionNegotiation(),
	}

	// TLS 配置
	if !c.config.SkipTLSVerify {
		// TODO: 加载 TLS 证书
		// 使用 TLSCertID 从数据库加载证书
		opts = append(opts, client.WithTLSClientConfigFromEnv())
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 TCP Docker 客户端失败: %w", err)
	}

	c.client = cli
	return c.client, nil
}

// Close 关闭连接
func (c *TCPConnection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Type 返回连接类型
func (c *TCPConnection) Type() ConnectionType {
	return ConnectionTypeTCP
}

// Info 返回连接信息
func (c *TCPConnection) Info() map[string]string {
	return map[string]string{
		"type":    "tcp",
		"host":    c.config.Host,
		"port":    fmt.Sprintf("%d", c.config.Port),
		"tls":     fmt.Sprintf("%v", !c.config.SkipTLSVerify),
		"desc":    "远程 TCP 连接",
	}
}

// Test 测试连接是否可用
func (c *TCPConnection) Test(ctx context.Context) error {
	cli, err := c.Connect(ctx)
	if err != nil {
		return err
	}

	_, err = cli.Ping(ctx)
	return err
}
