package handler

import (
	"rubick/internal/docker"
	"rubick/internal/model"
	"rubick/internal/repository"

	"github.com/gin-gonic/gin"
)

// ListHosts 列出所有主机
func ListHosts(c *gin.Context) {
	hosts, err := repository.ListHosts()
	if err != nil {
		ServerError(c, "获取主机列表失败: "+err.Error())
		return
	}
	// 清除敏感字段
	for i := range hosts {
		hosts[i].ClearSensitiveFields()
	}
	Success(c, hosts)
}

// CreateHost 创建主机
func CreateHost(c *gin.Context) {
	var host model.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// 验证主机类型
	if host.Type != "local" && host.Type != "tcp" && host.Type != "ssh" {
		BadRequest(c, "无效的主机类型，必须是 local、tcp 或 ssh")
		return
	}

	// 如果设为默认，取消其他默认主机
	if host.IsDefault {
		repository.ClearDefaultHost()
	}

	if err := repository.CreateHost(&host); err != nil {
		ServerError(c, "创建主机失败: "+err.Error())
		return
	}

	// 清除敏感字段后再返回
	host.ClearSensitiveFields()
	SuccessWithMessage(c, "主机创建成功", host)
}

// GetHost 获取主机详情
func GetHost(c *gin.Context) {
	id := c.Param("id")
	host, err := repository.GetHostByID(id)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}
	host.ClearSensitiveFields()
	Success(c, host)
}

// UpdateHost 更新主机
func UpdateHost(c *gin.Context) {
	id := c.Param("id")

	var updates model.Host
	if err := c.ShouldBindJSON(&updates); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// 如果设为默认，取消其他默认主机
	if updates.IsDefault {
		repository.ClearDefaultHost()
	}

	if err := repository.UpdateHost(id, &updates); err != nil {
		ServerError(c, "更新主机失败: "+err.Error())
		return
	}

	// 移除缓存的 Docker 连接，使新配置生效
	docker.GetManager().RemoveClient(id)

	host, _ := repository.GetHostByID(id)
	if host != nil {
		host.ClearSensitiveFields()
	}
	SuccessWithMessage(c, "主机更新成功", host)
}

// DeleteHost 删除主机
func DeleteHost(c *gin.Context) {
	id := c.Param("id")

	host, err := repository.GetHostByID(id)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}

	// 不能删除本地主机
	if host.Type == "local" {
		BadRequest(c, "不能删除本地主机")
		return
	}

	// 移除缓存的 Docker 连接
	docker.GetManager().RemoveClient(id)

	if err := repository.DeleteHost(id); err != nil {
		ServerError(c, "删除主机失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "主机删除成功", nil)
}

// TestHostConnection 测试主机连接
func TestHostConnection(c *gin.Context) {
	id := c.Param("id")

	host, err := repository.GetHostByID(id)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}

	// 测试连接
	err = docker.GetManager().TestConnection(c.Request.Context(), host)
	if err != nil {
		Success(c, gin.H{
			"success": false,
			"message": "连接失败: " + err.Error(),
			"host":    host.Name,
		})
		return
	}

	Success(c, gin.H{
		"success": true,
		"message": "连接成功",
		"host":    host.Name,
	})
}
