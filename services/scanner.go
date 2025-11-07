package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"zero-music/models"
)

// MusicScanner 音乐文件扫描器
type MusicScanner struct {
	directory string
	songs     []*models.Song
}

// NewMusicScanner 创建新的音乐扫描器
func NewMusicScanner(directory string) *MusicScanner {
	return &MusicScanner{
		directory: directory,
		songs:     make([]*models.Song, 0),
	}
}

// Scan 扫描指定目录中的所有音乐文件
func (s *MusicScanner) Scan() ([]*models.Song, error) {
	s.songs = make([]*models.Song, 0)

	// 检查目录是否存在
	if _, err := os.Stat(s.directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", s.directory)
	}

	// 遍历目录
	err := filepath.Walk(s.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查是否是 mp3 文件
		if strings.ToLower(filepath.Ext(path)) == ".mp3" {
			song := models.NewSong(path, info.Size())
			s.songs = append(s.songs, song)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描目录失败: %v", err)
	}

	return s.songs, nil
}

// GetSongs 获取已扫描的歌曲列表
func (s *MusicScanner) GetSongs() []*models.Song {
	return s.songs
}

// GetSongCount 获取歌曲数量
func (s *MusicScanner) GetSongCount() int {
	return len(s.songs)
}
