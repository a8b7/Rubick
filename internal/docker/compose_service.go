package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// ComposeOptions Compose 命令通用选项
type ComposeOptions struct {
	// ProjectName 项目名称
	ProjectName string
	// WorkDir 工作目录
	WorkDir string
	// EnvFile 环境变量文件路径
	EnvFile string
	// ComposeFile compose 文件名（相对于 WorkDir）
	ComposeFile string
	// UseWorkDir 是否使用工作目录模式（不从临时文件读取）
	UseWorkDir bool
}

// UpOptions docker compose up 选项
type UpOptions struct {
	ComposeOptions
	// Build 构建镜像
	Build bool
	// Detach 后台运行
	Detach bool
	// RemoveOrphans 移除孤立容器
	RemoveOrphans bool
	// Timeout 超时时间（秒）
	Timeout int
	// Services 指定服务
	Services []string
}

// DownOptions docker compose down 选项
type DownOptions struct {
	// RemoveImages 移除镜像类型: all, local
	RemoveImages string
	// RemoveVolumes 移除卷
	RemoveVolumes bool
	// RemoveOrphans 移除孤立容器
	RemoveOrphans bool
	// Timeout 超时时间（秒）
	Timeout int
}

// LogsOptions docker compose logs 选项
type LogsOptions struct {
	ComposeOptions
	// Tail 显示最后 N 行
	Tail string
	// Follow 跟踪日志输出
	Follow bool
	// Timestamps 显示时间戳
	Timestamps bool
	// Since 从指定时间开始的日志
	Since string
	// Services 指定服务
	Services []string
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	State     string `json:"state"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	ExitCode  int    `json:"exit_code"`
	Publishers []PortPublisher `json:"publishers,omitempty"`
}

// PortPublisher 端口映射
type PortPublisher struct {
	URL           string `json:"url,omitempty"`
	TargetPort    int    `json:"target_port"`
	PublishedPort int    `json:"published_port"`
	Protocol      string `json:"protocol"`
}

// ComposeService Compose 服务
type ComposeService struct {
	executor CommandExecutor
}

// NewComposeService 创建 Compose 服务
func NewComposeService(executor CommandExecutor) *ComposeService {
	return &ComposeService{
		executor: executor,
	}
}

// Up 启动 Compose 项目
func (s *ComposeService) Up(ctx context.Context, composeYAML string, opts UpOptions) (io.ReadCloser, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		// 目录模式：不需要写入临时文件
		filePath = "" // buildBaseArgs 会使用 WorkDir
		needsCleanup = false
	} else {
		// 内容模式：写入临时文件
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	// 构建命令参数
	args := s.buildBaseArgs(filePath, opts.ComposeOptions)
	args = append(args, "up")

	if opts.Build {
		args = append(args, "--build")
	}
	if opts.RemoveOrphans {
		args = append(args, "--remove-orphans")
	}
	if opts.Timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", opts.Timeout))
	}

	// 如果不是后台模式，添加 --abort-on-container-exit 以便在容器退出时结束
	if !opts.Detach {
		args = append(args, "--abort-on-container-exit")
	}

	// 添加指定服务
	args = append(args, opts.Services...)

	if opts.Detach {
		// 后台模式：直接执行并返回结果
		output, err := s.executor.Execute(ctx, "docker", args...)
		if err != nil {
			if needsCleanup {
				s.cleanupFile(ctx, filePath)
			}
			return nil, err
		}
		// 清理临时文件
		if needsCleanup {
			s.cleanupFile(ctx, filePath)
		}
		return io.NopCloser(strings.NewReader(string(output))), nil
	}

	// 前台模式：返回流式输出
	stream, err := s.executor.ExecuteStream(ctx, "docker", args...)
	if err != nil {
		if needsCleanup {
			s.cleanupFile(ctx, filePath)
		}
		return nil, err
	}

	// 如果不需要清理，直接返回流
	if !needsCleanup {
		return stream, nil
	}

	// 包装流以在关闭时清理临时文件
	return &cleanupReader{
		ReadCloser: stream,
		filePath:   filePath,
		executor:   s.executor,
		ctx:        ctx,
	}, nil
}

// Down 停止并删除 Compose 项目
func (s *ComposeService) Down(ctx context.Context, composeYAML string, opts ComposeOptions, downOpts DownOptions) ([]byte, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		filePath = ""
		needsCleanup = false
	} else {
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	if needsCleanup {
		defer s.cleanupFile(ctx, filePath)
	}

	// 构建命令参数
	args := s.buildBaseArgs(filePath, opts)
	args = append(args, "down")

	if downOpts.RemoveImages != "" {
		args = append(args, "--rmi", downOpts.RemoveImages)
	}
	if downOpts.RemoveVolumes {
		args = append(args, "--volumes")
	}
	if downOpts.RemoveOrphans {
		args = append(args, "--remove-orphans")
	}
	if downOpts.Timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", downOpts.Timeout))
	}

	return s.executor.Execute(ctx, "docker", args...)
}

// Start 启动 Compose 项目的服务
func (s *ComposeService) Start(ctx context.Context, composeYAML string, opts ComposeOptions, services []string) ([]byte, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		filePath = ""
		needsCleanup = false
	} else {
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	if needsCleanup {
		defer s.cleanupFile(ctx, filePath)
	}

	args := s.buildBaseArgs(filePath, opts)
	args = append(args, "start")
	args = append(args, services...)

	return s.executor.Execute(ctx, "docker", args...)
}

// Stop 停止 Compose 项目的服务
func (s *ComposeService) Stop(ctx context.Context, composeYAML string, opts ComposeOptions, timeout int, services []string) ([]byte, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		filePath = ""
		needsCleanup = false
	} else {
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	if needsCleanup {
		defer s.cleanupFile(ctx, filePath)
	}

	args := s.buildBaseArgs(filePath, opts)
	args = append(args, "stop")
	if timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", timeout))
	}
	args = append(args, services...)

	return s.executor.Execute(ctx, "docker", args...)
}

// Restart 重启 Compose 项目的服务
func (s *ComposeService) Restart(ctx context.Context, composeYAML string, opts ComposeOptions, timeout int, services []string) ([]byte, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		filePath = ""
		needsCleanup = false
	} else {
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	if needsCleanup {
		defer s.cleanupFile(ctx, filePath)
	}

	args := s.buildBaseArgs(filePath, opts)
	args = append(args, "restart")
	if timeout > 0 {
		args = append(args, "--timeout", fmt.Sprintf("%d", timeout))
	}
	args = append(args, services...)

	return s.executor.Execute(ctx, "docker", args...)
}

// Logs 获取 Compose 项目日志
func (s *ComposeService) Logs(ctx context.Context, composeYAML string, opts LogsOptions) (io.ReadCloser, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		filePath = ""
		needsCleanup = false
	} else {
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	args := s.buildBaseArgs(filePath, opts.ComposeOptions)
	args = append(args, "logs")

	if opts.Tail != "" {
		args = append(args, "--tail", opts.Tail)
	}
	if opts.Follow {
		args = append(args, "--follow")
	}
	if opts.Timestamps {
		args = append(args, "--timestamps")
	}
	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}
	args = append(args, opts.Services...)

	stream, err := s.executor.ExecuteStream(ctx, "docker", args...)
	if err != nil {
		if needsCleanup {
			s.cleanupFile(ctx, filePath)
		}
		return nil, err
	}

	if !needsCleanup {
		return stream, nil
	}

	return &cleanupReader{
		ReadCloser: stream,
		filePath:   filePath,
		executor:   s.executor,
		ctx:        ctx,
	}, nil
}

// Ps 列出 Compose 项目中的容器状态
func (s *ComposeService) Ps(ctx context.Context, composeYAML string, opts ComposeOptions) ([]ServiceStatus, error) {
	var filePath string
	var needsCleanup bool

	if opts.UseWorkDir {
		filePath = ""
		needsCleanup = false
	} else {
		var err error
		filePath, err = s.executor.WriteFile(ctx, composeYAML, "docker-compose.yml")
		if err != nil {
			return nil, fmt.Errorf("写入 compose 文件失败: %w", err)
		}
		needsCleanup = true
	}

	if needsCleanup {
		defer s.cleanupFile(ctx, filePath)
	}

	args := s.buildBaseArgs(filePath, opts)
	args = append(args, "ps", "--format", "json")

	output, err := s.executor.Execute(ctx, "docker", args...)
	if err != nil {
		return nil, err
	}

	return parsePsOutput(string(output))
}

// buildBaseArgs 构建基础命令参数
func (s *ComposeService) buildBaseArgs(filePath string, opts ComposeOptions) []string {
	var args []string

	if opts.UseWorkDir && opts.WorkDir != "" {
		// 目录模式：使用工作目录中的 compose 文件
		composeFile := opts.ComposeFile
		if composeFile == "" {
			composeFile = "docker-compose.yml"
		}
		composePath := filepath.Join(opts.WorkDir, composeFile)
		args = []string{"compose", "-f", composePath}
	} else {
		// 内容模式：使用临时文件
		args = []string{"compose", "-f", filePath}
	}

	if opts.ProjectName != "" {
		args = append(args, "-p", opts.ProjectName)
	}
	if opts.EnvFile != "" {
		args = append(args, "--env-file", opts.EnvFile)
	}

	return args
}

// cleanupFile 清理临时文件
func (s *ComposeService) cleanupFile(ctx context.Context, filePath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.executor.RemoveFile(ctx, filePath)
}

// parsePsOutput 解析 docker compose ps --format json 输出
func parsePsOutput(output string) ([]ServiceStatus, error) {
	var statuses []ServiceStatus

	// 处理可能的 JSON 数组格式
	output = strings.TrimSpace(output)
	if output == "" {
		return statuses, nil
	}

	// 尝试解析为数组
	if strings.HasPrefix(output, "[") {
		if err := json.Unmarshal([]byte(output), &statuses); err != nil {
			return nil, fmt.Errorf("解析 ps 输出失败: %w", err)
		}
		return statuses, nil
	}

	// 处理每行一个 JSON 对象的格式
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var status ServiceStatus
		if err := json.Unmarshal([]byte(line), &status); err != nil {
			// 尝试解析为嵌套结构
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				continue
			}

			status = ServiceStatus{
				Name:    getStringFromMap(raw, "Name"),
				Command: getStringFromMap(raw, "Command"),
				State:   getStringFromMap(raw, "State"),
				Status:  getStringFromMap(raw, "Status"),
				Health:  getStringFromMap(raw, "Health"),
			}

			if exitCode, ok := raw["ExitCode"].(float64); ok {
				status.ExitCode = int(exitCode)
			}

			if publishers, ok := raw["Publishers"].([]interface{}); ok {
				for _, p := range publishers {
					if pub, ok := p.(map[string]interface{}); ok {
						publisher := PortPublisher{
							URL:      getStringFromMap(pub, "URL"),
							Protocol: getStringFromMap(pub, "Protocol"),
						}
						if tp, ok := pub["TargetPort"].(float64); ok {
							publisher.TargetPort = int(tp)
						}
						if pp, ok := pub["PublishedPort"].(float64); ok {
							publisher.PublishedPort = int(pp)
						}
						status.Publishers = append(status.Publishers, publisher)
					}
				}
			}
		}
		statuses = append(statuses, status)
	}

	return statuses, scanner.Err()
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// cleanupReader 包装 ReadCloser，在关闭时清理临时文件
type cleanupReader struct {
	io.ReadCloser
	filePath string
	executor CommandExecutor
	ctx      context.Context
	closed   bool
}

func (r *cleanupReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true

	err := r.ReadCloser.Close()

	// 清理临时文件
	r.executor.RemoveFile(r.ctx, r.filePath)

	return err
}
