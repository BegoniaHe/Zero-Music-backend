package repository

import (
	"testing"

	"zero-music/models"
)

func TestSQLiteUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	user, err := repo.Create("testuser", "test@example.com", "hashedpassword", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.Role != models.RoleUser {
		t.Errorf("Expected role '%s', got '%s'", models.RoleUser, user.Role)
	}
}

func TestSQLiteUserRepository_Create_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	_, err := repo.Create("testuser", "test1@example.com", "hash1", models.RoleUser)
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	_, err = repo.Create("testuser", "test2@example.com", "hash2", models.RoleUser)
	if err == nil {
		t.Error("Expected error for duplicate username, got nil")
	}
}

func TestSQLiteUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	created, err := repo.Create("testuser", "test@example.com", "hash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find user, got nil")
	}
	if found.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", found.Username)
	}
}

func TestSQLiteUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	found, err := repo.FindByID(999)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestSQLiteUserRepository_FindByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	_, err := repo.Create("testuser", "test@example.com", "hash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByUsername("testuser")
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find user, got nil")
	}
	if found.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", found.Email)
	}
}

func TestSQLiteUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	_, err := repo.Create("testuser", "test@example.com", "hash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByEmail("test@example.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find user, got nil")
	}
	if found.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", found.Username)
	}
}

func TestSQLiteUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	user, err := repo.Create("testuser", "test@example.com", "hash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	user.Username = "updateduser"
	user.Role = models.RoleAdmin
	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.Username != "updateduser" {
		t.Errorf("Expected username 'updateduser', got '%s'", found.Username)
	}
	if found.Role != models.RoleAdmin {
		t.Errorf("Expected role '%s', got '%s'", models.RoleAdmin, found.Role)
	}
}

func TestSQLiteUserRepository_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	user, err := repo.Create("testuser", "test@example.com", "oldhash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.UpdatePassword(user.ID, "newhash")
	if err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}

	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.PasswordHash != "newhash" {
		t.Errorf("Expected password hash 'newhash', got '%s'", found.PasswordHash)
	}
}

func TestSQLiteUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	user, err := repo.Create("testuser", "test@example.com", "hash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found != nil {
		t.Error("Expected user to be deleted")
	}
}

func TestSQLiteUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	_, err := repo.Create("user1", "user1@example.com", "hash1", models.RoleUser)
	if err != nil {
		t.Fatalf("Create user1 failed: %v", err)
	}
	_, err = repo.Create("user2", "user2@example.com", "hash2", models.RoleAdmin)
	if err != nil {
		t.Fatalf("Create user2 failed: %v", err)
	}

	users, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestSQLiteUserRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)

	_, err := repo.Create("testuser", "test@example.com", "hash", models.RoleUser)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test existing username
	exists, err := repo.Exists("testuser", "other@example.com")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected user to exist by username")
	}

	// Test existing email
	exists, err = repo.Exists("otheruser", "test@example.com")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected user to exist by email")
	}

	// Test non-existing
	exists, err = repo.Exists("nonexistent", "nonexistent@example.com")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected user not to exist")
	}
}
