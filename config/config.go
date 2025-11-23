package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	// 默认服务器设置
	DefaultMaxRangeSize           = 100 * 1024 * 1024
	DefaultCacheTTLMinutes        = 5
	DefaultServerHost             = "0.0.0.0"
	DefaultServerPort             = 8080
	DefaultReadTimeoutSeconds     = 15
	DefaultWriteTimeoutSeconds    = 60
	DefaultIdleTimeoutSeconds     = 120
	DefaultShutdownTimeoutSeconds = 30

	// 约束
	MaxAllowedRangeSize              = 500 * 1024 * 1024
	MaxAllowedCacheTTL               = 1440
	MaxAllowedTimeoutSeconds         = 600
	MaxAllowedShutdownTimeoutSeconds = 300
)

// Config 定义了应用程序的所有配置项。
type Config struct {
	Server ServerConfig `json:"server"`
	Music  MusicConfig  `json:"music"`
}

// ServerConfig 定义了服务器相关的配置。
type ServerConfig struct {
	Host                   string `json:"host"`
	Port                   int    `json:"port"`
	MaxRangeSize           int64  `json:"max_range_size"`
	ReadTimeoutSeconds     int    `json:"read_timeout_seconds"`
	WriteTimeoutSeconds    int    `json:"write_timeout_seconds"`
	IdleTimeoutSeconds     int    `json:"idle_timeout_seconds"`
	ShutdownTimeoutSeconds int    `json:"shutdown_timeout_seconds"`
}

// MusicConfig 定义了音乐库相关的配置。
type MusicConfig struct {
	Directory        string   `json:"directory"`
	SupportedFormats []string `json:"supported_formats"`
	CacheTTLMinutes  int      `json:"cache_ttl_minutes"`
}

// Load 从指定路径加载配置文件，如果为空则返回默认配置。
func Load(configPath string) (*Config, error) {
	var cfg *Config
	if configPath == "" {
		cfg = GetDefaultConfig()
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		cfg = &Config{}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
		ensureDefaults(cfg)
	}

	applyEnvOverrides(cfg)
	cfg.Music.Directory = ensureAbsolutePath(cfg.Music.Directory)

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// ensureDefaults 为缺失字段填充默认值。
func ensureDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = DefaultServerHost
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = DefaultServerPort
	}
	if cfg.Server.MaxRangeSize == 0 {
		cfg.Server.MaxRangeSize = DefaultMaxRangeSize
	}
	if cfg.Server.ReadTimeoutSeconds <= 0 {
		cfg.Server.ReadTimeoutSeconds = DefaultReadTimeoutSeconds
	}
	if cfg.Server.WriteTimeoutSeconds <= 0 {
		cfg.Server.WriteTimeoutSeconds = DefaultWriteTimeoutSeconds
	}
	if cfg.Server.IdleTimeoutSeconds <= 0 {
		cfg.Server.IdleTimeoutSeconds = DefaultIdleTimeoutSeconds
	}
	if cfg.Server.ShutdownTimeoutSeconds <= 0 {
		cfg.Server.ShutdownTimeoutSeconds = DefaultShutdownTimeoutSeconds
	}
	if len(cfg.Music.SupportedFormats) == 0 {
		cfg.Music.SupportedFormats = []string{".mp3", ".flac", ".wav", ".m4a", ".ogg"}
	}
	if cfg.Music.CacheTTLMinutes <= 0 {
		cfg.Music.CacheTTLMinutes = DefaultCacheTTLMinutes
	}
	if cfg.Music.Directory == "" {
		cfg.Music.Directory = determineDefaultMusicDirectory()
	}
}

// applyEnvOverrides 使用环境变量覆盖配置。
func applyEnvOverrides(cfg *Config) {
	ensureDefaults(cfg)

	if host := os.Getenv("ZERO_MUSIC_SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := parseEnvInt("ZERO_MUSIC_SERVER_PORT", 1, 65535); port != nil {
		cfg.Server.Port = *port
	}
	if maxRange := parseEnvInt("ZERO_MUSIC_MAX_RANGE_SIZE", 1, int(MaxAllowedRangeSize)); maxRange != nil {
		cfg.Server.MaxRangeSize = int64(*maxRange)
	}
	if readTimeout := parseEnvInt("ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS", 1, MaxAllowedTimeoutSeconds); readTimeout != nil {
		cfg.Server.ReadTimeoutSeconds = *readTimeout
	}
	if writeTimeout := parseEnvInt("ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS", 1, MaxAllowedTimeoutSeconds); writeTimeout != nil {
		cfg.Server.WriteTimeoutSeconds = *writeTimeout
	}
	if idleTimeout := parseEnvInt("ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS", 1, MaxAllowedTimeoutSeconds); idleTimeout != nil {
		cfg.Server.IdleTimeoutSeconds = *idleTimeout
	}
	if shutdownTimeout := parseEnvInt("ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS", 1, MaxAllowedShutdownTimeoutSeconds); shutdownTimeout != nil {
		cfg.Server.ShutdownTimeoutSeconds = *shutdownTimeout
	}

	if musicDir := os.Getenv("ZERO_MUSIC_MUSIC_DIRECTORY"); musicDir != "" {
		cfg.Music.Directory = ensureAbsolutePath(musicDir)
	}
	if cacheTTL := parseEnvInt("ZERO_MUSIC_CACHE_TTL_MINUTES", 1, MaxAllowedCacheTTL); cacheTTL != nil {
		cfg.Music.CacheTTLMinutes = *cacheTTL
	}
}

