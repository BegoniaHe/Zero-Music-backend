package handlers

import (
	"net/http"
	"os"
	"zero-music/config"

	"github.com/gin-gonic/gin"
)

const (
	// APIName 是 API 的名称。
	APIName = "zero music API"
	// APIVersion 是 API 的版本号。
	// 可通过构建时 -ldflags 覆盖，例如：
	// go build -ldflags "-X zero-music/handlers.APIVersion=1.2.3"
	APIVersion = "1.0.0"
)

// SystemHandler 负责处理与系统相关的 API 请求。
type SystemHandler struct {
	cfg *config.Config
}

// NewSystemHandler 创建一个新的 SystemHandler 实例。
func NewSystemHandler(cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		cfg: cfg,
	}
}

// HealthCheck 处理健康检查请求。
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	// 检查音乐目录是否可访问。
	musicDirAccessible := true
	if _, err := os.Stat(h.cfg.Music.Directory); err != nil {
		musicDirAccessible = false
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !musicDirAccessible {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":               status,
		"message":              "zero music服务器正在运行",
		"music_dir_accessible": musicDirAccessible,
		"music_directory":      h.cfg.Music.Directory,
	})
}

// APIIndex 处理根请求并列出可用的端点。
func (h *SystemHandler) APIIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":    APIName,
		"version": APIVersion,
		"endpoints": []string{
			"GET /health - 健康检查",
			"GET /api/songs - 获取所有歌曲列表",
			"GET /api/song/:id - 获取指定歌曲信息",
			"GET /api/stream/:id - 流式传输音频",
		},
	})
}
