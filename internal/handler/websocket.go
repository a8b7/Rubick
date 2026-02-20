package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"rubick/internal/docker"
	"rubick/internal/repository"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// ContainerLogsWS WebSocket 容器日志
func ContainerLogsWS(c *gin.Context) {
	hostID := c.Query("host_id")
	containerID := c.Param("id")

	if hostID == "" || containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取主机信息
	host, err := repository.GetHostByID(hostID)
	if err != nil {
		sendWSError(conn, "主机不存在")
		return
	}

	executor, err := getExecutor(host)
	if err != nil {
		sendWSError(conn, "创建执行器失败")
		return
	}
	defer executor.Close()

	// 使用 Docker SDK 获取日志
	manager := docker.GetManager()
	dockerConn, err := manager.GetClient(c.Request.Context(), host)
	if err != nil {
		sendWSError(conn, "连接 Docker 失败")
		return
	}

	client, err := dockerConn.Connect(c.Request.Context())
	if err != nil {
		sendWSError(conn, "连接 Docker 失败")
		return
	}

	// 获取容器日志流
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logsReader, err := client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
		Timestamps: false,
	})
	if err != nil {
		sendWSError(conn, "获取日志失败")
		return
	}
	defer logsReader.Close()

	// 启动心跳协程
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.WriteJSON(WebSocketMessage{Type: "ping"}); err != nil {
					return
				}
			}
		}
	}()

	// 读取日志并发送到 WebSocket
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		line := scanner.Text()
		if err := conn.WriteJSON(WebSocketMessage{
			Type:    "log",
			Content: line,
		}); err != nil {
			break
		}
	}

	// 发送结束消息
	conn.WriteJSON(WebSocketMessage{Type: "end"})
}

// ComposeLogsWS WebSocket Compose 日志
func ComposeLogsWS(c *gin.Context) {
	projectID := c.Param("id")

	project, err := repository.GetComposeProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	executor, err := getExecutor(project.Host)
	if err != nil {
		sendWSError(conn, "创建执行器失败")
		return
	}
	defer executor.Close()

	composeService := docker.NewComposeService(executor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := composeService.Logs(ctx, project.Content, docker.LogsOptions{
		ComposeOptions: docker.ComposeOptions{
			ProjectName: project.Name,
		},
		Follow: true,
		Tail:   "100",
	})
	if err != nil {
		sendWSError(conn, "获取日志失败")
		return
	}
	defer stream.Close()

	// 启动心跳协程
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.WriteJSON(WebSocketMessage{Type: "ping"}); err != nil {
					return
				}
			}
		}
	}()

	// 读取日志并发送到 WebSocket
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		if err := conn.WriteJSON(WebSocketMessage{
			Type:    "log",
			Content: line,
		}); err != nil {
			break
		}
	}

	// 发送结束消息
	conn.WriteJSON(WebSocketMessage{Type: "end"})
}

// ExecWebSocketMessage Exec WebSocket 消息
type ExecWebSocketMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Cols    int    `json:"cols"`
	Rows    int    `json:"rows"`
}

// ContainerExecWS WebSocket 容器终端
func ContainerExecWS(c *gin.Context) {
	hostID := c.Query("host_id")
	execID := c.Query("exec_id")

	if hostID == "" || execID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取主机信息
	host, err := repository.GetHostByID(hostID)
	if err != nil {
		sendWSError(conn, "主机不存在")
		return
	}

	// 获取 Docker 客户端
	manager := docker.GetManager()
	dockerConn, err := manager.GetClient(c.Request.Context(), host)
	if err != nil {
		sendWSError(conn, "连接 Docker 失败")
		return
	}

	client, err := dockerConn.Connect(c.Request.Context())
	if err != nil {
		sendWSError(conn, "连接 Docker 失败")
		return
	}

	// 连接到 exec 实例
	svc := docker.NewContainerService(client)
	hijackedResp, err := svc.ExecAttach(c.Request.Context(), execID)
	if err != nil {
		sendWSError(conn, "连接终端失败: "+err.Error())
		return
	}
	defer hijackedResp.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 从 Docker 读取输出并发送到 WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := hijackedResp.Reader.Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					conn.WriteJSON(ExecWebSocketMessage{
						Type:    "data",
						Content: string(buf[:n]),
					})
				}
			}
		}
	}()

	// 从 WebSocket 读取输入并发送到 Docker
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var execMsg ExecWebSocketMessage
		if err := json.Unmarshal(msg, &execMsg); err != nil {
			continue
		}

		switch execMsg.Type {
		case "input":
			hijackedResp.Conn.Write([]byte(execMsg.Content))
		case "resize":
			if execMsg.Cols > 0 && execMsg.Rows > 0 {
				svc.ExecResize(ctx, execID, uint(execMsg.Cols), uint(execMsg.Rows))
			}
		}
	}
}

func sendWSError(conn *websocket.Conn, message string) {
	conn.WriteJSON(WebSocketMessage{
		Type:    "error",
		Content: message,
	})
}
