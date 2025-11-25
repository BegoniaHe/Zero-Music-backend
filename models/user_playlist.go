package models

import (
	"encoding/json"
	"time"
)

// UserPlaylist 用户播放列表
type UserPlaylist struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	IsSmart     bool      `json:"is_smart"`
	SmartRules  string    `json:"smart_rules"` // JSON 格式的智能规则
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	SongCount   int       `json:"song_count,omitempty"` // 非数据库字段
}

// SmartRule 智能播放列表规则
type SmartRule struct {
	Field    string `json:"field"`    // artist, album, genre, year, play_count
	Operator string `json:"operator"` // equals, contains, greater_than, less_than
	Value    string `json:"value"`
}

// MarshalSmartRules 序列化智能规则为 JSON 字符串
func MarshalSmartRules(rules []SmartRule) ([]byte, error) {
	return json.Marshal(rules)
}

// PlaylistSong 播放列表歌曲关联
type PlaylistSong struct {
	ID         int64     `json:"id"`
	PlaylistID int64     `json:"playlist_id"`
	SongID     string    `json:"song_id"`
	Position   int       `json:"position"`
	AddedAt    time.Time `json:"added_at"`
}
