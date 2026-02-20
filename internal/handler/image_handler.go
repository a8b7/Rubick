package handler

import (
	"rubick/internal/docker"

	"github.com/gin-gonic/gin"
)

// ListImages 列出镜像
func ListImages(c *gin.Context) {
	hostID := c.Query("host_id")
	all := c.Query("all") == "true"

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

	svc := docker.NewImageService(cli)
	images, err := svc.List(c.Request.Context(), all)
	if err != nil {
		ServerError(c, "获取镜像列表失败: "+err.Error())
		return
	}

	Success(c, images)
}

// PullImage 拉取镜像
func PullImage(c *gin.Context) {
	var req struct {
		HostID     string `json:"host_id"`
		Image      string `json:"image" binding:"required"`
		Registry   string `json:"registry"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Platform   string `json:"platform"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// TODO: 实现镜像拉取逻辑

	Success(c, gin.H{
		"status":  "pulling",
		"message": "正在拉取镜像: " + req.Image,
	})
}

// GetImage 获取镜像详情
func GetImage(c *gin.Context) {
	imageID := c.Param("id")
	hostID := c.Query("host_id")

	// TODO: 实现镜像详情获取逻辑
	_ = imageID
	_ = hostID

	Success(c, gin.H{
		"id": imageID,
	})
}

// RemoveImage 删除镜像
func RemoveImage(c *gin.Context) {
	imageID := c.Param("id")
	force := c.Query("force") == "true"

	// TODO: 实现镜像删除逻辑
	_ = imageID
	_ = force

	SuccessWithMessage(c, "镜像删除成功", nil)
}

// TagImage 标记镜像
func TagImage(c *gin.Context) {
	imageID := c.Param("id")

	var req struct {
		Repo  string `json:"repo" binding:"required"`
		Tag   string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// TODO: 实现镜像标记逻辑
	_ = imageID

	SuccessWithMessage(c, "镜像标记成功", nil)
}

// SearchImages 搜索镜像
func SearchImages(c *gin.Context) {
	term := c.Query("term")
	limit := c.DefaultQuery("limit", "25")

	if term == "" {
		BadRequest(c, "搜索关键词不能为空")
		return
	}

	// TODO: 实现镜像搜索逻辑
	_ = limit

	Success(c, []interface{}{})
}
