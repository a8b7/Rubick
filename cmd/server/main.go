package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rubick/internal/config"
	"rubick/internal/database"
	"rubick/internal/docker"
	"rubick/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Version 版本信息
var Version = "dev"

// BuildTime 构建时间
var BuildTime = "unknown"

func main() {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	// 解析命令行参数
	configPath := flag.String("config", "", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Rubick Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db, err := database.Initialize(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 输出启动信息
	log.Printf("Rubick %s 启动中...", Version)
	log.Printf("数据库: %s", cfg.Database.Path)

	// 初始化 Docker 客户端管理器
	dockerManager := docker.GetManager()
	defer dockerManager.CloseAll()

	// 创建路由
	router := handler.NewRouter()
	engine := router.Setup()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器（非阻塞）
	go func() {
		log.Printf("服务器监听: %s", cfg.Server.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 输出启动完成信息
	log.Printf("Rubick 已启动，访问 http://%s", cfg.Server.Addr())
	log.Printf("API 文档: http://%s/api/v1/health", cfg.Server.Addr())

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭错误: %v", err)
	}

	// 关闭数据库连接
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("服务器已关闭")

	_ = ctx // 避免未使用警告
}
