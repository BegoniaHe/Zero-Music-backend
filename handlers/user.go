package handlers

import (
	"net/http"
	"strconv"

	"zero-music/logger"
	"zero-music/middleware"
	"zero-music/models"
	"zero-music/repository"
	"zero-music/services"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户相关处理器
type UserHandler struct {
	scanner      services.Scanner
	favoriteRepo repository.FavoriteRepository
	playStats    repository.PlayStatsRepository
	playlistRepo repository.PlaylistRepository
}

// NewUserHandler 创建用户处理器
func NewUserHandler(
	scanner services.Scanner,
	favoriteRepo repository.FavoriteRepository,
	playStats repository.PlayStatsRepository,
	playlistRepo repository.PlaylistRepository,
) *UserHandler {
	return &UserHandler{
		scanner:      scanner,
		favoriteRepo: favoriteRepo,
		playStats:    playStats,
		playlistRepo: playlistRepo,
	}
}

// RecordPlayRequest 记录播放请求
type RecordPlayRequest struct {
	SongID   string `json:"song_id" binding:"required"`
	Duration int    `json:"duration"` // 播放时长（秒）
}

// CreatePlaylistRequest 创建播放列表请求
type CreatePlaylistRequest struct {
	Name        string             `json:"name" binding:"required,min=1,max=100"`
	Description string             `json:"description"`
	IsSmart     bool               `json:"is_smart"`
	SmartRules  []models.SmartRule `json:"smart_rules"`
}

// UpdatePlaylistRequest 更新播放列表请求
type UpdatePlaylistRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	SmartRules  []models.SmartRule `json:"smart_rules"`
}

// AddSongRequest 添加歌曲请求
type AddSongRequest struct {
	SongID string `json:"song_id" binding:"required"`
}

// ReorderSongsRequest 重排序歌曲请求
type ReorderSongsRequest struct {
	SongIDs []string `json:"song_ids" binding:"required"`
}

// --- 辅助函数 ---

// getUserIDOrAbort 获取当前用户ID，如果未登录则返回错误响应
func getUserIDOrAbort(c *gin.Context) (int64, bool) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("未登录"))
		return 0, false
	}
	return userID, true
}

// checkPlaylistOwnership 检查播放列表所有权，返回是否通过检查
func (h *UserHandler) checkPlaylistOwnership(c *gin.Context, playlistID, userID int64) bool {
	isOwner, err := h.playlistRepo.IsOwner(playlistID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("播放列表"))
		return false
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, NewForbiddenError("无权访问此播放列表"))
		return false
	}
	return true
}

// --- 收藏相关 ---

// GetFavorites 获取收藏列表
func (h *UserHandler) GetFavorites(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	favorites, err := h.favoriteRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 获取歌曲详细信息
	var songs []*models.Song
	for _, fav := range favorites {
		if song := h.scanner.GetSongByID(fav.SongID); song != nil {
			songs = append(songs, song)
		}
	}

	count, err := h.favoriteRepo.Count(userID)
	if err != nil {
		logger.Warnf("获取收藏数量失败: %v", err)
		// 继续返回结果，但 total 可能不准确
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"songs": songs,
			"total": count,
		},
	})
}

// AddFavorite 添加收藏
func (h *UserHandler) AddFavorite(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	songID := c.Param("id")
	if songID == "" {
		c.JSON(http.StatusBadRequest, NewBadRequestError("歌曲ID不能为空"))
		return
	}

	// 验证歌曲存在
	if song := h.scanner.GetSongByID(songID); song == nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("歌曲"))
		return
	}

	if err := h.favoriteRepo.Add(userID, songID); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "添加收藏成功"})
}

// RemoveFavorite 移除收藏
func (h *UserHandler) RemoveFavorite(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	songID := c.Param("id")
	if err := h.favoriteRepo.Remove(userID, songID); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "移除收藏成功"})
}

// CheckFavorite 检查是否已收藏
func (h *UserHandler) CheckFavorite(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	songID := c.Param("id")
	isFav, err := h.favoriteRepo.IsFavorite(userID, songID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"is_favorite": isFav},
	})
}

// --- 播放历史相关 ---

// RecordPlay 记录播放
func (h *UserHandler) RecordPlay(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	var req RecordPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	// 验证歌曲存在
	if song := h.scanner.GetSongByID(req.SongID); song == nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("歌曲"))
		return
	}

	if err := h.playStats.RecordPlay(userID, req.SongID, req.Duration); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "记录播放成功"})
}

