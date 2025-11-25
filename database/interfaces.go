// Package database 定义了数据库访问层的接口抽象。
// 通过接口隔离数据库实现，便于在不同数据库后端之间切换。
package database

import (
	"context"
	"database/sql"
	"io"
)

// DB 定义了数据库连接的抽象接口。
// 该接口封装了 database/sql 的核心功能，允许不同数据库后端的实现。
type DB interface {
	// Exec 执行不返回行的 SQL 语句（如 INSERT、UPDATE、DELETE）。
	Exec(query string, args ...interface{}) (sql.Result, error)

	// ExecContext 带上下文执行不返回行的 SQL 语句。
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Query 执行返回多行的 SQL 查询。
	Query(query string, args ...interface{}) (*sql.Rows, error)

	// QueryContext 带上下文执行返回多行的 SQL 查询。
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow 执行返回最多一行的 SQL 查询。
	QueryRow(query string, args ...interface{}) *sql.Row

	// QueryRowContext 带上下文执行返回最多一行的 SQL 查询。
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Prepare 创建预编译语句。
	Prepare(query string) (*sql.Stmt, error)

	// PrepareContext 带上下文创建预编译语句。
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	// Begin 开始一个数据库事务。
	Begin() (*sql.Tx, error)

	// BeginTx 带上下文和选项开始一个数据库事务。
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Ping 验证数据库连接是否可用。
	Ping() error

	// PingContext 带上下文验证数据库连接是否可用。
	PingContext(ctx context.Context) error

	// Close 关闭数据库连接。
	Close() error

	// Stats 返回数据库连接池统计信息。
	Stats() sql.DBStats

	// SetMaxOpenConns 设置最大打开连接数。
	SetMaxOpenConns(n int)

	// SetMaxIdleConns 设置最大空闲连接数。
	SetMaxIdleConns(n int)
}

// DBConfig 定义了数据库连接配置。
type DBConfig struct {
	// Driver 数据库驱动类型: "sqlite3", "mysql", "postgres"，等
	Driver string `json:"driver"`

	// DSN 数据源名称/连接字符串
	// SQLite: 文件路径，如 "data/zero-music.db"
	// MySQL: "user:password@tcp(host:port)/dbname?parseTime=true"
	// PostgreSQL: "host=localhost port=5432 user=xxx password=xxx dbname=xxx sslmode=disable"
	DSN string `json:"dsn"`

	// MaxOpenConns 最大打开连接数（0 表示无限制）
	MaxOpenConns int `json:"max_open_conns"`

	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int `json:"max_idle_conns"`
}

// DBProvider 定义了数据库提供者接口。
// 不同的数据库后端（SQLite、MySQL、PostgreSQL）需要实现此接口。
type DBProvider interface {
	// Open 打开数据库连接。
	Open(config *DBConfig) (DB, error)

	// Migrate 执行数据库迁移/初始化表结构。
	Migrate(db DB) error

	// DriverName 返回驱动名称。
	DriverName() string
}

// DBManager 管理数据库连接的生命周期。
type DBManager struct {
	db       DB
	provider DBProvider
	config   *DBConfig
}

// NewDBManager 创建数据库管理器实例。
func NewDBManager(provider DBProvider, config *DBConfig) *DBManager {
	return &DBManager{
		provider: provider,
		config:   config,
	}
}

// Connect 连接到数据库并执行迁移。
func (m *DBManager) Connect() error {
	db, err := m.provider.Open(m.config)
	if err != nil {
		return err
	}

	if err := m.provider.Migrate(db); err != nil {
		db.Close()
		return err
	}

	m.db = db
	return nil
}

// GetDB 返回数据库连接实例。
func (m *DBManager) GetDB() DB {
	return m.db
}

// Close 关闭数据库连接。
func (m *DBManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// Closer 返回一个 io.Closer 用于优雅关闭。
func (m *DBManager) Closer() io.Closer {
	return &dbCloser{manager: m}
}

type dbCloser struct {
	manager *DBManager
}

func (c *dbCloser) Close() error {
	return c.manager.Close()
}
