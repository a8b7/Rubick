package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

// ImageService 镜像服务
type ImageService struct {
	client *client.Client
}

// NewImageService 创建镜像服务
func NewImageService(cli *client.Client) *ImageService {
	return &ImageService{client: cli}
}

// ImageInfo 镜像信息
type ImageInfo struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repo_tags"`
	RepoDigests []string          `json:"repo_digests"`
	Created     int64             `json:"created"`
	Size        int64             `json:"size"`
	Labels      map[string]string `json:"labels"`
}

// PullOptions 拉取镜像选项
type PullOptions struct {
	Image    string `json:"image"`
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password"`
	Platform string `json:"platform"`
}

// SearchOptions 搜索镜像选项
type SearchOptions struct {
	Term     string `json:"term"`
	Limit    int    `json:"limit"`
	Registry string `json:"registry"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
	Official    bool   `json:"official"`
	Automated   bool   `json:"automated"`
}

// List 列出镜像
func (s *ImageService) List(ctx context.Context, all bool) ([]ImageInfo, error) {
	images, err := s.client.ImageList(ctx, image.ListOptions{
		All: all,
	})
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}

	result := make([]ImageInfo, 0, len(images))
	for _, img := range images {
		info := ImageInfo{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Created:     img.Created,
			Size:        img.Size,
			Labels:      img.Labels,
		}
		result = append(result, info)
	}

	// 按仓库标签排序，没有标签的排在最后
	sort.Slice(result, func(i, j int) bool {
		iTag := ""
		if len(result[i].RepoTags) > 0 {
			iTag = result[i].RepoTags[0]
		}
		jTag := ""
		if len(result[j].RepoTags) > 0 {
			jTag = result[j].RepoTags[0]
		}
		return iTag < jTag
	})

	return result, nil
}

// Get 获取镜像详情
func (s *ImageService) Get(ctx context.Context, imageID string) (image.InspectResponse, error) {
	inspect, _, err := s.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return image.InspectResponse{}, fmt.Errorf("获取镜像详情失败: %w", err)
	}
	return inspect, nil
}

// Pull 拉取镜像
func (s *ImageService) Pull(ctx context.Context, opts PullOptions) (io.ReadCloser, error) {
	pullOpts := image.PullOptions{
		Platform: opts.Platform,
	}

	// 设置认证信息
	if opts.Username != "" && opts.Password != "" {
		auth := registry.AuthConfig{
			Username:      opts.Username,
			Password:      opts.Password,
			ServerAddress: opts.Registry,
		}
		authBytes, err := json.Marshal(auth)
		if err != nil {
			return nil, fmt.Errorf("编码认证信息失败: %w", err)
		}
		pullOpts.RegistryAuth = base64.StdEncoding.EncodeToString(authBytes)
	}

	reader, err := s.client.ImagePull(ctx, opts.Image, pullOpts)
	if err != nil {
		return nil, fmt.Errorf("拉取镜像失败: %w", err)
	}

	return reader, nil
}

// Remove 删除镜像
func (s *ImageService) Remove(ctx context.Context, imageID string, force bool) ([]image.DeleteResponse, error) {
	resp, err := s.client.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force: force,
	})
	if err != nil {
		return nil, fmt.Errorf("删除镜像失败: %w", err)
	}
	return resp, nil
}

// Tag 标记镜像
func (s *ImageService) Tag(ctx context.Context, imageID, repo, tag string) error {
	ref := repo
	if tag != "" {
		ref = fmt.Sprintf("%s:%s", repo, tag)
	}

	err := s.client.ImageTag(ctx, imageID, ref)
	if err != nil {
		return fmt.Errorf("标记镜像失败: %w", err)
	}
	return nil
}

// Search 搜索镜像
func (s *ImageService) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit == 0 {
		opts.Limit = 25
	}

	results, err := s.client.ImageSearch(ctx, opts.Term, registry.SearchOptions{
		Limit: opts.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索镜像失败: %w", err)
	}

	searchResults := make([]SearchResult, 0, len(results))
	for _, r := range results {
		searchResults = append(searchResults, SearchResult{
			Name:        r.Name,
			Description: r.Description,
			Stars:       r.StarCount,
			Official:    r.IsOfficial,
			Automated:   r.IsAutomated,
		})
	}

	return searchResults, nil
}

// Prune 清理未使用的镜像
func (s *ImageService) Prune(ctx context.Context, all bool) (image.PruneReport, error) {
	args := filters.NewArgs()
	if all {
		args.Add("dangling", "false")
	} else {
		args.Add("dangling", "true")
	}

	report, err := s.client.ImagesPrune(ctx, args)
	if err != nil {
		return image.PruneReport{}, fmt.Errorf("清理镜像失败: %w", err)
	}

	return report, nil
}
