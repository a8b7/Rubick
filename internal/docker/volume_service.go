package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// VolumeService 卷服务
type VolumeService struct {
	client *client.Client
}

// NewVolumeService 创建卷服务
func NewVolumeService(cli *client.Client) *VolumeService {
	return &VolumeService{client: cli}
}

// VolumeInfo 卷信息
type VolumeInfo struct {
	Name        string            `json:"name"`
	Driver      string            `json:"driver"`
	Mountpoint  string            `json:"mountpoint"`
	CreatedAt   string            `json:"created_at"`
	Labels      map[string]string `json:"labels"`
	Scope       string            `json:"scope"`
	Options     map[string]string `json:"options"`
	UsageData   *VolumeUsageData  `json:"usage_data,omitempty"`
}

// VolumeUsageData 卷使用数据
type VolumeUsageData struct {
	Size     int64 `json:"size"`
	RefCount int64 `json:"ref_count"`
}

// CreateVolumeOptions 创建卷选项
type CreateVolumeOptions struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driver_opts"`
	Labels     map[string]string `json:"labels"`
}

// List 列出卷
func (s *VolumeService) List(ctx context.Context) ([]VolumeInfo, error) {
	volumes, err := s.client.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取卷列表失败: %w", err)
	}

	result := make([]VolumeInfo, 0, len(volumes.Volumes))
	for _, v := range volumes.Volumes {
		info := VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			Labels:     v.Labels,
			Scope:      v.Scope,
			Options:    v.Options,
		}

		if v.UsageData != nil {
			info.UsageData = &VolumeUsageData{
				Size:     v.UsageData.Size,
				RefCount: v.UsageData.RefCount,
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// Get 获取卷详情
func (s *VolumeService) Get(ctx context.Context, name string) (*VolumeInfo, error) {
	v, err := s.client.VolumeInspect(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("获取卷详情失败: %w", err)
	}

	info := &VolumeInfo{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		CreatedAt:  v.CreatedAt,
		Labels:     v.Labels,
		Scope:      v.Scope,
		Options:    v.Options,
	}

	if v.UsageData != nil {
		info.UsageData = &VolumeUsageData{
			Size:     v.UsageData.Size,
			RefCount: v.UsageData.RefCount,
		}
	}

	return info, nil
}

// Create 创建卷
func (s *VolumeService) Create(ctx context.Context, opts CreateVolumeOptions) (*VolumeInfo, error) {
	createOpts := volume.CreateOptions{
		Name:       opts.Name,
		Driver:     opts.Driver,
		DriverOpts: opts.DriverOpts,
		Labels:     opts.Labels,
	}

	v, err := s.client.VolumeCreate(ctx, createOpts)
	if err != nil {
		return nil, fmt.Errorf("创建卷失败: %w", err)
	}

	info := &VolumeInfo{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		CreatedAt:  v.CreatedAt,
		Labels:     v.Labels,
		Scope:      v.Scope,
		Options:    v.Options,
	}

	return info, nil
}

// Remove 删除卷
func (s *VolumeService) Remove(ctx context.Context, name string, force bool) error {
	err := s.client.VolumeRemove(ctx, name, force)
	if err != nil {
		return fmt.Errorf("删除卷失败: %w", err)
	}
	return nil
}