func parseEnvInt(key string, min, max int) *int {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	if value < min || value > max {
		return nil
	}
	return &value
}

// validateConfig 验证配置合法性。
func validateConfig(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("端口必须在 1-65535 范围内，当前值: %d", cfg.Server.Port)
	}
	if cfg.Server.MaxRangeSize < 1 || cfg.Server.MaxRangeSize > MaxAllowedRangeSize {
		return fmt.Errorf("MaxRangeSize 必须在 1-%d 范围内，当前值: %d", MaxAllowedRangeSize, cfg.Server.MaxRangeSize)
	}
	if cfg.Server.ReadTimeoutSeconds < 1 || cfg.Server.ReadTimeoutSeconds > MaxAllowedTimeoutSeconds {
		return fmt.Errorf("ReadTimeoutSeconds 必须在 1-%d 范围内", MaxAllowedTimeoutSeconds)
	}
	if cfg.Server.WriteTimeoutSeconds < 1 || cfg.Server.WriteTimeoutSeconds > MaxAllowedTimeoutSeconds {
		return fmt.Errorf("WriteTimeoutSeconds 必须在 1-%d 范围内", MaxAllowedTimeoutSeconds)
	}
	if cfg.Server.IdleTimeoutSeconds < 1 || cfg.Server.IdleTimeoutSeconds > MaxAllowedTimeoutSeconds {
		return fmt.Errorf("IdleTimeoutSeconds 必须在 1-%d 范围内", MaxAllowedTimeoutSeconds)
	}
	if cfg.Server.ShutdownTimeoutSeconds < 1 || cfg.Server.ShutdownTimeoutSeconds > MaxAllowedShutdownTimeoutSeconds {
		return fmt.Errorf("ShutdownTimeoutSeconds 必须在 1-%d 范围内", MaxAllowedShutdownTimeoutSeconds)
	}
	if cfg.Music.CacheTTLMinutes < 1 || cfg.Music.CacheTTLMinutes > MaxAllowedCacheTTL {
		return fmt.Errorf("CacheTTLMinutes 必须在 1-%d 范围内，当前值: %d", MaxAllowedCacheTTL, cfg.Music.CacheTTLMinutes)
	}
	if cfg.Music.Directory == "" {
		return fmt.Errorf("音乐目录不能为空")
	}
	if _, err := os.Stat(cfg.Music.Directory); err != nil {
		return fmt.Errorf("音乐目录不可访问: %w", err)
	}
	return nil
}

// ensureAbsolutePath 将路径转换为绝对路径。
func ensureAbsolutePath(path string) string {
	if path == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

// determineDefaultMusicDirectory 返回最合理的默认音乐目录。
func determineDefaultMusicDirectory() string {
	candidates := []string{}

	if envDir := os.Getenv("ZERO_MUSIC_MUSIC_DIRECTORY"); envDir != "" {
		candidates = append(candidates, envDir)
	}
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		candidates = append(candidates, filepath.Join(homeDir, "Music"))
	}
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		candidates = append(candidates, filepath.Join(cwd, "music"))
	}
	candidates = append(candidates, filepath.Join(os.TempDir(), "zero-music"))

	for _, candidate := range candidates {
		abs := ensureAbsolutePath(candidate)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
		if err := os.MkdirAll(abs, 0o755); err == nil {
			return abs
		}
	}

	return ensureAbsolutePath(candidates[len(candidates)-1])
}

// GetDefaultConfig 返回包含默认设置的配置实例。
func GetDefaultConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:                   DefaultServerHost,
			Port:                   DefaultServerPort,
			MaxRangeSize:           DefaultMaxRangeSize,
			ReadTimeoutSeconds:     DefaultReadTimeoutSeconds,
			WriteTimeoutSeconds:    DefaultWriteTimeoutSeconds,
			IdleTimeoutSeconds:     DefaultIdleTimeoutSeconds,
			ShutdownTimeoutSeconds: DefaultShutdownTimeoutSeconds,
		},
		Music: MusicConfig{
			Directory:        determineDefaultMusicDirectory(),
			SupportedFormats: []string{".mp3", ".flac", ".wav", ".m4a", ".ogg"},
			CacheTTLMinutes:  DefaultCacheTTLMinutes,
		},
	}
	return cfg
}
