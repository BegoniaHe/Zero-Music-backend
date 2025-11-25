package repository

import (
	"testing"
)

func TestSQLitePlayStatsRepository_RecordPlay(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	err := repo.RecordPlay(user.ID, "song1", 180)
	if err != nil {
		t.Fatalf("RecordPlay failed: %v", err)
	}

	// 验证历史记录
	history, err := repo.GetHistory(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 history record, got %d", len(history))
	}
	if history[0].SongID != "song1" {
		t.Errorf("Expected song ID 'song1', got '%s'", history[0].SongID)
	}
	if history[0].PlayDuration != 180 {
		t.Errorf("Expected play duration 180, got %d", history[0].PlayDuration)
	}

	// 验证统计记录
	stats, err := repo.GetStats(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if len(stats) != 1 {
		t.Errorf("Expected 1 stats record, got %d", len(stats))
	}
	if stats[0].PlayCount != 1 {
		t.Errorf("Expected play count 1, got %d", stats[0].PlayCount)
	}
}

func TestSQLitePlayStatsRepository_RecordPlay_MultiplePlays(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	// 播放同一首歌多次
	repo.RecordPlay(user.ID, "song1", 100)
	repo.RecordPlay(user.ID, "song1", 150)
	repo.RecordPlay(user.ID, "song1", 200)

	stats, err := repo.GetStats(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if len(stats) != 1 {
		t.Errorf("Expected 1 stats record, got %d", len(stats))
	}
	if stats[0].PlayCount != 3 {
		t.Errorf("Expected play count 3, got %d", stats[0].PlayCount)
	}
	if stats[0].TotalPlayTime != 450 {
		t.Errorf("Expected total play time 450, got %d", stats[0].TotalPlayTime)
	}
}

func TestSQLitePlayStatsRepository_GetHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	repo.RecordPlay(user.ID, "song1", 100)
	repo.RecordPlay(user.ID, "song2", 200)
	repo.RecordPlay(user.ID, "song3", 300)

	history, err := repo.GetHistory(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 history records, got %d", len(history))
	}

	// 测试分页
	history, err = repo.GetHistory(user.ID, 2, 0)
	if err != nil {
		t.Fatalf("GetHistory with limit failed: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("Expected 2 history records with limit, got %d", len(history))
	}

	// 测试偏移
	history, err = repo.GetHistory(user.ID, 10, 2)
	if err != nil {
		t.Fatalf("GetHistory with offset failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 history record with offset, got %d", len(history))
	}
}

func TestSQLitePlayStatsRepository_GetStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	// 播放不同歌曲不同次数
	repo.RecordPlay(user.ID, "song1", 100) // 1 次
	repo.RecordPlay(user.ID, "song2", 100)
	repo.RecordPlay(user.ID, "song2", 100) // 2 次
	repo.RecordPlay(user.ID, "song3", 100)
	repo.RecordPlay(user.ID, "song3", 100)
	repo.RecordPlay(user.ID, "song3", 100) // 3 次

	stats, err := repo.GetStats(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if len(stats) != 3 {
		t.Errorf("Expected 3 stats records, got %d", len(stats))
	}

	// 应该按播放次数降序排列
	if stats[0].PlayCount != 3 {
		t.Errorf("Expected first stat to have play count 3, got %d", stats[0].PlayCount)
	}
}

func TestSQLitePlayStatsRepository_GetMostPlayed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user1, _ := userRepo.Create("user1", "user1@example.com", "hash", "user")
	user2, _ := userRepo.Create("user2", "user2@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	// user1 播放 song1 两次，song2 一次
	repo.RecordPlay(user1.ID, "song1", 100)
	repo.RecordPlay(user1.ID, "song1", 100)
	repo.RecordPlay(user1.ID, "song2", 100)

	// user2 播放 song1 三次
	repo.RecordPlay(user2.ID, "song1", 100)
	repo.RecordPlay(user2.ID, "song1", 100)
	repo.RecordPlay(user2.ID, "song1", 100)

	mostPlayed, err := repo.GetMostPlayed(10)
	if err != nil {
		t.Fatalf("GetMostPlayed failed: %v", err)
	}

	if len(mostPlayed) != 2 {
		t.Errorf("Expected 2 songs in most played, got %d", len(mostPlayed))
	}

	// song1 总共播放 5 次，应该排第一
	if mostPlayed[0].SongID != "song1" {
		t.Errorf("Expected 'song1' to be most played, got '%s'", mostPlayed[0].SongID)
	}
	if mostPlayed[0].PlayCount != 5 {
		t.Errorf("Expected play count 5, got %d", mostPlayed[0].PlayCount)
	}
}

func TestSQLitePlayStatsRepository_GetRecentlyPlayed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	repo.RecordPlay(user.ID, "song1", 100)
	repo.RecordPlay(user.ID, "song2", 100)
	repo.RecordPlay(user.ID, "song3", 100)
	repo.RecordPlay(user.ID, "song1", 100) // song1 再次播放

	recently, err := repo.GetRecentlyPlayed(user.ID, 3)
	if err != nil {
		t.Fatalf("GetRecentlyPlayed failed: %v", err)
	}

	if len(recently) != 3 {
		t.Errorf("Expected 3 recently played songs, got %d", len(recently))
	}

	// 最近播放的是 song1
	if recently[0] != "song1" {
		t.Errorf("Expected 'song1' to be most recently played, got '%s'", recently[0])
	}
}

func TestSQLitePlayStatsRepository_GetUserStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	repo.RecordPlay(user.ID, "song1", 100)
	repo.RecordPlay(user.ID, "song1", 150)
	repo.RecordPlay(user.ID, "song2", 200)

	stats, err := repo.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats failed: %v", err)
	}

	if stats.TotalPlays != 3 {
		t.Errorf("Expected total plays 3, got %d", stats.TotalPlays)
	}
	if stats.TotalPlayTime != 450 {
		t.Errorf("Expected total play time 450, got %d", stats.TotalPlayTime)
	}
	if stats.UniqueSongs != 2 {
		t.Errorf("Expected unique songs 2, got %d", stats.UniqueSongs)
	}
}

func TestSQLitePlayStatsRepository_GetUserStats_NoData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlayStatsRepository(db)

	stats, err := repo.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("GetUserStats failed: %v", err)
	}

	if stats.TotalPlays != 0 {
		t.Errorf("Expected total plays 0, got %d", stats.TotalPlays)
	}
	if stats.TotalPlayTime != 0 {
		t.Errorf("Expected total play time 0, got %d", stats.TotalPlayTime)
	}
	if stats.UniqueSongs != 0 {
		t.Errorf("Expected unique songs 0, got %d", stats.UniqueSongs)
	}
}
