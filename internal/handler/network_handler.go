package handler

import (
	"rubick/internal/docker"

	"github.com/gin-gonic/gin"
)

// ListNetworks 列出网络
func ListNetworks(c *gin.Context) {
	hostID := c.Query("host_id")

	// 获取主机
	host, err := getHost(hostID)
	if err != nil {
		ServerError(c, "获取主机失败: "+err.Error())
		return
	}

	// 获取 Docker 客户端
	cli, err := getClient(c.Request.Context(), host)
	if err != nil {
		ServerError(c, "获取 Docker 客户端失败: "+err.Error())
		return
	}

	svc := docker.NewNetworkService(cli)
	networks, err := svc.List(c.Request.Context())
	if err != nil {
		ServerError(c, "获取网络列表失败: "+err.Error())
		return
	}

	Success(c, networks)
}

// GetNetwork 获取网络详情
func GetNetwork(c *gin.Context) {
	id := c.Param("id")
	hostID := c.Query("host_id")

	host, err := getHost(hostID)
	if err != nil {
		ServerError(c, "获取主机失败: "+err.Error())
		return
	}

	cli, err := getClient(c.Request.Context(), host)
	if err != nil {
		ServerError(c, "获取 Docker 客户端失败: "+err.Error())
		return
	}

	svc := docker.NewNetworkService(cli)
	info, err := svc.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "网络不存在: "+err.Error())
		return
	}

	Success(c, info)
}

// CreateNetwork 创建网络
func CreateNetwork(c *gin.Context) {
	var req struct {
		HostID      string                  `json:"host_id"`
		Name        string                  `json:"name" binding:"required"`
		Driver      string                  `json:"driver"`
		Scope       string                  `json:"scope"`
		Internal    bool                    `json:"internal"`
		Attachable  bool                    `json:"attachable"`
		Subnet      string                  `json:"subnet"`
		Gateway     string                  `json:"gateway"`
		Options     map[string]string       `json:"options"`
		Labels      map[string]string       `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	host, err := getHost(req.HostID)
	if err != nil {
		ServerError(c, "获取主机失败: "+err.Error())
		return
	}

	cli, err := getClient(c.Request.Context(), host)
	if err != nil {
		ServerError(c, "获取 Docker 客户端失败: "+err.Error())
		return
	}

	opts := docker.CreateNetworkOptions{
		Name:       req.Name,
		Driver:     req.Driver,
		Scope:      req.Scope,
		Internal:   req.Internal,
		Attachable: req.Attachable,
		Options:    req.Options,
		Labels:     req.Labels,
	}

	// 设置 IPAM 配置
	if req.Subnet != "" {
		opts.IPAM = &docker.IPAMInfo{
			Driver: "default",
			Config: []docker.IPAMConfig{
				{
					Subnet:  req.Subnet,
					Gateway: req.Gateway,
				},
			},
		}
	}

	svc := docker.NewNetworkService(cli)
	info, err := svc.Create(c.Request.Context(), opts)
	if err != nil {
		ServerError(c, "创建网络失败: "+err.Error())
		return
	}

	Success(c, info)
}

// RemoveNetwork 删除网络
func RemoveNetwork(c *gin.Context) {
	id := c.Param("id")
	hostID := c.Query("host_id")

	host, err := getHost(hostID)
	if err != nil {
		ServerError(c, "获取主机失败: "+err.Error())
		return
	}

	cli, err := getClient(c.Request.Context(), host)
	if err != nil {
		ServerError(c, "获取 Docker 客户端失败: "+err.Error())
		return
	}

	svc := docker.NewNetworkService(cli)
	if err := svc.Remove(c.Request.Context(), id); err != nil {
		ServerError(c, "删除网络失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "网络删除成功", nil)
}
