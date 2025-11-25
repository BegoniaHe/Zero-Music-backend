package models

import (
	"time"
)

// PlayHistory 播放历史记录
type PlayHistory struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	SongID       string    `json:"song_id"`
	PlayedAt     time.Time `json:"played_at"`
	PlayDuration int       `json:"play_duration"` // 播放时长（秒）
}

// PlayStats 播放统计
type PlayStats struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	SongID        string     `json:"song_id"`
	PlayCount     int        `json:"play_count"`
	TotalPlayTime int        `json:"total_play_time"` // 总播放时长（秒）
	LastPlayedAt  *time.Time `json:"last_played_at"`
}

// SongPlayCount 歌曲播放次数统计
type SongPlayCount struct {
	SongID    string `json:"song_id"`
	PlayCount int    `json:"play_count"`
}

// UserStatsResult 用户统计摘要
type UserStatsResult struct {
	TotalPlays    int `json:"total_plays"`
	TotalPlayTime int `json:"total_play_time"`
	UniqueSongs   int `json:"unique_songs"`
}
