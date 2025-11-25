package repository

import (
	"testing"
)

func TestSQLitePlaylistRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, err := repo.Create(user.ID, "My Playlist", "A test playlist", false, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if playlist.ID == 0 {
		t.Error("Expected playlist ID to be set")
	}
	if playlist.Name != "My Playlist" {
		t.Errorf("Expected name 'My Playlist', got '%s'", playlist.Name)
	}
	if playlist.Description != "A test playlist" {
		t.Errorf("Expected description 'A test playlist', got '%s'", playlist.Description)
	}
	if playlist.IsSmart {
		t.Error("Expected IsSmart to be false")
	}
}

func TestSQLitePlaylistRepository_Create_SmartPlaylist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	smartRules := `[{"field":"artist","operator":"equals","value":"Beatles"}]`
	playlist, err := repo.Create(user.ID, "Smart Playlist", "Auto-generated", true, smartRules)
	if err != nil {
		t.Fatalf("Create smart playlist failed: %v", err)
	}

	if !playlist.IsSmart {
		t.Error("Expected IsSmart to be true")
	}
	if playlist.SmartRules != smartRules {
		t.Errorf("Expected smart rules '%s', got '%s'", smartRules, playlist.SmartRules)
	}
}

func TestSQLitePlaylistRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	created, err := repo.Create(user.ID, "My Playlist", "Test", false, "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.Name != "My Playlist" {
		t.Errorf("Expected name 'My Playlist', got '%s'", found.Name)
	}
}

func TestSQLitePlaylistRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	repo.Create(user.ID, "Playlist 1", "", false, "")
	repo.Create(user.ID, "Playlist 2", "", false, "")
	repo.Create(user.ID, "Playlist 3", "", false, "")

	playlists, err := repo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	if len(playlists) != 3 {
		t.Errorf("Expected 3 playlists, got %d", len(playlists))
	}
}

func TestSQLitePlaylistRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "Original Name", "Original Desc", false, "")

	playlist.Name = "Updated Name"
	playlist.Description = "Updated Desc"
	playlist.CoverURL = "http://example.com/cover.jpg"

	err := repo.Update(playlist)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(playlist.ID)
	if found.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", found.Name)
	}
	if found.Description != "Updated Desc" {
		t.Errorf("Expected description 'Updated Desc', got '%s'", found.Description)
	}
	if found.CoverURL != "http://example.com/cover.jpg" {
		t.Errorf("Expected cover URL 'http://example.com/cover.jpg', got '%s'", found.CoverURL)
	}
}

func TestSQLitePlaylistRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "To Delete", "", false, "")

	err := repo.Delete(playlist.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.FindByID(playlist.ID)
	if err == nil {
		t.Error("Expected error when finding deleted playlist")
	}
}

func TestSQLitePlaylistRepository_AddSong(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "My Playlist", "", false, "")

	err := repo.AddSong(playlist.ID, "song1")
	if err != nil {
		t.Fatalf("AddSong failed: %v", err)
	}

	songs, err := repo.GetSongs(playlist.ID)
	if err != nil {
		t.Fatalf("GetSongs failed: %v", err)
	}

	if len(songs) != 1 {
		t.Errorf("Expected 1 song, got %d", len(songs))
	}
	if songs[0] != "song1" {
		t.Errorf("Expected song ID 'song1', got '%s'", songs[0])
	}
}

func TestSQLitePlaylistRepository_AddSong_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "My Playlist", "", false, "")

	repo.AddSong(playlist.ID, "song1")
	repo.AddSong(playlist.ID, "song1") // 重复添加

	songs, _ := repo.GetSongs(playlist.ID)
	if len(songs) != 1 {
		t.Errorf("Expected 1 song (duplicate ignored), got %d", len(songs))
	}
}

func TestSQLitePlaylistRepository_RemoveSong(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "My Playlist", "", false, "")

	repo.AddSong(playlist.ID, "song1")
	repo.AddSong(playlist.ID, "song2")

	err := repo.RemoveSong(playlist.ID, "song1")
	if err != nil {
		t.Fatalf("RemoveSong failed: %v", err)
	}

	songs, _ := repo.GetSongs(playlist.ID)
	if len(songs) != 1 {
		t.Errorf("Expected 1 song after removal, got %d", len(songs))
	}
	if songs[0] != "song2" {
		t.Errorf("Expected remaining song 'song2', got '%s'", songs[0])
	}
}

func TestSQLitePlaylistRepository_GetSongs_Order(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "My Playlist", "", false, "")

	repo.AddSong(playlist.ID, "song1")
	repo.AddSong(playlist.ID, "song2")
	repo.AddSong(playlist.ID, "song3")

	songs, err := repo.GetSongs(playlist.ID)
	if err != nil {
		t.Fatalf("GetSongs failed: %v", err)
	}

	// 应该按添加顺序（position）返回
	expectedOrder := []string{"song1", "song2", "song3"}
	for i, expected := range expectedOrder {
		if songs[i] != expected {
			t.Errorf("Expected song at position %d to be '%s', got '%s'", i, expected, songs[i])
		}
	}
}

func TestSQLitePlaylistRepository_ReorderSongs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "My Playlist", "", false, "")

	repo.AddSong(playlist.ID, "song1")
	repo.AddSong(playlist.ID, "song2")
	repo.AddSong(playlist.ID, "song3")

	// 重新排序: song3, song1, song2
	newOrder := []string{"song3", "song1", "song2"}
	err := repo.ReorderSongs(playlist.ID, newOrder)
	if err != nil {
		t.Fatalf("ReorderSongs failed: %v", err)
	}

	songs, _ := repo.GetSongs(playlist.ID)

	for i, expected := range newOrder {
		if songs[i] != expected {
			t.Errorf("Expected song at position %d to be '%s', got '%s'", i, expected, songs[i])
		}
	}
}

func TestSQLitePlaylistRepository_IsOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user1, _ := userRepo.Create("user1", "user1@example.com", "hash", "user")
	user2, _ := userRepo.Create("user2", "user2@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user1.ID, "User1's Playlist", "", false, "")

	// user1 应该是所有者
	isOwner, err := repo.IsOwner(playlist.ID, user1.ID)
	if err != nil {
		t.Fatalf("IsOwner failed: %v", err)
	}
	if !isOwner {
		t.Error("Expected user1 to be owner")
	}

	// user2 不应该是所有者
	isOwner, err = repo.IsOwner(playlist.ID, user2.ID)
	if err != nil {
		t.Fatalf("IsOwner failed: %v", err)
	}
	if isOwner {
		t.Error("Expected user2 not to be owner")
	}
}

func TestSQLitePlaylistRepository_SongCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLitePlaylistRepository(db)

	playlist, _ := repo.Create(user.ID, "My Playlist", "", false, "")

	repo.AddSong(playlist.ID, "song1")
	repo.AddSong(playlist.ID, "song2")
	repo.AddSong(playlist.ID, "song3")

	found, err := repo.FindByID(playlist.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.SongCount != 3 {
		t.Errorf("Expected song count 3, got %d", found.SongCount)
	}
}
