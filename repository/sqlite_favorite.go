package repository

import (
	"zero-music/database"
	"zero-music/models"
)

// SQLiteFavoriteRepository 是 FavoriteRepository 的 SQLite 实现。
type SQLiteFavoriteRepository struct {
	db database.DB
}

// NewSQLiteFavoriteRepository 创建 SQLite 收藏仓储实例。
func NewSQLiteFavoriteRepository(db database.DB) *SQLiteFavoriteRepository {
	return &SQLiteFavoriteRepository{db: db}
}

// Add 添加收藏。
func (r *SQLiteFavoriteRepository) Add(userID int64, songID string) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO favorites (user_id, song_id) VALUES (?, ?)`,
		userID, songID,
	)
	return err
}

// Remove 移除收藏。
func (r *SQLiteFavoriteRepository) Remove(userID int64, songID string) error {
	_, err := r.db.Exec(
		`DELETE FROM favorites WHERE user_id = ? AND song_id = ?`,
		userID, songID,
	)
	return err
}

// IsFavorite 检查是否已收藏。
func (r *SQLiteFavoriteRepository) IsFavorite(userID int64, songID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM favorites WHERE user_id = ? AND song_id = ?`,
		userID, songID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByUserID 获取用户收藏列表。
func (r *SQLiteFavoriteRepository) GetByUserID(userID int64, limit, offset int) ([]*models.Favorite, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, song_id, created_at
		FROM favorites
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*models.Favorite
	for rows.Next() {
		f := &models.Favorite{}
		if err := rows.Scan(&f.ID, &f.UserID, &f.SongID, &f.CreatedAt); err != nil {
			return nil, err
		}
		favorites = append(favorites, f)
	}
	return favorites, rows.Err()
}

// GetSongIDs 获取用户收藏的歌曲ID列表。
func (r *SQLiteFavoriteRepository) GetSongIDs(userID int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT song_id FROM favorites WHERE user_id = ? ORDER BY created_at DESC
	`, userID)
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

// Count 获取用户收藏数量。
func (r *SQLiteFavoriteRepository) Count(userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM favorites WHERE user_id = ?`,
		userID,
	).Scan(&count)
	return count, err
}
