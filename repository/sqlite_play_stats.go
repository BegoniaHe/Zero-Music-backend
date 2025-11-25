package repository

import (
	"zero-music/database"
	"zero-music/models"
)

// SQLitePlayStatsRepository 是 PlayStatsRepository 的 SQLite 实现。
type SQLitePlayStatsRepository struct {
	db database.DB
}

// NewSQLitePlayStatsRepository 创建 SQLite 播放统计仓储实例。
func NewSQLitePlayStatsRepository(db database.DB) *SQLitePlayStatsRepository {
	return &SQLitePlayStatsRepository{db: db}
}

// RecordPlay 记录播放（包含历史和统计）。
func (r *SQLitePlayStatsRepository) RecordPlay(userID int64, songID string, duration int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 插入播放历史
	_, err = tx.Exec(
		`INSERT INTO play_history (user_id, song_id, play_duration) VALUES (?, ?, ?)`,
		userID, songID, duration,
	)
	if err != nil {
		return err
	}

	// 更新或插入播放统计
	_, err = tx.Exec(`
		INSERT INTO play_stats (user_id, song_id, play_count, total_play_time, last_played_at)
		VALUES (?, ?, 1, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, song_id) DO UPDATE SET
			play_count = play_count + 1,
			total_play_time = total_play_time + excluded.total_play_time,
			last_played_at = CURRENT_TIMESTAMP
	`, userID, songID, duration)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetHistory 获取用户播放历史。
func (r *SQLitePlayStatsRepository) GetHistory(userID int64, limit, offset int) ([]*models.PlayHistory, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, song_id, played_at, play_duration 
		FROM play_history 
		WHERE user_id = ? 
		ORDER BY played_at DESC 
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.PlayHistory
	for rows.Next() {
		h := &models.PlayHistory{}
		if err := rows.Scan(&h.ID, &h.UserID, &h.SongID, &h.PlayedAt, &h.PlayDuration); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

// GetStats 获取用户播放统计。
func (r *SQLitePlayStatsRepository) GetStats(userID int64, limit, offset int) ([]*models.PlayStats, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, song_id, play_count, total_play_time, last_played_at
		FROM play_stats
		WHERE user_id = ?
		ORDER BY play_count DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PlayStats
	for rows.Next() {
		s := &models.PlayStats{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.SongID, &s.PlayCount, &s.TotalPlayTime, &s.LastPlayedAt); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// GetMostPlayed 获取播放次数最多的歌曲（全局）。
func (r *SQLitePlayStatsRepository) GetMostPlayed(limit int) ([]models.SongPlayCount, error) {
	rows, err := r.db.Query(`
		SELECT song_id, SUM(play_count) as total_plays
		FROM play_stats
		GROUP BY song_id
		ORDER BY total_plays DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.SongPlayCount
	for rows.Next() {
		var item models.SongPlayCount
		if err := rows.Scan(&item.SongID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// GetRecentlyPlayed 获取最近播放的歌曲。
func (r *SQLitePlayStatsRepository) GetRecentlyPlayed(userID int64, limit int) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT song_id
		FROM play_history
		WHERE user_id = ?
		ORDER BY played_at DESC
		LIMIT ?
	`, userID, limit)
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

// GetUserStats 获取用户统计摘要。
func (r *SQLitePlayStatsRepository) GetUserStats(userID int64) (*models.UserStatsResult, error) {
	stats := &models.UserStatsResult{}

	// 总播放次数
	err := r.db.QueryRow(`SELECT COALESCE(SUM(play_count), 0) FROM play_stats WHERE user_id = ?`, userID).Scan(&stats.TotalPlays)
	if err != nil {
		return nil, err
	}

	// 总播放时长
	err = r.db.QueryRow(`SELECT COALESCE(SUM(total_play_time), 0) FROM play_stats WHERE user_id = ?`, userID).Scan(&stats.TotalPlayTime)
	if err != nil {
		return nil, err
	}

	// 独立歌曲数
	err = r.db.QueryRow(`SELECT COUNT(DISTINCT song_id) FROM play_stats WHERE user_id = ?`, userID).Scan(&stats.UniqueSongs)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
