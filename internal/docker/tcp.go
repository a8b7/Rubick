package docker

import (
	"context"
	"fmt"
	"time"

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

	// 先创建临时 client 用于版本协商
	tempOpts := []client.Opt{
		client.WithHost(hostURL),
	}

	// TLS 配置
	if !c.config.SkipTLSVerify {
		tempOpts = append(tempOpts, client.WithTLSClientConfigFromEnv())
	}

	tempCli, err := client.NewClientWithOpts(tempOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建临时 Docker 客户端失败: %w", err)
	}

	// 协商 API 版本
	apiVersion := c.negotiateAPIVersion(tempCli)
	tempCli.Close()

	// 使用协商后的版本创建正式 client
	opts := []client.Opt{
		client.WithHost(hostURL),
		client.WithVersion(apiVersion),
	}

	if !c.config.SkipTLSVerify {
		opts = append(opts, client.WithTLSClientConfigFromEnv())
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 TCP Docker 客户端失败: %w", err)
	}

	c.client = cli
	return c.client, nil
}

// negotiateAPIVersion 获取 Docker 服务器的 API 版本
func (c *TCPConnection) negotiateAPIVersion(cli *client.Client) string {
	// 使用 client 的 Ping 方法获取版本
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ping, err := cli.Ping(ctx)
	if err != nil {
		// 如果 ping 失败，使用默认版本
		return "1.45"
	}

	// 使用服务器返回的 API 版本
	if ping.APIVersion != "" {
		return ping.APIVersion
	}

	return "1.45"
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
		"type":     "tcp",
		"host":     c.config.Host,
		"port":     fmt.Sprintf("%d", c.config.Port),
		"tls":      fmt.Sprintf("%v", !c.config.SkipTLSVerify),
		"desc":     "TCP+TLS 连接",
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
