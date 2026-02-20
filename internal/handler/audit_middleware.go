package handler

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"rubick/internal/model"
	"rubick/internal/repository"
)

// responseWriter 用于捕获响应状态码
type responseWriter struct {
	gin.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.ResponseWriter.Write(b)
}

// AuditMiddleware 审计日志中间件
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（如果需要）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装 ResponseWriter 以捕获状态码
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			status:         200,
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 记录审计日志
		latency := time.Since(startTime).Milliseconds()

		// 只记录非健康检查的请求
		if c.Request.URL.Path != "/api/v1/health" {
			auditLog := &model.AuditLog{
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				Status:    writer.status,
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Latency:   latency,
				Message:   getStatusMessage(writer.status),
			}

			// 异步写入日志，避免影响请求性能
			go func() {
				repository.CreateAuditLog(auditLog)
			}()
		}

		_ = requestBody // 避免未使用警告
	}
}

func getStatusMessage(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "成功"
	case status >= 400 && status < 500:
		return "客户端错误"
	case status >= 500:
		return "服务器错误"
	default:
		return "未知"
	}
}
