package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// 成功响应带消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// 分页响应
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// 失败响应
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// 失败响应带 HTTP 状态码
func FailWithStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// 参数错误
func BadRequest(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusBadRequest, 400, message)
}

// 未授权
func Unauthorized(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusUnauthorized, 401, message)
}

// 禁止访问
func Forbidden(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusForbidden, 403, message)
}

// 资源未找到
func NotFound(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusNotFound, 404, message)
}

// 服务器错误
func ServerError(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusInternalServerError, 500, message)
}

// 错误码定义
const (
	CodeSuccess           = 0
	CodeBadRequest        = 400
	CodeUnauthorized      = 401
	CodeForbidden         = 403
	CodeNotFound          = 404
	CodeServerError       = 500
	CodeDockerError       = 1001
	CodeConnectionFailed  = 1002
	CodeContainerNotFound = 1003
	CodeImageNotFound     = 1004
	CodeInvalidConfig     = 1005
)
