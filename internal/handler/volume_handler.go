package handler

import (
	"rubick/internal/docker"

	"github.com/gin-gonic/gin"
)

// ListVolumes 列出卷
func ListVolumes(c *gin.Context) {
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

	svc := docker.NewVolumeService(cli)
	volumes, err := svc.List(c.Request.Context())
	if err != nil {
		ServerError(c, "获取卷列表失败: "+err.Error())
		return
	}

	Success(c, volumes)
}

// GetVolume 获取卷详情
func GetVolume(c *gin.Context) {
	name := c.Param("name")
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

	svc := docker.NewVolumeService(cli)
	info, err := svc.Get(c.Request.Context(), name)
	if err != nil {
		NotFound(c, "卷不存在: "+err.Error())
		return
	}

	Success(c, info)
}

// CreateVolume 创建卷
func CreateVolume(c *gin.Context) {
	var req docker.CreateVolumeOptions
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

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

	svc := docker.NewVolumeService(cli)
	info, err := svc.Create(c.Request.Context(), req)
	if err != nil {
		ServerError(c, "创建卷失败: "+err.Error())
		return
	}

	Success(c, info)
}

// RemoveVolume 删除卷
func RemoveVolume(c *gin.Context) {
	name := c.Param("name")
	force := c.Query("force") == "true"
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

	svc := docker.NewVolumeService(cli)
	if err := svc.Remove(c.Request.Context(), name, force); err != nil {
		ServerError(c, "删除卷失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "卷删除成功", nil)
}
