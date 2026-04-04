package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 全局配置结构体
type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Server    ServerConfig    `yaml:"server"`
	Admin     AdminConfig     `yaml:"admin"`
	Upload    UploadConfig    `yaml:"upload"`
	Session   SessionConfig   `yaml:"session"`
	Static    StaticConfig    `yaml:"static"`
	Templates TemplatesConfig `yaml:"templates"`
	Docker    DockerConfig    `yaml:"docker"`
	Logging   LoggingConfig   `yaml:"logging"`
	Security  SecurityConfig  `yaml:"security"`
	App       AppConfig       `yaml:"app"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Name      string `yaml:"name"`
	Charset   string `yaml:"charset"`
	ParseTime bool   `yaml:"parseTime"`
	Location  string `yaml:"location"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	Directory    string   `yaml:"directory"`
	MaxSize      int64    `yaml:"maxSize"`
	AllowedTypes []string `yaml:"allowedTypes"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	CookieName  string `yaml:"cookieName"`
	CookieValue string `yaml:"cookieValue"`
	CookiePath  string `yaml:"cookiePath"`
	MaxAge      int    `yaml:"maxAge"`
}

// StaticConfig 静态文件配置
type StaticConfig struct {
	Directory string `yaml:"directory"`
	URLPrefix string `yaml:"urlPrefix"`
}

// TemplatesConfig 模板配置
type TemplatesConfig struct {
	Directory string `yaml:"directory"`
}

// DockerConfig Docker配置
type DockerConfig struct {
	MySQL DockerMySQLConfig `yaml:"mysql"`
	App   DockerAppConfig   `yaml:"app"`
}

type DockerMySQLConfig struct {
	RootPassword string `yaml:"rootPassword"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	Collation    string `yaml:"collation"`
	Port         string `yaml:"port"`
}

type DockerAppConfig struct {
	Port string `yaml:"port"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableCORS      bool     `yaml:"enableCORS"`
	AllowedOrigins  []string `yaml:"allowedOrigins"`
	EnableRateLimit bool     `yaml:"enableRateLimit"`
	RateLimit       RateLimit `yaml:"rateLimit"`
}

type RateLimit struct {
	Requests int `yaml:"requests"`
	Window   int `yaml:"window"`
}

// AppConfig 应用信息配置
type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
}

// 全局配置实例
var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 确定配置文件路径
	if configPath == "" {
		env := GetEnv("APP_ENV", "development")
		switch env {
		case "production":
			configPath = "config/production.yml"
		case "test":
			configPath = "config/test.yml"
		case "development":
			configPath = "config/development.yml"
		default:
			// 默认配置文件顺序：环境配置 -> 默认配置 -> 根目录兼容配置
			possiblePaths := []string{
				"config/default.yml",
				"config.yml", // 向后兼容
			}
			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					configPath = path
					break
				}
			}
			if configPath == "" {
				configPath = "config/default.yml" // 默认使用新位置
			}
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 环境变量覆盖
	overrideWithEnv(config)

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 设置全局配置
	GlobalConfig = config

	log.Printf("配置文件加载成功: %s", configPath)
	return config, nil
}

// overrideWithEnv 使用环境变量覆盖配置
func overrideWithEnv(config *Config) {
	// 数据库配置
	if val := GetEnv("DB_HOST", ""); val != "" {
		config.Database.Host = val
	}
	if val := GetEnv("DB_PORT", ""); val != "" {
		config.Database.Port = val
	}
	if val := GetEnv("DB_USER", ""); val != "" {
		config.Database.User = val
	}
	if val := GetEnv("DB_PASSWORD", ""); val != "" {
		config.Database.Password = val
	}
	if val := GetEnv("DB_NAME", ""); val != "" {
		config.Database.Name = val
	}

	// 服务器配置
	if val := GetEnv("SERVER_PORT", ""); val != "" {
		config.Server.Port = val
	}
	if val := GetEnv("SERVER_HOST", ""); val != "" {
		config.Server.Host = val
	}

	// 管理员配置
	if val := GetEnv("ADMIN_USERNAME", ""); val != "" {
		config.Admin.Username = val
	}
	if val := GetEnv("ADMIN_PASSWORD", ""); val != "" {
		config.Admin.Password = val
	}

	// 上传配置
	if val := GetEnv("UPLOAD_DIRECTORY", ""); val != "" {
		config.Upload.Directory = val
	}
	if val := GetEnv("UPLOAD_MAX_SIZE", ""); val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.Upload.MaxSize = size
		}
	}

	// 日志配置
	if val := GetEnv("LOG_LEVEL", ""); val != "" {
		config.Logging.Level = val
	}
	if val := GetEnv("LOG_FILE", ""); val != "" {
		config.Logging.File = val
	}
}

// validateConfig 验证配置的有效性
func validateConfig(config *Config) error {
	// 验证必要的配置项
	if config.Database.Host == "" {
		return fmt.Errorf("数据库主机地址不能为空")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if config.Server.Port == "" {
		return fmt.Errorf("服务器端口不能为空")
	}
	if config.Admin.Username == "" {
		return fmt.Errorf("管理员用户名不能为空")
	}
	if config.Admin.Password == "" {
		return fmt.Errorf("管理员密码不能为空")
	}
	if config.Upload.Directory == "" {
		return fmt.Errorf("上传目录不能为空")
	}

	// 验证端口格式
	if port, err := strconv.Atoi(config.Server.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("服务器端口格式不正确: %s", config.Server.Port)
	}
	if port, err := strconv.Atoi(config.Database.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("数据库端口格式不正确: %s", config.Database.Port)
	}

	return nil
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.Charset,
		c.Database.ParseTime,
		c.Database.Location,
	)
}

// GetServerAddr 获取服务器监听地址
func (c *Config) GetServerAddr() string {
	return c.Server.Host + ":" + c.Server.Port
}

// IsAllowedFileType 检查文件类型是否允许上传
func (c *Config) IsAllowedFileType(fileType string) bool {
	// 如果没有限制，则允许所有类型
	if len(c.Upload.AllowedTypes) == 0 {
		return true
	}

	// 检查是否在允许列表中
	fileType = strings.ToLower(fileType)
	for _, allowedType := range c.Upload.AllowedTypes {
		if strings.ToLower(allowedType) == fileType {
			return true
		}
	}
	return false
}