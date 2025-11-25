// Package repository 提供数据访问层的 SQLite 实现。
package repository

import (
	"database/sql"
	"errors"
	"time"

	"zero-music/database"
	"zero-music/models"
)

// ErrNotFound 表示记录未找到的错误
var ErrNotFound = sql.ErrNoRows

// SQLiteUserRepository 是 UserRepository 的 SQLite 实现。
type SQLiteUserRepository struct {
	db database.DB
}

// NewSQLiteUserRepository 创建 SQLite 用户仓储实例。
func NewSQLiteUserRepository(db database.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// Create 创建新用户。
func (r *SQLiteUserRepository) Create(username, email, passwordHash string, role models.Role) (*models.User, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (username, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		username, email, passwordHash, role,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// FindByID 根据 ID 查找用户。
func (r *SQLiteUserRepository) FindByID(id int64) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// FindByUsername 根据用户名查找用户。
func (r *SQLiteUserRepository) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// FindByEmail 根据邮箱查找用户。
func (r *SQLiteUserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, updated_at 
		 FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Update 更新用户信息。
func (r *SQLiteUserRepository) Update(user *models.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET username = ?, email = ?, role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		user.Username, user.Email, user.Role, user.ID,
	)
	return err
}

// UpdatePassword 更新用户密码。
func (r *SQLiteUserRepository) UpdatePassword(userID int64, passwordHash string) error {
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		passwordHash, userID,
	)
	return err
}

// Delete 删除用户。
func (r *SQLiteUserRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

// List 获取所有用户。
func (r *SQLiteUserRepository) List() ([]*models.User, error) {
	rows, err := r.db.Query(
		`SELECT id, username, email, password_hash, role, created_at, updated_at FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// Exists 检查用户名或邮箱是否已存在。
func (r *SQLiteUserRepository) Exists(username, email string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE username = ? OR email = ?`,
		username, email,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
