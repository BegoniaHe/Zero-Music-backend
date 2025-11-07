package main

import (
	"fmt"
	"log"
	"zero-music/services"
)

func main() {
	fmt.Println("=== 测试音乐文件扫描功能 ===\n")

	// 直接指定音乐目录(避免相对路径问题)
	musicDir := "./music"

	fmt.Printf("扫描目录: %s\n\n", musicDir)

	// 创建扫描器
	scanner := services.NewMusicScanner(musicDir)

	// 执行扫描
	songs, err := scanner.Scan()
	if err != nil {
		log.Fatalf("❌ 扫描失败: %v", err)
	}

	// 显示结果
	fmt.Printf("✅ 扫描完成!\n")
	fmt.Printf("📊 找到 %d 首歌曲\n\n", scanner.GetSongCount())

	if len(songs) > 0 {
		fmt.Println("歌曲列表:")
		for i, song := range songs {
			fmt.Printf("%d. %s\n", i+1, song.Title)
			fmt.Printf("   文件: %s\n", song.FileName)
			fmt.Printf("   大小: %.2f MB\n", float64(song.FileSize)/(1024*1024))
			fmt.Printf("   路径: %s\n\n", song.FilePath)
		}
	} else {
		fmt.Println("⚠️  未找到任何 mp3 文件")
		fmt.Printf("请将 mp3 文件放入目录: %s\n", musicDir)
	}
}
