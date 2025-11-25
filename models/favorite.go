package models

import (
	"time"
)

// Favorite 收藏记录
type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	SongID    string    `json:"song_id"`
	CreatedAt time.Time `json:"created_at"`
}
