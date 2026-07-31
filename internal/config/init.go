package config

import (
	"log"
	"os"
)

// InitConfig 初始化配置系统
// 自动检测环境并加载对应配置文件
func InitConfig() *Config {
	// 获取环境变量
	env := GetEnv("APP_ENV", "development")

	// 根据环境加载配置文件
	var configPath string
	switch env {
	case "production":
		configPath = "configs/production.yml"
	case "test":
		configPath = "configs/test.yml"
	case "development":
		configPath = "configs/development.yml"
	default:
		// 默认配置文件查找顺序
		possiblePaths := []string{
			"configs/default.yml",
			"config.yml", // 向后兼容
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
		if configPath == "" {
			configPath = "configs/default.yml"
		}
	}

	// 如果指定的配置文件不存在，尝试回退方案
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("警告: 配置文件 %s 不存在", configPath)

		// 尝试回退到默认配置文件
		fallbackPaths := []string{
			"configs/default.yml",
			"config.yml",
		}

		for _, path := range fallbackPaths {
			if _, err := os.Stat(path); err == nil {
				log.Printf("使用回退配置文件: %s", path)
				configPath = path
				break
			}
		}
	}

	// 加载配置
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	log.Printf("成功加载 %s 环境配置 (文件: %s)", env, configPath)
	return config
}

// MustGetConfig 获取全局配置，如果未初始化则panic
func MustGetConfig() *Config {
	if GlobalConfig == nil {
		log.Fatal("配置尚未初始化，请先调用 InitConfig()")
	}
	return GlobalConfig
}

// GetConfig 安全获取全局配置
func GetConfig() (*Config, bool) {
	if GlobalConfig == nil {
		return nil, false
	}
	return GlobalConfig, true
}
