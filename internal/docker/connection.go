package docker

import (
	"context"

	"github.com/docker/docker/client"
)

// ConnectionType 连接类型
type ConnectionType string

const (
	ConnectionTypeLocal ConnectionType = "local"
	ConnectionTypeTCP   ConnectionType = "tcp"
	ConnectionTypeSSH   ConnectionType = "ssh"
)

// Connection Docker 连接接口
type Connection interface {
	// Connect 建立连接
	Connect(ctx context.Context) (*client.Client, error)

	// Close 关闭连接
	Close() error

	// Type 返回连接类型
	Type() ConnectionType

	// Info 返回连接信息
	Info() map[string]string

	// Test 测试连接是否可用
	Test(ctx context.Context) error
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Type ConnectionType

	// TCP 配置
	Host         string
	Port         int
	SkipTLSVerify bool
	TLSCertID    string

	// SSH 配置
	SSHUser       string
	SSHAuthType   string // key, password
	SSHPrivateKey string
	SSHPassword   string
	SSHPort       int
}
