package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"zero-music/logger"
	"zero-music/models"
)

// MusicScanner 负责扫描音乐目录并管理歌曲列表缓存。
// 它实现了 Scanner 接口。
type MusicScanner struct {
	directory        string
	supportedFormats []string
	songs            []*models.Song
	songIndex        map[string]*models.Song // ID -> Song 的索引，用于快速查找
	mu               sync.RWMutex
	lastScan         time.Time
	cacheTTL         time.Duration
	lastDirModTime   time.Time
}

// NewMusicScanner 创建并返回一个新的 MusicScanner 实例。
func NewMusicScanner(directory string, supportedFormats []string, cacheTTLMinutes int) *MusicScanner {
	if len(supportedFormats) == 0 {
		supportedFormats = []string{".mp3"}
	}
	if cacheTTLMinutes <= 0 {
		cacheTTLMinutes = 5
	}
	return &MusicScanner{
		directory:        directory,
		supportedFormats: supportedFormats,
		songs:            make([]*models.Song, 0),
		songIndex:        make(map[string]*models.Song),
		cacheTTL:         time.Duration(cacheTTLMinutes) * time.Minute,
	}
}

// Scan 扫描音乐目录并返回歌曲列表（带缓存）。
func (s *MusicScanner) Scan(ctx context.Context) ([]*models.Song, error) {
	// 先在锁外获取目录信息，减少持锁时间
	dirInfo, err := os.Stat(s.directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("音乐目录不存在: %s", s.directory)
		}
		return nil, fmt.Errorf("音乐目录不可访问: %w", err)
	}

	s.mu.RLock()
	if s.canServeFromCacheWithDirInfo(dirInfo) {
		songs := cloneSongs(s.songs)
		s.mu.RUnlock()
		return songs, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if s.canServeFromCacheWithDirInfo(dirInfo) {
		return cloneSongs(s.songs), nil
	}

	return s.scanInternal(ctx, dirInfo)
}

// canServeFromCacheWithDirInfo 检查是否可以从缓存返回（使用预先获取的目录信息）
func (s *MusicScanner) canServeFromCacheWithDirInfo(dirInfo os.FileInfo) bool {
	if len(s.songs) == 0 {
		return false
	}
	if time.Since(s.lastScan) >= s.cacheTTL {
		return false
	}
	if dirInfo.ModTime().After(s.lastDirModTime) {
		return false
	}
	return true
}

// scanInternal 是实际的扫描逻辑。调用此函数前必须获取写锁。
func (s *MusicScanner) scanInternal(ctx context.Context, dirInfo os.FileInfo) ([]*models.Song, error) {

	newSongs := make([]*models.Song, 0)
	newIndex := make(map[string]*models.Song)

	err := filepath.WalkDir(s.directory, func(path string, d os.DirEntry, walkErr error) error {
		// 检查 context 是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if walkErr != nil {
			// 记录具体的路径错误
			return fmt.Errorf("访问路径 %s 失败: %w", path, walkErr)
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, supported := range s.supportedFormats {
			if ext == strings.ToLower(supported) {
				info, err := d.Info()
				if err != nil {
					// 记录获取文件信息失败，但不中断扫描
					logger.Warnf("获取文件信息失败 %s: %v", path, err)
					return nil
				}
				song := models.NewSong(path, info.Size())
				song.UpdateMetadata()
				newSongs = append(newSongs, song)
				newIndex[song.ID] = song
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描目录时出错: %v", err)
	}

	s.songs = newSongs
	s.songIndex = newIndex
	s.lastScan = time.Now()
	s.lastDirModTime = dirInfo.ModTime()

	return cloneSongs(newSongs), nil
}

// Refresh 强制执行一次新的扫描,并刷新歌曲列表缓存。
func (s *MusicScanner) Refresh(ctx context.Context) error {
	// 在锁外获取目录信息
	dirInfo, err := os.Stat(s.directory)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("音乐目录不存在: %s", s.directory)
		}
		return fmt.Errorf("音乐目录不可访问: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.scanInternal(ctx, dirInfo)
	return err
}

// GetSongs 返回当前缓存的歌曲列表的深度拷贝。
func (s *MusicScanner) GetSongs() []*models.Song {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneSongs(s.songs)
}

// GetSongCount 返回当前缓存的歌曲数量。
func (s *MusicScanner) GetSongCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.songs)
}

// GetSongByID 根据 ID 查找并返回指定的歌曲。
// 如果未找到歌曲，则返回 nil。
// 此方法使用索引进行高效查找。
func (s *MusicScanner) GetSongByID(id string) *models.Song {
	s.mu.RLock()
	defer s.mu.RUnlock()
	song, ok := s.songIndex[id]
	if !ok || song == nil {
		return nil
	}
	copiedSong := *song
	return &copiedSong
}

func cloneSongs(src []*models.Song) []*models.Song {
	if len(src) == 0 {
		return []*models.Song{}
	}
	dst := make([]*models.Song, len(src))
	for i, song := range src {
		if song == nil {
			continue
		}
		copied := *song
		dst[i] = &copied
	}
	return dst
}
