// Package repository 定义了数据访问层的接口抽象。
// 通过接口隔离数据存储实现，便于测试和替换底层存储。
package repository

import (
	"zero-music/models"
)

// UserRepository 定义了用户数据访问接口。
type UserRepository interface {
	// Create 创建新用户。
	Create(username, email, passwordHash string, role models.Role) (*models.User, error)

	// FindByID 根据 ID 查找用户。
	FindByID(id int64) (*models.User, error)

	// FindByUsername 根据用户名查找用户。
	FindByUsername(username string) (*models.User, error)

	// FindByEmail 根据邮箱查找用户。
	FindByEmail(email string) (*models.User, error)

	// Update 更新用户信息。
	Update(user *models.User) error

	// UpdatePassword 更新用户密码。
	UpdatePassword(userID int64, passwordHash string) error

	// Delete 删除用户。
	Delete(id int64) error

	// List 获取所有用户。
	List() ([]*models.User, error)

	// Exists 检查用户名或邮箱是否已存在。
	Exists(username, email string) (bool, error)
}

// FavoriteRepository 定义了收藏数据访问接口。
type FavoriteRepository interface {
	// Add 添加收藏。
	Add(userID int64, songID string) error

	// Remove 移除收藏。
	Remove(userID int64, songID string) error

	// IsFavorite 检查是否已收藏。
	IsFavorite(userID int64, songID string) (bool, error)

	// GetByUserID 获取用户收藏列表。
	GetByUserID(userID int64, limit, offset int) ([]*models.Favorite, error)

	// GetSongIDs 获取用户收藏的歌曲ID列表。
	GetSongIDs(userID int64) ([]string, error)

	// Count 获取用户收藏数量。
	Count(userID int64) (int, error)
}

// PlayStatsRepository 定义了播放统计数据访问接口。
type PlayStatsRepository interface {
	// RecordPlay 记录播放（包含历史和统计）。
	RecordPlay(userID int64, songID string, duration int) error

	// GetHistory 获取用户播放历史。
	GetHistory(userID int64, limit, offset int) ([]*models.PlayHistory, error)

	// GetStats 获取用户播放统计。
	GetStats(userID int64, limit, offset int) ([]*models.PlayStats, error)

	// GetMostPlayed 获取播放次数最多的歌曲（全局）。
	GetMostPlayed(limit int) ([]models.SongPlayCount, error)

	// GetRecentlyPlayed 获取最近播放的歌曲。
	GetRecentlyPlayed(userID int64, limit int) ([]string, error)

	// GetUserStats 获取用户统计摘要。
	GetUserStats(userID int64) (*models.UserStatsResult, error)
}

// PlaylistRepository 定义了播放列表数据访问接口。
type PlaylistRepository interface {
	// Create 创建播放列表。
	Create(userID int64, name, description string, isSmart bool, smartRules string) (*models.UserPlaylist, error)

	// FindByID 根据 ID 获取播放列表。
	FindByID(id int64) (*models.UserPlaylist, error)

	// GetByUserID 获取用户的所有播放列表。
	GetByUserID(userID int64) ([]*models.UserPlaylist, error)

	// Update 更新播放列表。
	Update(playlist *models.UserPlaylist) error

	// Delete 删除播放列表。
	Delete(id int64) error

	// AddSong 添加歌曲到播放列表。
	AddSong(playlistID int64, songID string) error

	// RemoveSong 从播放列表移除歌曲。
	RemoveSong(playlistID int64, songID string) error

	// GetSongs 获取播放列表中的歌曲。
	GetSongs(playlistID int64) ([]string, error)

	// ReorderSongs 重新排序播放列表歌曲。
	ReorderSongs(playlistID int64, songIDs []string) error

	// IsOwner 检查是否是播放列表所有者。
	IsOwner(playlistID, userID int64) (bool, error)
}
