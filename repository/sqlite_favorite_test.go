package repository

import (
	"testing"
)

func TestSQLiteFavoriteRepository_Add(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 先创建一个用户
	userRepo := NewSQLiteUserRepository(db)
	user, err := userRepo.Create("testuser", "test@example.com", "hash", "user")
	if err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	repo := NewSQLiteFavoriteRepository(db)

	err = repo.Add(user.ID, "song1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 验证添加成功
	isFav, err := repo.IsFavorite(user.ID, "song1")
	if err != nil {
		t.Fatalf("IsFavorite failed: %v", err)
	}
	if !isFav {
		t.Error("Expected song to be favorite")
	}
}

func TestSQLiteFavoriteRepository_Add_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLiteFavoriteRepository(db)

	// 添加两次同一首歌，应该不报错（INSERT OR IGNORE）
	err := repo.Add(user.ID, "song1")
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	err = repo.Add(user.ID, "song1")
	if err != nil {
		t.Fatalf("Second add should not fail: %v", err)
	}

	// 数量应该只有 1
	count, err := repo.Count(user.ID)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestSQLiteFavoriteRepository_Remove(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLiteFavoriteRepository(db)

	repo.Add(user.ID, "song1")

	err := repo.Remove(user.ID, "song1")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	isFav, err := repo.IsFavorite(user.ID, "song1")
	if err != nil {
		t.Fatalf("IsFavorite failed: %v", err)
	}
	if isFav {
		t.Error("Expected song not to be favorite after removal")
	}
}

func TestSQLiteFavoriteRepository_IsFavorite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLiteFavoriteRepository(db)

	// 未添加时应该返回 false
	isFav, err := repo.IsFavorite(user.ID, "song1")
	if err != nil {
		t.Fatalf("IsFavorite failed: %v", err)
	}
	if isFav {
		t.Error("Expected song not to be favorite initially")
	}

	// 添加后应该返回 true
	repo.Add(user.ID, "song1")
	isFav, err = repo.IsFavorite(user.ID, "song1")
	if err != nil {
		t.Fatalf("IsFavorite failed: %v", err)
	}
	if !isFav {
		t.Error("Expected song to be favorite after adding")
	}
}

func TestSQLiteFavoriteRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLiteFavoriteRepository(db)

	repo.Add(user.ID, "song1")
	repo.Add(user.ID, "song2")
	repo.Add(user.ID, "song3")

	favorites, err := repo.GetByUserID(user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	if len(favorites) != 3 {
		t.Errorf("Expected 3 favorites, got %d", len(favorites))
	}

	// 测试分页
	favorites, err = repo.GetByUserID(user.ID, 2, 0)
	if err != nil {
		t.Fatalf("GetByUserID with limit failed: %v", err)
	}
	if len(favorites) != 2 {
		t.Errorf("Expected 2 favorites with limit, got %d", len(favorites))
	}
}

func TestSQLiteFavoriteRepository_GetSongIDs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLiteFavoriteRepository(db)

	repo.Add(user.ID, "song1")
	repo.Add(user.ID, "song2")

	songIDs, err := repo.GetSongIDs(user.ID)
	if err != nil {
		t.Fatalf("GetSongIDs failed: %v", err)
	}

	if len(songIDs) != 2 {
		t.Errorf("Expected 2 song IDs, got %d", len(songIDs))
	}
}

func TestSQLiteFavoriteRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewSQLiteUserRepository(db)
	user, _ := userRepo.Create("testuser", "test@example.com", "hash", "user")

	repo := NewSQLiteFavoriteRepository(db)

	// 初始计数应该为 0
	count, err := repo.Count(user.ID)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// 添加歌曲后计数增加
	repo.Add(user.ID, "song1")
	repo.Add(user.ID, "song2")

	count, err = repo.Count(user.ID)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}
