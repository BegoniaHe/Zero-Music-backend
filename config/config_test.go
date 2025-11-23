package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, cfg *Config) string {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("无法序列化配置: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("无法写入配置文件: %v", err)
	}
	return configPath
}

func TestLoadAppliesEnvOverrides(t *testing.T) {
	musicDir := t.TempDir()
	cfgPath := writeConfigFile(t, &Config{
		Server: ServerConfig{
			Host:         "127.0.0.1",
			Port:         8081,
			MaxRangeSize: 1024,
		},
		Music: MusicConfig{
			Directory:        musicDir,
			SupportedFormats: []string{".mp3"},
			CacheTTLMinutes:  10,
		},
	})

	t.Setenv("ZERO_MUSIC_SERVER_PORT", "9090")
	t.Setenv("ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS", "30")
	t.Setenv("ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS", "45")
	t.Setenv("ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS", "90")
	t.Setenv("ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS", "40")
	t.Setenv("ZERO_MUSIC_MAX_RANGE_SIZE", "2048")
	t.Setenv("ZERO_MUSIC_CACHE_TTL_MINUTES", "30")
	t.Setenv("ZERO_MUSIC_MUSIC_DIRECTORY", musicDir)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("期望加载成功, 但出现错误: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Fatalf("期望端口 9090, 实际 %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeoutSeconds != 30 {
		t.Fatalf("期望 ReadTimeoutSeconds=30, 实际 %d", cfg.Server.ReadTimeoutSeconds)
	}
	if cfg.Server.WriteTimeoutSeconds != 45 {
		t.Fatalf("期望 WriteTimeoutSeconds=45, 实际 %d", cfg.Server.WriteTimeoutSeconds)
	}
	if cfg.Server.IdleTimeoutSeconds != 90 {
		t.Fatalf("期望 IdleTimeoutSeconds=90, 实际 %d", cfg.Server.IdleTimeoutSeconds)
	}
	if cfg.Server.ShutdownTimeoutSeconds != 40 {
		t.Fatalf("期望 ShutdownTimeoutSeconds=40, 实际 %d", cfg.Server.ShutdownTimeoutSeconds)
	}
	if cfg.Server.MaxRangeSize != 2048 {
		t.Fatalf("期望 MaxRangeSize=2048, 实际 %d", cfg.Server.MaxRangeSize)
	}
	if cfg.Music.CacheTTLMinutes != 30 {
		t.Fatalf("期望 CacheTTLMinutes=30, 实际 %d", cfg.Music.CacheTTLMinutes)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	musicDir := t.TempDir()
	cfgPath := writeConfigFile(t, &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         70000,
			MaxRangeSize: DefaultMaxRangeSize,
		},
		Music: MusicConfig{
			Directory:        musicDir,
			SupportedFormats: []string{".mp3"},
			CacheTTLMinutes:  DefaultCacheTTLMinutes,
		},
	})

	if _, err := Load(cfgPath); err == nil {
		t.Fatal("端口超过范围时应返回错误")
	}
}
