package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rubick/internal/config"
	"rubick/internal/model"
)

var db *gorm.DB

// Initialize 初始化数据库连接
func Initialize(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var err error

	// 确保数据目录存在
	dbPath := cfg.Path
	dbDir := filepath.Dir(dbPath)
	if err = os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 配置 GORM 日志
	logLevel := logger.Info

	// 连接数据库
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移
	if err = autoMigrate(); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化默认数据
	if err = initDefaultData(); err != nil {
		return nil, fmt.Errorf("初始化默认数据失败: %w", err)
	}

	return db, nil
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate() error {
	return db.AutoMigrate(
		&model.Host{},
		&model.Certificate{},
		&model.ComposeProject{},
		&model.AuditLog{},
	)
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 检查是否已存在本地主机
	var count int64
	db.Model(&model.Host{}).Where("type = ?", "local").Count(&count)
	if count == 0 {
		// 创建默认本地主机
		localHost := &model.Host{
			Name:        "local",
			Type:        "local",
			IsDefault:   true,
			IsActive:    true,
			Description: "本地 Docker 主机",
		}
		if err := db.Create(localHost).Error; err != nil {
			return fmt.Errorf("创建默认本地主机失败: %w", err)
		}
	}
	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
