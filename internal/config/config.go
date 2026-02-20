package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

// DockerConfig Docker 配置
type DockerConfig struct {
	DefaultHost string `mapstructure:"default_host"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json, text
}

var cfg *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 配置文件设置
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/rubick")
	}

	// 环境变量支持
	v.SetEnvPrefix("RUBICK")
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时使用默认值
	}

	// 解析配置
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// 服务器配置
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")

	// 数据库配置
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.path", "./data/rubick.db")

	// Docker 配置
	v.SetDefault("docker.default_host", "local")

	// 日志配置
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

// Get 获取当前配置
func Get() *Config {
	if cfg == nil {
		// 尝试加载默认配置
		var err error
		cfg, err = Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v，使用默认配置\n", err)
			cfg = &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
					Mode: "debug",
				},
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: "./data/rubick.db",
				},
				Docker: DockerConfig{
					DefaultHost: "local",
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			}
		}
	}
	return cfg
}

// Addr 返回服务器监听地址
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
