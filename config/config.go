package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `json:"server"`
	Music  MusicConfig  `json:"music"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// MusicConfig 音乐相关配置
type MusicConfig struct {
	Directory string `json:"directory"` // 音乐文件目录
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	// 如果没有指定配置文件路径,使用默认配置
	if configPath == "" {
		return GetDefaultConfig(), nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 转换相对路径为绝对路径
	if !filepath.IsAbs(cfg.Music.Directory) {
		absPath, err := filepath.Abs(cfg.Music.Directory)
		if err == nil {
			cfg.Music.Directory = absPath
		}
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		globalConfig = GetDefaultConfig()
	}
	return globalConfig
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	musicDir, _ := filepath.Abs("./backend/music")
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Music: MusicConfig{
			Directory: musicDir,
		},
	}
}