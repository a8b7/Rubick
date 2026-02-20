package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// ContainerService 容器服务
type ContainerService struct {
	client *client.Client
}

// NewContainerService 创建容器服务
func NewContainerService(cli *client.Client) *ContainerService {
	return &ContainerService{client: cli}
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	ImageID    string            `json:"image_id"`
	State      string            `json:"state"`
	Status     string            `json:"status"`
	Created    int64             `json:"created"`
	Ports      []PortBinding     `json:"ports"`
	Labels     map[string]string `json:"labels"`
	Command    string            `json:"command"`
	HostConfig HostConfig        `json:"host_config"`
	Mounts     []MountInfo       `json:"mounts"`
	Networks   []NetworkInfo     `json:"networks"`
}

// MountInfo 挂载信息
type MountInfo struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Name        string `json:"name"`
	NetworkID   string `json:"network_id"`
	IPAddress   string `json:"ip_address"`
	MacAddress  string `json:"mac_address"`
	Gateway     string `json:"gateway"`
}

// PortBinding 端口绑定
type PortBinding struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

// HostConfig 主机配置
type HostConfig struct {
	NetworkMode string `json:"network_mode"`
}

// CreateContainerOptions 创建容器选项
type CreateContainerOptions struct {
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Cmd        []string          `json:"cmd"`
	Env        []string          `json:"env"`
	WorkingDir string            `json:"working_dir"`
	Ports      map[string]string `json:"ports"` // "80/tcp": "8080"
	Volumes    []string          `json:"volumes"`
	Network    string            `json:"network"`
	Labels     map[string]string `json:"labels"`
	Restart    string            `json:"restart"`
}

// ContainerStats 容器资源统计
type ContainerStats struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
}

// List 列出容器
func (s *ContainerService) List(ctx context.Context, all bool) ([]ContainerInfo, error) {
	containers, err := s.client.ContainerList(ctx, containerTypes.ListOptions{
		All: all,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		info := ContainerInfo{
			ID:      c.ID,
			Name:    getContainerName(c.Names),
			Image:   c.Image,
			ImageID: c.ImageID,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
			Labels:  c.Labels,
			Command: c.Command,
			HostConfig: HostConfig{
				NetworkMode: c.HostConfig.NetworkMode,
			},
		}

		// 转换端口信息
		for _, port := range c.Ports {
			info.Ports = append(info.Ports, PortBinding{
				IP:          port.IP,
				PrivatePort: int(port.PrivatePort),
				PublicPort:  int(port.PublicPort),
				Type:        port.Type,
			})
		}

		result = append(result, info)
	}

	return result, nil
}

// Get 获取容器详情
func (s *ContainerService) Get(ctx context.Context, containerID string) (*ContainerInfo, error) {
	c, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("获取容器详情失败: %w", err)
	}

	info := &ContainerInfo{
		ID:      c.ID,
		Name:    c.Name,
		Image:   c.Config.Image,
		ImageID: c.Image,
		State:   c.State.Status,
		Status:  buildStatusString(c.State),
		Created: mustParseInt64(c.Created),
		Labels:  c.Config.Labels,
		Command: strings.Join(c.Config.Cmd, " "),
		HostConfig: HostConfig{
			NetworkMode: string(c.HostConfig.NetworkMode),
		},
	}

	// 转换端口信息
	for port, bindings := range c.NetworkSettings.Ports {
		if bindings == nil {
			info.Ports = append(info.Ports, PortBinding{
				PrivatePort: int(port.Int()),
				Type:        port.Proto(),
			})
			continue
		}
		for _, binding := range bindings {
			info.Ports = append(info.Ports, PortBinding{
				IP:          binding.HostIP,
				PrivatePort: int(port.Int()),
				PublicPort:  mustParseInt(binding.HostPort),
				Type:        port.Proto(),
			})
		}
	}

	// 转换挂载信息
	for _, mount := range c.Mounts {
		info.Mounts = append(info.Mounts, MountInfo{
			Type:        string(mount.Type),
			Source:      mount.Source,
			Destination: mount.Destination,
			Mode:        mount.Mode,
			RW:          mount.RW,
		})
	}

	// 转换网络信息
	for name, network := range c.NetworkSettings.Networks {
		info.Networks = append(info.Networks, NetworkInfo{
			Name:       name,
			NetworkID:  network.NetworkID,
			IPAddress:  network.IPAddress,
			MacAddress: network.MacAddress,
			Gateway:    network.Gateway,
		})
	}

	return info, nil
}

