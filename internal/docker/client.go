package docker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"rubick/internal/model"
)

// ClientManager Docker 客户端管理器
type ClientManager struct {
	connections map[string]Connection
	mu          sync.RWMutex
}

var manager *ClientManager
var once sync.Once

// GetManager 获取客户端管理器单例
func GetManager() *ClientManager {
	once.Do(func() {
		manager = &ClientManager{
			connections: make(map[string]Connection),
		}
	})
	return manager
}

// GetClient 获取 Docker 客户端
func (m *ClientManager) GetClient(ctx context.Context, host *model.Host) (Connection, error) {
	m.mu.RLock()
	conn, exists := m.connections[host.ID]
	m.mu.RUnlock()

	// 如果连接存在，测试是否有效
	if exists {
		if err := conn.Test(ctx); err == nil {
			return conn, nil
		}
		// 连接失效，关闭并删除
		m.mu.Lock()
		conn.Close()
		delete(m.connections, host.ID)
		m.mu.Unlock()
	}

	// 创建新连接
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if conn, exists = m.connections[host.ID]; exists {
		if err := conn.Test(ctx); err == nil {
			return conn, nil
		}
		conn.Close()
		delete(m.connections, host.ID)
	}

	conn, err := m.createConnection(host)
	if err != nil {
		return nil, err
	}

	// 测试新连接
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := conn.Test(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("连接测试失败: %w", err)
	}

	m.connections[host.ID] = conn
	return conn, nil
}

// createConnection 根据主机配置创建连接
func (m *ClientManager) createConnection(host *model.Host) (Connection, error) {
	switch host.Type {
	case string(ConnectionTypeLocal):
		return NewLocalConnection(), nil
	case string(ConnectionTypeTCP):
		return NewTCPConnection(&ConnectionConfig{
			Type:          ConnectionTypeTCP,
			Host:          host.Host,
			Port:          host.DockerPort,
			SkipTLSVerify: host.SkipTLSVerify,
			TLSCertID:     host.TLSCertID,
		}), nil
	case string(ConnectionTypeSSH):
		return NewSSHConnection(&ConnectionConfig{
			Type:          ConnectionTypeSSH,
			Host:          host.Host,
			Port:          host.DockerPort,
			SSHUser:       host.SSHUser,
			SSHAuthType:   host.SSHAuthType,
			SSHPrivateKey: host.SSHPrivateKey,
			SSHPassword:   host.SSHPassword,
			SSHPort:       host.SSHPort,
		}), nil
	default:
		return nil, fmt.Errorf("不支持的主机类型: %s", host.Type)
	}
}

// RemoveClient 移除客户端连接
func (m *ClientManager) RemoveClient(hostID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[hostID]
	if !exists {
		return nil
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("关闭连接失败: %w", err)
	}

	delete(m.connections, hostID)
	return nil
}

// TestConnection 测试连接
func (m *ClientManager) TestConnection(ctx context.Context, host *model.Host) error {
	conn, err := m.createConnection(host)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Test(ctx)
}

// CloseAll 关闭所有连接
func (m *ClientManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for id, conn := range m.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭连接 %s 失败: %w", id, err))
		}
	}

	m.connections = make(map[string]Connection)

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
	}
	return nil
}
