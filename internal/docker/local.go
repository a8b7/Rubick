package docker

import (
	"context"
	"fmt"

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

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("创建本地 Docker 客户端失败: %w", err)
	}

	c.client = cli
	return c.client, nil
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
