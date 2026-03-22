package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

// LocalConnection 本地 Docker 连接
type LocalConnection struct {
	client *client.Client
}

// NewLocalConnection 创建本地连接
func NewLocalConnection() *LocalConnection {
	return &LocalConnection{}
}

// Connect 建立连接
func (c *LocalConnection) Connect(ctx context.Context) (*client.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	// 先创建临时 client 用于版本协商
	tempCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("创建临时本地 Docker 客户端失败: %w", err)
	}

	// 协商 API 版本
	apiVersion := c.negotiateAPIVersion(tempCli)
	tempCli.Close()

	// 使用协商后的版本创建正式 client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion(apiVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("创建本地 Docker 客户端失败: %w", err)
	}

	c.client = cli
	return c.client, nil
}

// negotiateAPIVersion 获取 Docker 服务器的 API 版本
func (c *LocalConnection) negotiateAPIVersion(cli *client.Client) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ping, err := cli.Ping(ctx)
	if err != nil {
		return "1.45"
	}

	if ping.APIVersion != "" {
		return ping.APIVersion
	}

	return "1.45"
}

// Close 关闭连接
func (c *LocalConnection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Type 返回连接类型
func (c *LocalConnection) Type() ConnectionType {
	return ConnectionTypeLocal
}

// Info 返回连接信息
func (c *LocalConnection) Info() map[string]string {
	return map[string]string{
		"type": "local",
		"desc": "本地 Docker 连接",
	}
}

// Test 测试连接是否可用
func (c *LocalConnection) Test(ctx context.Context) error {
	cli, err := c.Connect(ctx)
	if err != nil {
		return err
	}

	_, err = cli.Ping(ctx)
	return err
}
