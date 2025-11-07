package models

import (
	"path/filepath"
	"strings"
	"time"
)

// Song 歌曲信息结构
type Song struct {
	ID       string    `json:"id"`        // 唯一标识符
	Title    string    `json:"title"`     // 歌曲标题
	Artist   string    `json:"artist"`    // 艺术家
	Album    string    `json:"album"`     // 专辑
	Duration int       `json:"duration"`  // 时长(秒)
	FilePath string    `json:"file_path"` // 文件路径
	FileName string    `json:"file_name"` // 文件名
	FileSize int64     `json:"file_size"` // 文件大小(字节)
	AddedAt  time.Time `json:"added_at"`  // 添加时间
}

// NewSong 从文件路径创建歌曲对象
func NewSong(filePath string, fileSize int64) *Song {
	fileName := filepath.Base(filePath)
	// 移除扩展名作为标题
	title := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return &Song{
		ID:       generateID(filePath),
		Title:    title,
		Artist:   "Unknown", // 后续可以从 ID3 标签读取
		Album:    "Unknown",
		Duration: 0, // 后续可以从文件读取
		FilePath: filePath,
		FileName: fileName,
		FileSize: fileSize,
		AddedAt:  time.Now(),
	}
}

// generateID 生成歌曲唯一标识符
func generateID(filePath string) string {
	// 使用文件路径的哈希作为 ID
	// 这里简化处理,实际可以使用更复杂的哈希算法
	return filepath.Base(filePath)
}
