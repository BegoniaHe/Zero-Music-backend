// Package database 提供 SQLite 数据库的具体实现。
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteProvider 是 SQLite 数据库的提供者实现。
type SQLiteProvider struct{}

// NewSQLiteProvider 创建 SQLite 提供者实例。
func NewSQLiteProvider() *SQLiteProvider {
	return &SQLiteProvider{}
}

// DriverName 返回驱动名称。
func (p *SQLiteProvider) DriverName() string {
	return "sqlite3"
}

// Open 打开 SQLite 数据库连接。
func (p *SQLiteProvider) Open(config *DBConfig) (DB, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(config.DSN)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	// 打开数据库连接，启用外键约束
	db, err := sql.Open("sqlite3", config.DSN+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 设置连接池参数
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}

	return &sqlDBWrapper{db: db}, nil
}

// Migrate 执行 SQLite 数据库迁移。
func (p *SQLiteProvider) Migrate(db DB) error {
	schemas := []string{
		// 用户表
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// 用户偏好设置表
		`CREATE TABLE IF NOT EXISTS user_preferences (
			user_id INTEGER PRIMARY KEY,
			preferences TEXT DEFAULT '{}',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// 播放历史表
		`CREATE TABLE IF NOT EXISTS play_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			song_id TEXT NOT NULL,
			played_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			play_duration INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// 播放统计表
		`CREATE TABLE IF NOT EXISTS play_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			song_id TEXT NOT NULL,
			play_count INTEGER DEFAULT 0,
			total_play_time INTEGER DEFAULT 0,
			last_played_at DATETIME,
			UNIQUE(user_id, song_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// 收藏表
		`CREATE TABLE IF NOT EXISTS favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			song_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, song_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// 播放列表表
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			cover_url TEXT DEFAULT '',
			is_smart BOOLEAN DEFAULT FALSE,
			smart_rules TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// 播放列表歌曲关联表
		`CREATE TABLE IF NOT EXISTS playlist_songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL,
			song_id TEXT NOT NULL,
			position INTEGER NOT NULL,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(playlist_id, song_id),
			FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
		)`,
		// 索引
		`CREATE INDEX IF NOT EXISTS idx_play_history_user_id ON play_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_play_history_song_id ON play_history(song_id)`,
		`CREATE INDEX IF NOT EXISTS idx_play_history_played_at ON play_history(played_at)`,
		`CREATE INDEX IF NOT EXISTS idx_play_stats_user_id ON play_stats(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_play_stats_song_id ON play_stats(song_id)`,
		`CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playlists_user_id ON playlists(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playlist_songs_playlist_id ON playlist_songs(playlist_id)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("执行 SQL 失败: %s, 错误: %w", schema, err)
		}
	}

	return nil
}

// sqlDBWrapper 包装 *sql.DB 以实现 DB 接口。
type sqlDBWrapper struct {
	db *sql.DB
}

func (w *sqlDBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return w.db.Exec(query, args...)
}

func (w *sqlDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w *sqlDBWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.Query(query, args...)
}

func (w *sqlDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func (w *sqlDBWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	return w.db.QueryRow(query, args...)
}

func (w *sqlDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

func (w *sqlDBWrapper) Prepare(query string) (*sql.Stmt, error) {
	return w.db.Prepare(query)
}

func (w *sqlDBWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.db.PrepareContext(ctx, query)
}

func (w *sqlDBWrapper) Begin() (*sql.Tx, error) {
	return w.db.Begin()
}

func (w *sqlDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.db.BeginTx(ctx, opts)
}

func (w *sqlDBWrapper) Ping() error {
	return w.db.Ping()
}

func (w *sqlDBWrapper) PingContext(ctx context.Context) error {
	return w.db.PingContext(ctx)
}

func (w *sqlDBWrapper) Close() error {
	return w.db.Close()
}

func (w *sqlDBWrapper) Stats() sql.DBStats {
	return w.db.Stats()
}

func (w *sqlDBWrapper) SetMaxOpenConns(n int) {
	w.db.SetMaxOpenConns(n)
}

func (w *sqlDBWrapper) SetMaxIdleConns(n int) {
	w.db.SetMaxIdleConns(n)
}
