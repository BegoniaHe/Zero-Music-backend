// Package utils 提供项目共享的工具函数。
package utils

import (
	"mime"
	"path/filepath"
	"strings"
)

// GetAudioMimeType 根据文件扩展名返回对应的音频 MIME 类型。
func GetAudioMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// 为常见音频格式提供备选 MIME 类型
		switch ext {
		case ".mp3":
			return "audio/mpeg"
		case ".flac":
			return "audio/flac"
		case ".wav":
			return "audio/wav"
		case ".m4a":
			return "audio/mp4"
		case ".ogg":
			return "audio/ogg"
		default:
			return "application/octet-stream"
		}
	}
	return mimeType
}
