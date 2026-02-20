package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"rubick/internal/repository"
)

// ListAuditLogs 列出审计日志
func ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	method := c.Query("method")
	path := c.Query("path")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := repository.ListAuditLogs(page, pageSize, method, path)
	if err != nil {
		ServerError(c, "获取审计日志失败: "+err.Error())
		return
	}

	SuccessWithPage(c, logs, total, page, pageSize)
}