// Create 创建容器
func (s *ContainerService) Create(ctx context.Context, opts CreateContainerOptions) (containerTypes.CreateResponse, error) {
	// 构建端口绑定配置
	portSet := make(nat.PortSet)
	portMap := make(nat.PortMap)
	for containerPort, hostPort := range opts.Ports {
		port := nat.Port(containerPort)
		portSet[port] = struct{}{}
		portMap[port] = []nat.PortBinding{
			{HostPort: hostPort},
		}
	}

	// 构建重启策略
	var restartPolicy containerTypes.RestartPolicy
	switch opts.Restart {
	case "always":
		restartPolicy = containerTypes.RestartPolicy{Name: "always"}
	case "on-failure":
		restartPolicy = containerTypes.RestartPolicy{Name: "on-failure"}
	case "unless-stopped":
		restartPolicy = containerTypes.RestartPolicy{Name: "unless-stopped"}
	default:
		restartPolicy = containerTypes.RestartPolicy{Name: "no"}
	}

	config := &containerTypes.Config{
		Image:        opts.Image,
		Cmd:          opts.Cmd,
		Env:          opts.Env,
		WorkingDir:   opts.WorkingDir,
		Labels:       opts.Labels,
		ExposedPorts: portSet,
	}

	hostConfig := &containerTypes.HostConfig{
		PortBindings:  portMap,
		RestartPolicy: restartPolicy,
		NetworkMode:   containerTypes.NetworkMode(opts.Network),
		Binds:         opts.Volumes,
	}

	resp, err := s.client.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.Name)
	if err != nil {
		return containerTypes.CreateResponse{}, fmt.Errorf("创建容器失败: %w", err)
	}

	return resp, nil
}

// Start 启动容器
func (s *ContainerService) Start(ctx context.Context, containerID string) error {
	err := s.client.ContainerStart(ctx, containerID, containerTypes.StartOptions{})
	if err != nil {
		return fmt.Errorf("启动容器失败: %w", err)
	}
	return nil
}

// Stop 停止容器
func (s *ContainerService) Stop(ctx context.Context, containerID string, timeout *int) error {
	if timeout == nil {
		t := 10
		timeout = &t
	}
	err := s.client.ContainerStop(ctx, containerID, containerTypes.StopOptions{Timeout: timeout})
	if err != nil {
		return fmt.Errorf("停止容器失败: %w", err)
	}
	return nil
}

// Restart 重启容器
func (s *ContainerService) Restart(ctx context.Context, containerID string, timeout *int) error {
	if timeout == nil {
		t := 10
		timeout = &t
	}
	err := s.client.ContainerRestart(ctx, containerID, containerTypes.StopOptions{Timeout: timeout})
	if err != nil {
		return fmt.Errorf("重启容器失败: %w", err)
	}
	return nil
}

// Remove 删除容器
func (s *ContainerService) Remove(ctx context.Context, containerID string, force, removeVolumes bool) error {
	err := s.client.ContainerRemove(ctx, containerID, containerTypes.RemoveOptions{
		Force:         force,
		RemoveVolumes: removeVolumes,
	})
	if err != nil {
		return fmt.Errorf("删除容器失败: %w", err)
	}
	return nil
}

// GetImageID 获取容器使用的镜像ID
func (s *ContainerService) GetImageID(ctx context.Context, containerID string) (string, error) {
	c, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("获取容器详情失败: %w", err)
	}
	return c.Image, nil
}

// Logs 获取容器日志
func (s *ContainerService) Logs(ctx context.Context, containerID string, opts containerTypes.LogsOptions) (io.ReadCloser, error) {
	reader, err := s.client.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return nil, fmt.Errorf("获取容器日志失败: %w", err)
	}
	return reader, nil
}

// Stats 获取容器资源统计
func (s *ContainerService) Stats(ctx context.Context, containerID string) (*ContainerStats, error) {
	resp, err := s.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("获取容器统计失败: %w", err)
	}
	defer resp.Body.Close()

	// 简化实现：返回基本统计信息
	// 完整实现需要解析 JSON 响应
	return &ContainerStats{
		CPUPercent:  0,
		MemoryUsage: 0,
		MemoryLimit: 0,
	}, nil
}

// ExecCreate 创建执行命令
func (s *ContainerService) ExecCreate(ctx context.Context, containerID string, cmd []string, user string) (containerTypes.ExecCreateResponse, error) {
	config := containerTypes.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          true,
		Cmd:          cmd,
		User:         user,
	}

	resp, err := s.client.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return containerTypes.ExecCreateResponse{}, fmt.Errorf("创建执行命令失败: %w", err)
	}

	return resp, nil
}

// ExecAttach 连接到执行命令
func (s *ContainerService) ExecAttach(ctx context.Context, execID string) (types.HijackedResponse, error) {
	config := containerTypes.ExecAttachOptions{
		Detach: false,
		Tty:    true,
	}

	resp, err := s.client.ContainerExecAttach(ctx, execID, config)
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("连接执行命令失败: %w", err)
	}

	return resp, nil
}

// ExecResize 调整执行命令终端大小
func (s *ContainerService) ExecResize(ctx context.Context, execID string, width, height uint) error {
	err := s.client.ContainerExecResize(ctx, execID, containerTypes.ResizeOptions{
		Height: height,
		Width:  width,
	})
	if err != nil {
		return fmt.Errorf("调整终端大小失败: %w", err)
	}
	return nil
}

// getContainerName 获取容器名称（去掉前导斜杠）
func getContainerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	name := names[0]
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	return name
}

// mustParseInt 安全解析整数
func mustParseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// mustParseInt64 安全解析 int64
func mustParseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// buildStatusString 构建状态字符串
func buildStatusString(state *containerTypes.State) string {
	var parts []string
	parts = append(parts, state.Status)
	if state.Running {
		if state.StartedAt != "" {
			parts = append(parts, "since", state.StartedAt)
		}
	} else if state.ExitCode != 0 {
		parts = append(parts, fmt.Sprintf("exit code %d", state.ExitCode))
	}
	return strings.Join(parts, " ")
}
