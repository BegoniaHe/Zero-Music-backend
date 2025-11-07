package handlers

import (
	"net/http"
	"zero-music/config"
	"zero-music/services"

	"github.com/gin-gonic/gin"
)

// PlaylistHandler 播放列表处理器
type PlaylistHandler struct {
	scanner *services.MusicScanner
}

// NewPlaylistHandler 创建新的播放列表处理器
func NewPlaylistHandler(cfg *config.Config) *PlaylistHandler {
	scanner := services.NewMusicScanner(cfg.Music.Directory)
	return &PlaylistHandler{
		scanner: scanner,
	}
}

// GetAllSongs 获取所有歌曲列表
// @Summary 获取所有歌曲
// @Description 返回音乐目录中所有可用的歌曲列表
// @Tags playlist
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回歌曲列表"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/songs [get]
func (h *PlaylistHandler) GetAllSongs(c *gin.Context) {
	// 扫描音乐文件
	songs, err := h.scanner.Scan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to scan music files",
			"message": err.Error(),
		})
		return
	}

	// 返回歌曲列表
	c.JSON(http.StatusOK, gin.H{
		"total": len(songs),
		"songs": songs,
	})
}

// GetSongByID 根据ID获取歌曲信息
// @Summary 获取指定歌曲信息
// @Description 根据歌曲ID返回歌曲详细信息
// @Tags playlist
// @Produce json
// @Param id path string true "歌曲ID"
// @Success 200 {object} models.Song "成功返回歌曲信息"
// @Failure 404 {object} map[string]interface{} "歌曲未找到"
// @Router /api/song/{id} [get]
func (h *PlaylistHandler) GetSongByID(c *gin.Context) {
	id := c.Param("id")

	// 扫描获取所有歌曲
	songs, err := h.scanner.Scan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to scan music files",
			"message": err.Error(),
		})
		return
	}

	// 查找指定ID的歌曲
	for _, song := range songs {
		if song.ID == id {
			c.JSON(http.StatusOK, song)
			return
		}
	}

	// 未找到歌曲
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "Song not found",
		"message": "The requested song does not exist",
	})
}
