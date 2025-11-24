package models

import "regexp"

var (
	// ValidIDRegex 用于验证歌曲 ID 是否为有效的 SHA256 哈希（32 字节十六进制，即 64 个字符）
	ValidIDRegex = regexp.MustCompile(ValidIDPattern())
)
