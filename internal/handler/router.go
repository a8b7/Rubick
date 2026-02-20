package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"rubick/internal/static"
)

// Router 路由管理
type Router struct {
	engine *gin.Engine
}

// NewRouter 创建路由
func NewRouter() *Router {
	return &Router{
		engine: gin.New(),
	}
}

// Setup 设置路由
func (r *Router) Setup() *gin.Engine {
	// 全局中间件
	r.engine.Use(gin.Logger())
	r.engine.Use(Recovery())
	r.engine.Use(CORS())
	r.engine.Use(AuditMiddleware())

	// API 路由组
	api := r.engine.Group("/api/v1")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			Success(c, gin.H{
				"status": "ok",
			})
		})

		// 审计日志路由
		api.GET("/audit/logs", ListAuditLogs)

		// 主机管理路由
		setupHostRoutes(api)

		// 容器管理路由
		setupContainerRoutes(api)

		// 镜像管理路由
		setupImageRoutes(api)

		// 卷管理路由
		setupVolumeRoutes(api)

		// 网络管理路由
		setupNetworkRoutes(api)

		// Compose 管理路由
		setupComposeRoutes(api)

		// WebSocket 路由
		api.GET("/ws/containers/:id/logs", ContainerLogsWS)
		api.GET("/ws/containers/:id/exec", ContainerExecWS)
		api.GET("/ws/compose/:id/logs", ComposeLogsWS)
	}

	// 静态文件服务
	r.setupStaticFiles()

	return r.engine
}

// setupStaticFiles 设置静态文件服务
func (r *Router) setupStaticFiles() {
	distFS, err := static.GetDistFS()
	if err != nil {
		// 如果没有前端构建产物，跳过静态文件服务
		return
	}

	// 静态资源文件
	r.engine.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(distFS))
	})

	// SPA fallback - 所有非 API 路由返回 index.html
	r.engine.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求但路由不存在，返回 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "路由不存在",
			})
			return
		}

		// 其他请求返回 index.html（SPA 路由）
		indexHTML, err := static.GetIndexHTML()
		if err != nil {
			c.String(http.StatusInternalServerError, "无法加载前端页面")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}

// setupHostRoutes 设置主机管理路由
func setupHostRoutes(rg *gin.RouterGroup) {
	hosts := rg.Group("/hosts")
	{
		hosts.GET("", ListHosts)
		hosts.POST("", CreateHost)
		hosts.GET("/:id", GetHost)
		hosts.PUT("/:id", UpdateHost)
		hosts.DELETE("/:id", DeleteHost)
		hosts.POST("/:id/test", TestHostConnection)
	}
}

// setupContainerRoutes 设置容器管理路由
func setupContainerRoutes(rg *gin.RouterGroup) {
	containers := rg.Group("/containers")
	{
		containers.GET("", ListContainers)
		containers.POST("", CreateContainer)
		containers.GET("/:id", GetContainer)
		containers.POST("/:id/start", StartContainer)
		containers.POST("/:id/stop", StopContainer)
		containers.POST("/:id/restart", RestartContainer)
		containers.DELETE("/:id", RemoveContainer)
		containers.GET("/:id/logs", GetContainerLogs)
		containers.GET("/:id/stats", GetContainerStats)
		containers.POST("/:id/exec", ExecContainer)
	}
}

// setupImageRoutes 设置镜像管理路由
func setupImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	{
		images.GET("", ListImages)
		images.POST("/pull", PullImage)
		images.GET("/search", SearchImages)
		images.GET("/:id", GetImage)
		images.DELETE("/:id", RemoveImage)
		images.POST("/:id/tag", TagImage)
	}
}

// setupVolumeRoutes 设置卷管理路由
func setupVolumeRoutes(rg *gin.RouterGroup) {
	volumes := rg.Group("/volumes")
	{
		volumes.GET("", ListVolumes)
		volumes.POST("", CreateVolume)
		volumes.GET("/:name", GetVolume)
		volumes.DELETE("/:name", RemoveVolume)
	}
}

// setupNetworkRoutes 设置网络管理路由
func setupNetworkRoutes(rg *gin.RouterGroup) {
	networks := rg.Group("/networks")
	{
		networks.GET("", ListNetworks)
		networks.POST("", CreateNetwork)
		networks.GET("/:id", GetNetwork)
		networks.DELETE("/:id", RemoveNetwork)
	}
}

// setupComposeRoutes 设置 Compose 管理路由
func setupComposeRoutes(rg *gin.RouterGroup) {
	compose := rg.Group("/compose")
	{
		compose.GET("/projects", ListComposeProjects)
		compose.POST("/projects", CreateComposeProject)
		compose.GET("/projects/:id", GetComposeProject)
		compose.PUT("/projects/:id", UpdateComposeProject)
		compose.DELETE("/projects/:id", DeleteComposeProject)
		compose.POST("/projects/:id/up", ComposeUp)
		compose.POST("/projects/:id/down", ComposeDown)
		compose.POST("/projects/:id/start", ComposeStart)
		compose.POST("/projects/:id/stop", ComposeStop)
		compose.POST("/projects/:id/restart", ComposeRestart)
		compose.GET("/projects/:id/logs", GetComposeLogs)
		compose.GET("/projects/:id/ps", ComposePs)

		// 目录浏览和上传
		compose.GET("/browse", BrowseDir)
		compose.GET("/scan", ScanComposeFiles)
		compose.POST("/upload", UploadDirectory)
	}
}
