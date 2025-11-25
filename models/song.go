package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

const (
	// SongIDBytes 是歌曲 ID 的字节长度（SHA256 哈希的前 16 字节）
	SongIDBytes = 16
	// SongIDHexLength 是歌曲 ID 的十六进制字符长度（16 字节 = 32 个十六进制字符）
	SongIDHexLength = SongIDBytes * 2
)

// Song 定义了歌曲的基本信息结构。
type Song struct {
	// ID 是歌曲的唯一标识符，通过文件路径的 SHA256 哈希生成。
	ID string `json:"id"`
	// Title 是歌曲的标题，通常从文件名中提取。
	Title string `json:"title"`
	// Artist 是歌曲的艺术家，默认为 "Unknown"。
	Artist string `json:"artist"`
	// Album 是歌曲所属的专辑，默认为 "Unknown"。
	Album string `json:"album"`
	// Duration 是歌曲的时长（以秒为单位），默认为 0。
	Duration int `json:"duration"`
	// DurationFormatted 是格式化后的时长字符串（如 "3:45"）。
	DurationFormatted string `json:"duration_formatted"`
	// FilePath 是歌曲文件的绝对路径。
	FilePath string `json:"file_path"`
	// FileName 是歌曲的文件名。
	FileName string `json:"file_name"`
	// FileSize 是歌曲文件的大小（以字节为单位）。
	FileSize int64 `json:"file_size"`
	// AddedAt 是歌曲文件最后修改的时间。
	AddedAt time.Time `json:"added_at"`
	// Format 是音频文件的格式/扩展名（如 .mp3, .flac）。
	Format string `json:"format"`
	// HasCover 标识该歌曲是否有嵌入的封面图片。
	HasCover bool `json:"has_cover"`
	// Year 是歌曲的发行年份。
	Year int `json:"year,omitempty"`
	// Track 是歌曲在专辑中的曲目编号。
	Track int `json:"track,omitempty"`
	// Genre 是歌曲的流派。
	Genre string `json:"genre,omitempty"`
}

// NewSong 根据给定的文件路径和文件大小创建一个新的 Song 实例。
func NewSong(filePath string, fileSize int64) *Song {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	// 默认使用移除了扩展名的文件名作为标题。
	title := strings.TrimSuffix(fileName, ext)

	// 使用文件的修改时间作为添加时间。
	addedAt := time.Now()
	if info, err := os.Stat(filePath); err == nil {
		addedAt = info.ModTime()
	}

	// 默认值
	artist := "Unknown"
	album := "Unknown"
	duration := 0

	song := &Song{
		ID:                generateID(filePath),
		Title:             title,
		Artist:            artist,
		Album:             album,
		Duration:          duration,
		DurationFormatted: "0:00",
		FilePath:          filePath,
		FileName:          fileName,
		FileSize:          fileSize,
		AddedAt:           addedAt,
		Format:            strings.ToLower(ext),
		HasCover:          false,
	}

	return song
}

// UpdateMetadata 尝试从文件中读取 ID3 标签等元数据并更新歌曲信息。
func (s *Song) UpdateMetadata() {
	file, err := os.Open(s.FilePath)
	if err != nil {
		return
	}
	defer file.Close()

	metadata, metaErr := tag.ReadFrom(file)
	if metaErr != nil {
		return
	}

	if metadata.Title() != "" {
		s.Title = metadata.Title()
	}
	if metadata.Artist() != "" {
		s.Artist = metadata.Artist()
	}
	if metadata.Album() != "" {
		s.Album = metadata.Album()
	}
	if metadata.Genre() != "" {
		s.Genre = metadata.Genre()
	}
	if metadata.Year() != 0 {
		s.Year = metadata.Year()
	}
	track, _ := metadata.Track()
	if track != 0 {
		s.Track = track
	}

	// 检查是否有封面
	s.HasCover = metadata.Picture() != nil

	// 解析时长
	s.parseDuration()
}

// parseDuration 解析音频文件的时长
func (s *Song) parseDuration() {
	// 对于 MP3 文件使用 mp3 库解析
	if s.Format == ".mp3" {
		s.Duration = s.parseMP3Duration()
	} else {
		// 对于其他格式，尝试根据文件大小和比特率估算
		s.Duration = s.estimateDuration()
	}

	s.DurationFormatted = FormatDuration(s.Duration)
}

// parseMP3Duration 解析 MP3 文件的时长
func (s *Song) parseMP3Duration() int {
	file, err := os.Open(s.FilePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	decoder := mp3.NewDecoder(file)
	var frame mp3.Frame
	var totalDuration time.Duration
	skipped := 0

	for {
		if err := decoder.Decode(&frame, &skipped); err != nil {
			break
		}
		totalDuration += frame.Duration()
	}

	return int(totalDuration.Seconds())
}

// estimateDuration 根据文件大小估算时长（用于非 MP3 格式）
func (s *Song) estimateDuration() int {
	// 假设平均比特率为 256 kbps
	// 时长（秒）= 文件大小（字节）/ (比特率 / 8)
	// 对于 256 kbps: 时长 = 文件大小 / 32000
	if s.FileSize == 0 {
		return 0
	}

	var bytesPerSecond int64
	switch s.Format {
	case ".flac":
		// FLAC 通常是 800-1200 kbps
		bytesPerSecond = 125000 // ~1000 kbps
	case ".wav":
		// WAV 16bit 44.1kHz 立体声
		bytesPerSecond = 176400
	case ".m4a", ".aac":
		// AAC 通常是 128-256 kbps
		bytesPerSecond = 24000 // ~192 kbps
	case ".ogg":
		// OGG 通常是 128-320 kbps
		bytesPerSecond = 25000 // ~200 kbps
	default:
		bytesPerSecond = 32000 // ~256 kbps
	}

	return int(s.FileSize / bytesPerSecond)
}

// FormatDuration 将秒数格式化为 "分:秒" 格式
func FormatDuration(seconds int) string {
	if seconds <= 0 {
		return "0:00"
	}

	minutes := seconds / 60
	secs := seconds % 60

	if minutes >= 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}

	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// generateID 使用文件路径的 SHA256 哈希值的前 16 字节生成一个唯一的歌曲 ID。
func generateID(filePath string) string {
	hash := sha256.Sum256([]byte(filePath))
	return hex.EncodeToString(hash[:SongIDBytes])
}

// ValidIDPattern 返回用于验证歌曲 ID 格式的正则表达式字符串
// ID 应为 32 个十六进制字符（16 字节的十六进制编码）
func ValidIDPattern() string {
	return fmt.Sprintf(`^[a-f0-9]{%d}$`, SongIDHexLength)
}
