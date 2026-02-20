package handler

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"rubick/internal/docker"
	"rubick/internal/model"
	"rubick/internal/repository"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ListComposeProjects 列出 Compose 项目
func ListComposeProjects(c *gin.Context) {
	hostID := c.Query("host_id")

	projects, err := repository.ListComposeProjects(hostID)
	if err != nil {
		ServerError(c, "获取 Compose 项目列表失败: "+err.Error())
		return
	}

	Success(c, projects)
}

// GetComposeProject 获取 Compose 项目详情
func GetComposeProject(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	Success(c, project)
}

// CreateComposeProject 创建 Compose 项目
func CreateComposeProject(c *gin.Context) {
	var project model.ComposeProject
	if err := c.ShouldBindJSON(&project); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	// 验证必填字段
	if project.Name == "" {
		BadRequest(c, "项目名称不能为空")
		return
	}
	if project.HostID == "" {
		BadRequest(c, "主机 ID 不能为空")
		return
	}

	// 设置默认源类型
	if project.SourceType == "" {
		project.SourceType = "content"
	}

	// 根据源类型验证不同字段
	switch project.SourceType {
	case "content":
		if project.Content == "" {
			BadRequest(c, "Compose 内容不能为空")
			return
		}
	case "directory":
		if project.WorkDir == "" {
			BadRequest(c, "工作目录不能为空")
			return
		}
		// 设置默认 compose 文件名
		if project.ComposeFile == "" {
			project.ComposeFile = "docker-compose.yml"
		}
	default:
		BadRequest(c, "无效的源类型: "+project.SourceType)
		return
	}

	// 验证主机存在
	host, err := repository.GetHostByID(project.HostID)
	if err != nil {
		BadRequest(c, "主机不存在")
		return
	}

	// 对于 directory 模式，验证 compose 文件存在
	if project.SourceType == "directory" {
		executor, err := getExecutor(host)
		if err != nil {
			BadRequest(c, "创建执行器失败: "+err.Error())
			return
		}
		defer executor.Close()

		composeFilePath := filepath.Join(project.WorkDir, project.ComposeFile)
		exists, err := executor.FileExists(c.Request.Context(), composeFilePath)
		if err != nil {
			BadRequest(c, "检查 compose 文件失败: "+err.Error())
			return
		}
		if !exists {
			BadRequest(c, "Compose 文件不存在: "+composeFilePath)
			return
		}
	}

	if err := repository.CreateComposeProject(&project); err != nil {
		ServerError(c, "创建 Compose 项目失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "Compose 项目创建成功", project)
}

// UpdateComposeProject 更新 Compose 项目
func UpdateComposeProject(c *gin.Context) {
	id := c.Param("id")

	var updates model.ComposeProject
	if err := c.ShouldBindJSON(&updates); err != nil {
		BadRequest(c, "无效的请求参数: "+err.Error())
		return
	}

	if err := repository.UpdateComposeProject(id, &updates); err != nil {
		ServerError(c, "更新 Compose 项目失败: "+err.Error())
		return
	}

	project, _ := repository.GetComposeProjectByID(id)
	SuccessWithMessage(c, "Compose 项目更新成功", project)
}

// DeleteComposeProject 删除 Compose 项目
func DeleteComposeProject(c *gin.Context) {
	id := c.Param("id")

	if err := repository.DeleteComposeProject(id); err != nil {
		ServerError(c, "删除 Compose 项目失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "Compose 项目删除成功", nil)
}

// getExecutor 根据主机类型获取对应的命令执行器
func getExecutor(host *model.Host) (docker.CommandExecutor, error) {
	switch host.Type {
	case "local":
		return docker.NewLocalExecutor()
	case "ssh":
		return docker.NewSSHExecutor(&docker.ConnectionConfig{
			Type:          docker.ConnectionTypeSSH,
			Host:          host.Host,
			SSHUser:       host.SSHUser,
			SSHAuthType:   host.SSHAuthType,
			SSHPrivateKey: host.SSHPrivateKey,
			SSHPassword:   host.SSHPassword,
			SSHPort:       host.SSHPort,
		})
	default:
		return nil, fmt.Errorf("不支持的主机类型: %s", host.Type)
	}
}

// getComposeOptions 根据 project 构建 ComposeOptions
func getComposeOptions(project *model.ComposeProject) docker.ComposeOptions {
	opts := docker.ComposeOptions{
		ProjectName: project.Name,
	}

	if project.SourceType == "directory" {
		opts.UseWorkDir = true
		opts.WorkDir = project.WorkDir
		opts.ComposeFile = project.ComposeFile
		opts.EnvFile = project.EnvFile
	}

	return opts
}

// ComposeUp 启动 Compose 项目
func ComposeUp(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	var req struct {
		Build         bool     `json:"build"`
		Detach        bool     `json:"detach"`
		RemoveOrphans bool     `json:"remove_orphans"`
		Timeout       int      `json:"timeout"`
		Services      []string `json:"services"`
	}
	c.ShouldBindJSON(&req)

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	opts := docker.UpOptions{
		ComposeOptions: getComposeOptions(project),
		Build:          req.Build,
		Detach:         req.Detach,
		RemoveOrphans:  req.RemoveOrphans,
		Timeout:        req.Timeout,
		Services:       req.Services,
	}

	if req.Detach {
		// 后台模式
		_, err = composeService.Up(ctx, project.Content, opts)
		if err != nil {
			ServerError(c, "启动 Compose 项目失败: "+err.Error())
			return
		}

		// 更新项目状态
		repository.UpdateComposeProjectStatus(id, "running")

		SuccessWithMessage(c, "Compose 项目已启动", nil)
	} else {
		// 流式输出模式
		stream, err := composeService.Up(ctx, project.Content, opts)
		if err != nil {
			ServerError(c, "启动 Compose 项目失败: "+err.Error())
			return
		}
		defer stream.Close()

		// 设置 SSE 响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// 流式传输输出
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", line)))
			c.Writer.Flush()
		}

		if err := scanner.Err(); err != nil {
			c.Writer.Write([]byte(fmt.Sprintf("data: ERROR: %s\n\n", err.Error())))
			c.Writer.Flush()
		}

		// 更新项目状态
		repository.UpdateComposeProjectStatus(id, "running")
	}
}

// ComposeDown 停止并删除 Compose 项目
func ComposeDown(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	var req struct {
		RemoveImages  string `json:"remove_images"` // all, local
		RemoveVolumes bool   `json:"remove_volumes"`
		RemoveOrphans bool   `json:"remove_orphans"`
		Timeout       int    `json:"timeout"`
	}
	c.ShouldBindJSON(&req)

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	output, err := composeService.Down(ctx, project.Content, getComposeOptions(project), docker.DownOptions{
		RemoveImages:  req.RemoveImages,
		RemoveVolumes: req.RemoveVolumes,
		RemoveOrphans: req.RemoveOrphans,
		Timeout:       req.Timeout,
	})
	if err != nil {
		ServerError(c, "停止 Compose 项目失败: "+err.Error())
		return
	}

	// 更新项目状态
	repository.UpdateComposeProjectStatus(id, "stopped")

	SuccessWithMessage(c, "Compose 项目已停止并删除", gin.H{
		"output": string(output),
	})
}

// ComposeStart 启动 Compose 项目（不创建新容器）
func ComposeStart(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	var req struct {
		Services []string `json:"services"`
	}
	c.ShouldBindJSON(&req)

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	output, err := composeService.Start(ctx, project.Content, getComposeOptions(project), req.Services)
	if err != nil {
		ServerError(c, "启动 Compose 项目失败: "+err.Error())
		return
	}

	// 更新项目状态
	repository.UpdateComposeProjectStatus(id, "running")

	SuccessWithMessage(c, "Compose 项目已启动", gin.H{
		"output": string(output),
	})
}

// ComposeStop 停止 Compose 项目
func ComposeStop(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	var req struct {
		Timeout  int      `json:"timeout"`
		Services []string `json:"services"`
	}
	c.ShouldBindJSON(&req)

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	output, err := composeService.Stop(ctx, project.Content, getComposeOptions(project), req.Timeout, req.Services)
	if err != nil {
		ServerError(c, "停止 Compose 项目失败: "+err.Error())
		return
	}

	// 更新项目状态
	repository.UpdateComposeProjectStatus(id, "stopped")

	SuccessWithMessage(c, "Compose 项目已停止", gin.H{
		"output": string(output),
	})
}

// ComposeRestart 重启 Compose 项目
func ComposeRestart(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	var req struct {
		Timeout  int      `json:"timeout"`
		Services []string `json:"services"`
	}
	c.ShouldBindJSON(&req)

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	output, err := composeService.Restart(ctx, project.Content, getComposeOptions(project), req.Timeout, req.Services)
	if err != nil {
		ServerError(c, "重启 Compose 项目失败: "+err.Error())
		return
	}

	// 更新项目状态
	repository.UpdateComposeProjectStatus(id, "running")

	SuccessWithMessage(c, "Compose 项目已重启", gin.H{
		"output": string(output),
	})
}

// GetComposeLogs 获取 Compose 项目日志
func GetComposeLogs(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	tail := c.DefaultQuery("tail", "100")
	follow := c.Query("follow") == "true"
	timestamps := c.Query("timestamps") == "true"
	since := c.Query("since")
	services := c.QueryArray("services")

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx := c.Request.Context()

	opts := docker.LogsOptions{
		ComposeOptions: getComposeOptions(project),
		Tail:           tail,
		Follow:         follow,
		Timestamps:     timestamps,
		Since:          since,
		Services:       services,
	}

	stream, err := composeService.Logs(ctx, project.Content, opts)
	if err != nil {
		ServerError(c, "获取 Compose 日志失败: "+err.Error())
		return
	}
	defer stream.Close()

	if follow {
		// 流式输出模式
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", line)))
			c.Writer.Flush()
		}
	} else {
		// 一次性返回所有日志
		c.Header("Content-Type", "text/plain")
		io.Copy(c.Writer, stream)
	}
}

// ComposePs 列出 Compose 项目中的容器
func ComposePs(c *gin.Context) {
	id := c.Param("id")

	project, err := repository.GetComposeProjectByID(id)
	if err != nil {
		NotFound(c, "Compose 项目不存在")
		return
	}

	executor, err := getExecutor(project.Host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	statuses, err := composeService.Ps(ctx, project.Content, getComposeOptions(project))
	if err != nil {
		ServerError(c, "获取 Compose 容器状态失败: "+err.Error())
		return
	}

	// 根据容器状态更新项目状态
	projectStatus := "stopped"
	for _, s := range statuses {
		if s.State == "running" {
			projectStatus = "running"
			break
		}
	}
	repository.UpdateComposeProjectStatus(id, projectStatus)

	Success(c, statuses)
}

// BrowseDir 浏览服务器目录
func BrowseDir(c *gin.Context) {
	hostID := c.Query("host_id")
	path := c.Query("path")

	if hostID == "" {
		BadRequest(c, "host_id 参数不能为空")
		return
	}

	// 默认从根目录开始
	if path == "" {
		path = "/"
	}

	// 获取主机信息
	host, err := repository.GetHostByID(hostID)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}

	executor, err := getExecutor(host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	files, err := executor.ListDir(ctx, path)
	if err != nil {
		ServerError(c, "浏览目录失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"path":  path,
		"files": files,
	})
}

// ScanComposeFiles 扫描目录中的 compose 文件
func ScanComposeFiles(c *gin.Context) {
	hostID := c.Query("host_id")
	path := c.Query("path")

	if hostID == "" {
		BadRequest(c, "host_id 参数不能为空")
		return
	}
	if path == "" {
		BadRequest(c, "path 参数不能为空")
		return
	}

	// 获取主机信息
	host, err := repository.GetHostByID(hostID)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}

	executor, err := getExecutor(host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// 列出目录中的文件
	files, err := executor.ListDir(ctx, path)
	if err != nil {
		ServerError(c, "扫描目录失败: "+err.Error())
		return
	}

	// 过滤出 compose 文件
	var composeFiles []string
	composePatterns := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, file := range files {
		if file.IsDir {
			continue
		}
		for _, pattern := range composePatterns {
			if file.Name == pattern || strings.HasSuffix(file.Name, ".compose.yml") || strings.HasSuffix(file.Name, ".compose.yaml") {
				composeFiles = append(composeFiles, file.Name)
				break
			}
		}
	}

	// 同时查找 .env 文件
	var envFiles []string
	for _, file := range files {
		if file.IsDir {
			continue
		}
		if file.Name == ".env" || strings.HasSuffix(file.Name, ".env") {
			envFiles = append(envFiles, file.Name)
		}
	}

	Success(c, gin.H{
		"path":          path,
		"compose_files": composeFiles,
		"env_files":     envFiles,
	})
}

// UploadDirectory 上传目录到服务器
func UploadDirectory(c *gin.Context) {
	hostID := c.PostForm("host_id")
	targetPath := c.PostForm("target_path")

	if hostID == "" {
		BadRequest(c, "host_id 参数不能为空")
		return
	}
	if targetPath == "" {
		BadRequest(c, "target_path 参数不能为空")
		return
	}

	// 获取主机信息
	host, err := repository.GetHostByID(hostID)
	if err != nil {
		NotFound(c, "主机不存在")
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		BadRequest(c, "解析表单失败: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		BadRequest(c, "没有上传文件")
		return
	}

	executor, err := getExecutor(host)
	if err != nil {
		ServerError(c, "创建执行器失败: "+err.Error())
		return
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// 创建目标目录
	if err := executor.MkdirAll(ctx, targetPath); err != nil {
		ServerError(c, "创建目标目录失败: "+err.Error())
		return
	}

	// 上传每个文件
	uploadedFiles := []string{}
	for _, file := range files {
		// 打开上传的文件
		src, err := file.Open()
		if err != nil {
			ServerError(c, "打开上传文件失败: "+err.Error())
			return
		}

		// 读取文件内容
		content := make([]byte, file.Size)
		_, err = src.Read(content)
		src.Close()
		if err != nil {
			ServerError(c, "读取上传文件失败: "+err.Error())
			return
		}

		// 构建目标路径（保留相对路径结构）
		relativePath := file.Filename
		targetFilePath := filepath.Join(targetPath, relativePath)

		// 确保父目录存在
		targetDir := filepath.Dir(targetFilePath)
		if err := executor.MkdirAll(ctx, targetDir); err != nil {
			ServerError(c, "创建子目录失败: "+err.Error())
			return
		}

		// 写入文件
		if err := executor.WriteFileToPath(ctx, string(content), targetFilePath); err != nil {
			ServerError(c, "写入文件失败: "+err.Error())
			return
		}

		uploadedFiles = append(uploadedFiles, relativePath)
	}

	SuccessWithMessage(c, "目录上传成功", gin.H{
		"path":           targetPath,
		"uploaded_files": uploadedFiles,
		"count":          len(uploadedFiles),
	})
}
