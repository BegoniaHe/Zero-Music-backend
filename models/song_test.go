package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSong(t *testing.T) {
	// 创建一个临时音乐文件用于测试。
	tmpDir, err := os.MkdirTemp("", "models_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test_song.mp3")
	content := []byte("dummy audio content")
	err = os.WriteFile(filePath, content, 0644)
	assert.NoError(t, err)

	fileSize := int64(len(content))
	song := NewSong(filePath, fileSize)

	assert.NotNil(t, song)
	assert.Equal(t, "test_song", song.Title)
	assert.Equal(t, "Unknown", song.Artist)
	assert.Equal(t, "Unknown", song.Album)
	assert.Equal(t, filePath, song.FilePath)
	assert.Equal(t, "test_song.mp3", song.FileName)
	assert.Equal(t, fileSize, song.FileSize)
	assert.Equal(t, ".mp3", song.Format)
	assert.NotEmpty(t, song.ID)
	assert.Len(t, song.ID, 32)
}

func TestValidIDRegex(t *testing.T) {
	validID := "a1b2c3d4e5f678901234567890abcdef"
	assert.True(t, ValidIDRegex.MatchString(validID))

	invalidIDs := []string{
		"short",
		"toolong123456789012345678901234567890",
		"invalid-chars!",
		"",
	}

	for _, id := range invalidIDs {
		assert.False(t, ValidIDRegex.MatchString(id), "ID should be invalid: %s", id)
	}
}
