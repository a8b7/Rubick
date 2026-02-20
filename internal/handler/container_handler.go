package handler

import (
	"context"
	"io"

	"rubick/internal/docker"
	"rubick/internal/model"
	"rubick/internal/repository"

	containerTypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

// ListContainers 列出容器
func ListContainers(c *gin.Context) {
	hostID := c.Query("host_id")
	all := c.Query("all") == "true"

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

	svc := docker.NewContainerService(cli)
	containers, err := svc.List(c.Request.Context(), all)
	if err != nil {
		ServerError(c, "获取容器列表失败: "+err.Error())
		return
	}

	Success(c, containers)
}

// CreateContainer 创建容器
func CreateContainer(c *gin.Context) {
	hostID := c.Query("host_id")

	var req docker.CreateContainerOptions
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

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

	svc := docker.NewContainerService(cli)
	resp, err := svc.Create(c.Request.Context(), req)
	if err != nil {
		ServerError(c, "创建容器失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id":       resp.ID,
		"warnings": resp.Warnings,
	})
}

// GetContainer 获取容器详情
func GetContainer(c *gin.Context) {
	containerID := c.Param("id")
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

	svc := docker.NewContainerService(cli)
	inspect, err := svc.Get(c.Request.Context(), containerID)
	if err != nil {
		NotFound(c, "容器不存在: "+err.Error())
		return
	}

	Success(c, inspect)
}

// StartContainer 启动容器
func StartContainer(c *gin.Context) {
	containerID := c.Param("id")
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

	svc := docker.NewContainerService(cli)
	if err := svc.Start(c.Request.Context(), containerID); err != nil {
		ServerError(c, "启动容器失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "容器启动成功", nil)
}

// StopContainer 停止容器
func StopContainer(c *gin.Context) {
	containerID := c.Param("id")
	hostID := c.Query("host_id")

	var req struct {
		Timeout int `json:"timeout"`
	}
	c.ShouldBindJSON(&req)

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

	svc := docker.NewContainerService(cli)
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 10
	}
	if err := svc.Stop(c.Request.Context(), containerID, &timeout); err != nil {
		ServerError(c, "停止容器失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "容器停止成功", nil)
}

// RestartContainer 重启容器
func RestartContainer(c *gin.Context) {
	containerID := c.Param("id")
	hostID := c.Query("host_id")

	var req struct {
		Timeout int `json:"timeout"`
	}
	c.ShouldBindJSON(&req)

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

	svc := docker.NewContainerService(cli)
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 10
	}
	if err := svc.Restart(c.Request.Context(), containerID, &timeout); err != nil {
		ServerError(c, "重启容器失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "容器重启成功", nil)
}

// RemoveContainer 删除容器
func RemoveContainer(c *gin.Context) {
	containerID := c.Param("id")
	hostID := c.Query("host_id")
	force := c.Query("force") == "true"
	volumes := c.Query("volumes") == "true"
	removeImage := c.Query("remove_image") == "true"

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

	svc := docker.NewContainerService(cli)

	// 如果需要删除镜像，先获取容器的镜像信息
	var imageID string
	if removeImage {
		imageID, err = svc.GetImageID(c.Request.Context(), containerID)
		if err != nil {
			ServerError(c, "获取容器镜像信息失败: "+err.Error())
			return
		}
	}

	// 删除容器
	if err := svc.Remove(c.Request.Context(), containerID, force, volumes); err != nil {
		ServerError(c, "删除容器失败: "+err.Error())
		return
	}

	// 删除镜像
	if removeImage && imageID != "" {
		imageSvc := docker.NewImageService(cli)
		if _, err := imageSvc.Remove(c.Request.Context(), imageID, true); err != nil {
			SuccessWithMessage(c, "容器已删除，但删除镜像失败: "+err.Error(), nil)
			return
		}
	}

	SuccessWithMessage(c, "容器删除成功", nil)
}

// GetContainerLogs 获取容器日志
func GetContainerLogs(c *gin.Context) {
	containerID := c.Param("id")
	hostID := c.Query("host_id")

	tail := c.DefaultQuery("tail", "100")
	follow := c.Query("follow") == "true"

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

	svc := docker.NewContainerService(cli)
	reader, err := svc.Logs(c.Request.Context(), containerID, containerTypes.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
	})
	if err != nil {
		ServerError(c, "获取容器日志失败: "+err.Error())
		return
	}
	defer reader.Close()

	// 读取日志
	logs, err := io.ReadAll(reader)
	if err != nil {
		ServerError(c, "读取日志失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"logs": string(logs),
	})
}

// GetContainerStats 获取容器资源统计
func GetContainerStats(c *gin.Context) {
	containerID := c.Param("id")
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

	svc := docker.NewContainerService(cli)
	stats, err := svc.Stats(c.Request.Context(), containerID)
	if err != nil {
		ServerError(c, "获取容器统计失败: "+err.Error())
		return
	}

	Success(c, stats)
}

// ExecContainer 在容器中执行命令
func ExecContainer(c *gin.Context) {
	containerID := c.Param("id")
	hostID := c.Query("host_id")

	var req struct {
		Cmd     []string `json:"cmd" binding:"required"`
		User    string   `json:"user"`
		WorkDir string   `json:"work_dir"`
		Tty     bool     `json:"tty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

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

	svc := docker.NewContainerService(cli)
	resp, err := svc.ExecCreate(c.Request.Context(), containerID, req.Cmd, req.User)
	if err != nil {
		ServerError(c, "创建执行命令失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"id": resp.ID,
	})
}

// getHost 获取主机配置
func getHost(hostID string) (*model.Host, error) {
	if hostID == "" {
		return repository.GetDefaultHost()
	}
	return repository.GetHostByID(hostID)
}

// getClient 获取 Docker 客户端
func getClient(ctx context.Context, host *model.Host) (*client.Client, error) {
	conn, err := docker.GetManager().GetClient(ctx, host)
	if err != nil {
		return nil, err
	}

	// 通过 Connection 接口的 Connect 方法获取客户端
	return conn.Connect(ctx)
}
