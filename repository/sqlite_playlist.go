package repository

import (
	"time"

	"zero-music/database"
	"zero-music/logger"
	"zero-music/models"
)

// SQLitePlaylistRepository 是 PlaylistRepository 的 SQLite 实现。
type SQLitePlaylistRepository struct {
	db database.DB
}

// NewSQLitePlaylistRepository 创建 SQLite 播放列表仓储实例。
func NewSQLitePlaylistRepository(db database.DB) *SQLitePlaylistRepository {
	return &SQLitePlaylistRepository{db: db}
}

// Create 创建播放列表。
func (r *SQLitePlaylistRepository) Create(userID int64, name, description string, isSmart bool, smartRules string) (*models.UserPlaylist, error) {
	result, err := r.db.Exec(`
		INSERT INTO playlists (user_id, name, description, is_smart, smart_rules)
		VALUES (?, ?, ?, ?, ?)
	`, userID, name, description, isSmart, smartRules)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.UserPlaylist{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		IsSmart:     isSmart,
		SmartRules:  smartRules,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// FindByID 根据 ID 获取播放列表。
func (r *SQLitePlaylistRepository) FindByID(id int64) (*models.UserPlaylist, error) {
	playlist := &models.UserPlaylist{}
	err := r.db.QueryRow(`
		SELECT id, user_id, name, description, cover_url, is_smart, smart_rules, created_at, updated_at
		FROM playlists WHERE id = ?
	`, id).Scan(&playlist.ID, &playlist.UserID, &playlist.Name, &playlist.Description,
		&playlist.CoverURL, &playlist.IsSmart, &playlist.SmartRules, &playlist.CreatedAt, &playlist.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 获取歌曲数量
	r.db.QueryRow(`SELECT COUNT(*) FROM playlist_songs WHERE playlist_id = ?`, id).Scan(&playlist.SongCount)

	return playlist, nil
}

// GetByUserID 获取用户的所有播放列表。
func (r *SQLitePlaylistRepository) GetByUserID(userID int64) ([]*models.UserPlaylist, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.user_id, p.name, p.description, p.cover_url, p.is_smart, p.smart_rules, 
		       p.created_at, p.updated_at,
		       (SELECT COUNT(*) FROM playlist_songs ps WHERE ps.playlist_id = p.id) as song_count
		FROM playlists p
		WHERE p.user_id = ?
		ORDER BY p.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*models.UserPlaylist
	for rows.Next() {
		p := &models.UserPlaylist{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CoverURL,
			&p.IsSmart, &p.SmartRules, &p.CreatedAt, &p.UpdatedAt, &p.SongCount); err != nil {
			return nil, err
		}
		playlists = append(playlists, p)
	}
	return playlists, rows.Err()
}

// Update 更新播放列表。
func (r *SQLitePlaylistRepository) Update(playlist *models.UserPlaylist) error {
	_, err := r.db.Exec(`
		UPDATE playlists 
		SET name = ?, description = ?, cover_url = ?, smart_rules = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, playlist.Name, playlist.Description, playlist.CoverURL, playlist.SmartRules, playlist.ID)
	return err
}

// Delete 删除播放列表。
func (r *SQLitePlaylistRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM playlists WHERE id = ?`, id)
	return err
}

// AddSong 添加歌曲到播放列表。
func (r *SQLitePlaylistRepository) AddSong(playlistID int64, songID string) error {
	// 获取当前最大位置
	var maxPos int
	if err := r.db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM playlist_songs WHERE playlist_id = ?`, playlistID).Scan(&maxPos); err != nil {
		logger.Warnf("获取播放列表 %d 的最大位置失败: %v", playlistID, err)
	}

	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO playlist_songs (playlist_id, song_id, position)
		VALUES (?, ?, ?)
	`, playlistID, songID, maxPos+1)

	// 更新播放列表的更新时间
	if _, updateErr := r.db.Exec(`UPDATE playlists SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, playlistID); updateErr != nil {
		logger.Warnf("更新播放列表 %d 的时间戳失败: %v", playlistID, updateErr)
	}

	return err
}

// RemoveSong 从播放列表移除歌曲。
func (r *SQLitePlaylistRepository) RemoveSong(playlistID int64, songID string) error {
	_, err := r.db.Exec(`
		DELETE FROM playlist_songs WHERE playlist_id = ? AND song_id = ?
	`, playlistID, songID)

	// 更新播放列表的更新时间
	if _, updateErr := r.db.Exec(`UPDATE playlists SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, playlistID); updateErr != nil {
		logger.Warnf("更新播放列表 %d 的时间戳失败: %v", playlistID, updateErr)
	}

	return err
}

// GetSongs 获取播放列表中的歌曲。
func (r *SQLitePlaylistRepository) GetSongs(playlistID int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT song_id FROM playlist_songs 
		WHERE playlist_id = ? 
		ORDER BY position
	`, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songIDs []string
	for rows.Next() {
		var songID string
		if err := rows.Scan(&songID); err != nil {
			return nil, err
		}
		songIDs = append(songIDs, songID)
	}
	return songIDs, rows.Err()
}

// ReorderSongs 重新排序播放列表歌曲。
func (r *SQLitePlaylistRepository) ReorderSongs(playlistID int64, songIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, songID := range songIDs {
		_, err := tx.Exec(`
			UPDATE playlist_songs SET position = ? WHERE playlist_id = ? AND song_id = ?
		`, i+1, playlistID, songID)
		if err != nil {
			return err
		}
	}

	// 更新播放列表的更新时间
	if _, updateErr := tx.Exec(`UPDATE playlists SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, playlistID); updateErr != nil {
		logger.Warnf("更新播放列表 %d 的时间戳失败: %v", playlistID, updateErr)
	}

	return tx.Commit()
}

// IsOwner 检查是否是播放列表所有者。
func (r *SQLitePlaylistRepository) IsOwner(playlistID, userID int64) (bool, error) {
	var ownerID int64
	err := r.db.QueryRow(`SELECT user_id FROM playlists WHERE id = ?`, playlistID).Scan(&ownerID)
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}
