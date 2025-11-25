package models

import "regexp"

var (
	// ValidIDRegex 用于验证歌曲 ID 格式。
	// ID 应为 32 个十六进制字符（16 字节的十六进制编码）。
	ValidIDRegex = regexp.MustCompile(ValidIDPattern())
)
