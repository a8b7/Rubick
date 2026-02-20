package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// NetworkService 网络服务
type NetworkService struct {
	client *client.Client
}

// NewNetworkService 创建网络服务
func NewNetworkService(cli *client.Client) *NetworkService {
	return &NetworkService{client: cli}
}

// DockerNetwork Docker 网络信息
type DockerNetwork struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Driver      string            `json:"driver"`
	Scope       string            `json:"scope"`
	IPAM        IPAMInfo          `json:"ipam"`
	Created     string            `json:"created"`
	Labels      map[string]string `json:"labels"`
	Internal    bool              `json:"internal"`
	Attachable  bool              `json:"attachable"`
	Ingress     bool              `json:"ingress"`
	EnableIPv6  bool              `json:"enable_ipv6"`
}

// IPAMInfo IPAM 信息
type IPAMInfo struct {
	Driver string        `json:"driver"`
	Config []IPAMConfig  `json:"config"`
}

// IPAMConfig IPAM 配置
type IPAMConfig struct {
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway,omitempty"`
}

// CreateNetworkOptions 创建网络选项
type CreateNetworkOptions struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	IPAM       *IPAMInfo         `json:"ipam"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
}

// List 列出网络
func (s *NetworkService) List(ctx context.Context) ([]DockerNetwork, error) {
	networks, err := s.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取网络列表失败: %w", err)
	}

	result := make([]DockerNetwork, 0, len(networks))
	for _, n := range networks {
		info := DockerNetwork{
			ID:         n.ID,
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Created:    n.Created.Format("2006-01-02T15:04:05Z07:00"),
			Labels:     n.Labels,
			Internal:   n.Internal,
			Attachable: n.Attachable,
			Ingress:    n.Ingress,
			EnableIPv6: n.EnableIPv6,
			IPAM: IPAMInfo{
				Driver: n.IPAM.Driver,
				Config: make([]IPAMConfig, 0, len(n.IPAM.Config)),
			},
		}

		for _, cfg := range n.IPAM.Config {
			info.IPAM.Config = append(info.IPAM.Config, IPAMConfig{
				Subnet:  cfg.Subnet,
				Gateway: cfg.Gateway,
			})
		}

		result = append(result, info)
	}

	return result, nil
}

// Get 获取网络详情
func (s *NetworkService) Get(ctx context.Context, id string) (*DockerNetwork, error) {
	n, err := s.client.NetworkInspect(ctx, id, network.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取网络详情失败: %w", err)
	}

	info := &DockerNetwork{
		ID:         n.ID,
		Name:       n.Name,
		Driver:     n.Driver,
		Scope:      n.Scope,
		Created:    n.Created.Format("2006-01-02T15:04:05Z07:00"),
		Labels:     n.Labels,
		Internal:   n.Internal,
		Attachable: n.Attachable,
		Ingress:    n.Ingress,
		EnableIPv6: n.EnableIPv6,
		IPAM: IPAMInfo{
			Driver: n.IPAM.Driver,
			Config: make([]IPAMConfig, 0, len(n.IPAM.Config)),
		},
	}

	for _, cfg := range n.IPAM.Config {
		info.IPAM.Config = append(info.IPAM.Config, IPAMConfig{
			Subnet:  cfg.Subnet,
			Gateway: cfg.Gateway,
		})
	}

	return info, nil
}

// Create 创建网络
func (s *NetworkService) Create(ctx context.Context, opts CreateNetworkOptions) (*DockerNetwork, error) {
	createOpts := network.CreateOptions{
		Driver:     opts.Driver,
		Scope:      opts.Scope,
		Internal:   opts.Internal,
		Attachable: opts.Attachable,
		Options:    opts.Options,
		Labels:     opts.Labels,
	}

	if opts.IPAM != nil {
		createOpts.IPAM = &network.IPAM{
			Driver: opts.IPAM.Driver,
		}
		for _, cfg := range opts.IPAM.Config {
			createOpts.IPAM.Config = append(createOpts.IPAM.Config, network.IPAMConfig{
				Subnet:  cfg.Subnet,
				Gateway: cfg.Gateway,
			})
		}
	}

	resp, err := s.client.NetworkCreate(ctx, opts.Name, createOpts)
	if err != nil {
		return nil, fmt.Errorf("创建网络失败: %w", err)
	}

	// 获取创建的网络详情
	return s.Get(ctx, resp.ID)
}

// Remove 删除网络
func (s *NetworkService) Remove(ctx context.Context, id string) error {
	err := s.client.NetworkRemove(ctx, id)
	if err != nil {
		return fmt.Errorf("删除网络失败: %w", err)
	}
	return nil
}
