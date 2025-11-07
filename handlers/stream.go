package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"zero-music/config"
	"zero-music/services"

	"github.com/gin-gonic/gin"
)

// StreamHandler 音频流处理器
type StreamHandler struct {
	scanner  *services.MusicScanner
	musicDir string
}

// NewStreamHandler 创建新的音频流处理器
func NewStreamHandler(cfg *config.Config) *StreamHandler {
	scanner := services.NewMusicScanner(cfg.Music.Directory)
	return &StreamHandler{
		scanner:  scanner,
		musicDir: cfg.Music.Directory,
	}
}

// StreamAudio 流式传输音频文件
// @Summary 流式传输音频
// @Description 通过 HTTP 流式传输指定的音频文件
// @Tags stream
// @Produce audio/mpeg
// @Param id path string true "歌曲ID(文件名)"
// @Success 200 {file} binary "音频流"
// @Failure 404 {object} map[string]interface{} "文件未找到"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/stream/{id} [get]
func (h *StreamHandler) StreamAudio(c *gin.Context) {
	id := c.Param("id")

	// 验证文件是否存在于音乐列表中
	songs, err := h.scanner.Scan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to scan music files",
			"message": err.Error(),
		})
		return
	}

	// 查找歌曲
	var songPath string
	found := false
	for _, song := range songs {
		if song.ID == id {
			songPath = song.FilePath
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Song not found",
			"message": fmt.Sprintf("Song with ID '%s' does not exist", id),
		})
		return
	}

	// 打开音频文件
	file, err := os.Open(songPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to open audio file",
			"message": err.Error(),
		})
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get file info",
			"message": err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(songPath)))
	c.Header("Accept-Ranges", "bytes")

	// 流式传输文件
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		// 连接可能已断开,记录错误但不响应
		fmt.Printf("Error streaming audio: %v\n", err)
	}
}