// GetPlayHistory 获取播放历史
func (h *UserHandler) GetPlayHistory(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	history, err := h.playStats.GetHistory(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 组合歌曲详细信息
	type HistoryItem struct {
		*models.PlayHistory
		Song *models.Song `json:"song,omitempty"`
	}

	var items []HistoryItem
	for _, hist := range history {
		item := HistoryItem{PlayHistory: hist}
		if song := h.scanner.GetSongByID(hist.SongID); song != nil {
			item.Song = song
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    items,
	})
}

// GetPlayStats 获取播放统计
func (h *UserHandler) GetPlayStats(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	stats, err := h.playStats.GetStats(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 组合歌曲详细信息
	type StatsItem struct {
		*models.PlayStats
		Song *models.Song `json:"song,omitempty"`
	}

	var items []StatsItem
	for _, stat := range stats {
		item := StatsItem{PlayStats: stat}
		if song := h.scanner.GetSongByID(stat.SongID); song != nil {
			item.Song = song
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    items,
	})
}

// GetUserStats 获取用户统计摘要
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	stats, err := h.playStats.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 添加收藏数量
	favCount, err := h.favoriteRepo.Count(userID)
	if err != nil {
		logger.Warnf("获取收藏数量失败: %v", err)
		// 继续返回结果，但 favorite_count 可能不准确
	}

	result := map[string]interface{}{
		"total_plays":     stats.TotalPlays,
		"total_play_time": stats.TotalPlayTime,
		"unique_songs":    stats.UniqueSongs,
		"favorite_count":  favCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// --- 播放列表相关 ---

// GetPlaylists 获取用户播放列表
func (h *UserHandler) GetPlaylists(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlists, err := h.playlistRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    playlists,
	})
}

// CreatePlaylist 创建播放列表
func (h *UserHandler) CreatePlaylist(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	// 序列化智能规则
	var rulesJSON string
	if len(req.SmartRules) > 0 {
		rulesBytes, err := models.MarshalSmartRules(req.SmartRules)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewBadRequestError("智能规则格式错误"))
			return
		}
		rulesJSON = string(rulesBytes)
	}

	playlist, err := h.playlistRepo.Create(userID, req.Name, req.Description, req.IsSmart, rulesJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    playlist,
	})
}

// GetPlaylist 获取播放列表详情
func (h *UserHandler) GetPlaylist(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的播放列表ID"))
		return
	}

	// 检查权限
	if !h.checkPlaylistOwnership(c, playlistID, userID) {
		return
	}

	playlist, err := h.playlistRepo.FindByID(playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 获取歌曲列表
	songIDs, _ := h.playlistRepo.GetSongs(playlistID)
	var songs []*models.Song
	for _, sid := range songIDs {
		if song := h.scanner.GetSongByID(sid); song != nil {
			songs = append(songs, song)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"playlist": playlist,
			"songs":    songs,
		},
	})
}

// UpdatePlaylist 更新播放列表
func (h *UserHandler) UpdatePlaylist(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的播放列表ID"))
		return
	}

	// 检查权限
	if !h.checkPlaylistOwnership(c, playlistID, userID) {
		return
	}

	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	playlist, err := h.playlistRepo.FindByID(playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	if req.Name != "" {
		playlist.Name = req.Name
	}
	playlist.Description = req.Description

	if err := h.playlistRepo.Update(playlist); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    playlist,
	})
}

// DeletePlaylist 删除播放列表
func (h *UserHandler) DeletePlaylist(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的播放列表ID"))
		return
	}

	// 检查权限
	if !h.checkPlaylistOwnership(c, playlistID, userID) {
		return
	}

	if err := h.playlistRepo.Delete(playlistID); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// AddSongToPlaylist 添加歌曲到播放列表
func (h *UserHandler) AddSongToPlaylist(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的播放列表ID"))
		return
	}

	// 检查权限
	if !h.checkPlaylistOwnership(c, playlistID, userID) {
		return
	}

	var req AddSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	// 验证歌曲存在
	if song := h.scanner.GetSongByID(req.SongID); song == nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("歌曲"))
		return
	}

	if err := h.playlistRepo.AddSong(playlistID, req.SongID); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "添加成功"})
}

// RemoveSongFromPlaylist 从播放列表移除歌曲
func (h *UserHandler) RemoveSongFromPlaylist(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的播放列表ID"))
		return
	}

	songID := c.Param("songId")

	// 检查权限
	if !h.checkPlaylistOwnership(c, playlistID, userID) {
		return
	}

	if err := h.playlistRepo.RemoveSong(playlistID, songID); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "移除成功"})
}

// ReorderPlaylistSongs 重新排序播放列表歌曲
func (h *UserHandler) ReorderPlaylistSongs(c *gin.Context) {
	userID, ok := getUserIDOrAbort(c)
	if !ok {
		return
	}

	playlistID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的播放列表ID"))
		return
	}

	// 检查权限
	if !h.checkPlaylistOwnership(c, playlistID, userID) {
		return
	}

	var req ReorderSongsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	if err := h.playlistRepo.ReorderSongs(playlistID, req.SongIDs); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "重排序成功"})
}
