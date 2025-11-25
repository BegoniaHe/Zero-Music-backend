package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteProvider_DriverName(t *testing.T) {
	provider := NewSQLiteProvider()
	assert.Equal(t, "sqlite3", provider.DriverName())
}

func TestSQLiteProvider_Open_Success(t *testing.T) {
	provider := NewSQLiteProvider()

	// 使用临时目录
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := &DBConfig{
		DSN:          dbPath,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	db, err := provider.Open(config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// 验证连接
	err = db.Ping()
	assert.NoError(t, err)
}

func TestSQLiteProvider_Open_CreatesDirectory(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "dir")
	dbPath := filepath.Join(nestedDir, "test.db")

	config := &DBConfig{
		DSN: dbPath,
	}

	db, err := provider.Open(config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// 验证目录已创建
	_, err = os.Stat(nestedDir)
	assert.NoError(t, err)
}

func TestSQLiteProvider_Migrate(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "migrate_test.db")

	config := &DBConfig{
		DSN: dbPath,
	}

	db, err := provider.Open(config)
	require.NoError(t, err)
	defer db.Close()

	// 执行迁移
	err = provider.Migrate(db)
	require.NoError(t, err)

	// 验证表已创建
	tables := []string{"users", "user_preferences", "play_history", "play_stats", "favorites", "playlists", "playlist_songs"}

	for _, table := range tables {
		t.Run("table_exists_"+table, func(t *testing.T) {
			var count int
			err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count, "table %s should exist", table)
		})
	}
}

func TestSQLiteProvider_Migrate_Idempotent(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "idempotent_test.db")

	config := &DBConfig{
		DSN: dbPath,
	}

	db, err := provider.Open(config)
	require.NoError(t, err)
	defer db.Close()

	// 多次执行迁移应该不会报错
	for i := 0; i < 3; i++ {
		err = provider.Migrate(db)
		assert.NoError(t, err, "migration %d should succeed", i+1)
	}
}

func TestSqlDBWrapper_Exec(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "exec_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// 创建测试表
	result, err := db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// 插入数据
	result, err = db.Exec(`INSERT INTO test_table (name) VALUES (?)`, "test")
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	affected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
}

func TestSqlDBWrapper_Query(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "query_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// 准备测试数据
	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_table (name) VALUES (?), (?), (?)`, "a", "b", "c")
	require.NoError(t, err)

	// 查询数据
	rows, err := db.Query(`SELECT id, name FROM test_table ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		require.NoError(t, err)
		names = append(names, name)
	}

	assert.Equal(t, []string{"a", "b", "c"}, names)
	assert.NoError(t, rows.Err())
}

func TestSqlDBWrapper_QueryRow(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "queryrow_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// 准备测试数据
	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_table (name) VALUES (?)`, "test_name")
	require.NoError(t, err)

	// 查询单行
	var name string
	err = db.QueryRow(`SELECT name FROM test_table WHERE id = ?`, 1).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "test_name", name)
}

func TestSqlDBWrapper_Begin(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tx_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	// 开始事务
	tx, err := db.Begin()
	require.NoError(t, err)

	// 在事务中插入
	_, err = tx.Exec(`INSERT INTO test_table (name) VALUES (?)`, "tx_test")
	require.NoError(t, err)

	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	// 验证数据已保存
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM test_table`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSqlDBWrapper_Begin_Rollback(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "rollback_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	// 开始事务
	tx, err := db.Begin()
	require.NoError(t, err)

	// 在事务中插入
	_, err = tx.Exec(`INSERT INTO test_table (name) VALUES (?)`, "rollback_test")
	require.NoError(t, err)

	// 回滚事务
	err = tx.Rollback()
	require.NoError(t, err)

	// 验证数据未保存
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM test_table`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSqlDBWrapper_Prepare(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "prepare_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	// 准备语句
	stmt, err := db.Prepare(`INSERT INTO test_table (name) VALUES (?)`)
	require.NoError(t, err)
	defer stmt.Close()

	// 使用准备的语句多次执行
	for _, name := range []string{"a", "b", "c"} {
		_, err := stmt.Exec(name)
		require.NoError(t, err)
	}

	// 验证插入的数据
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM test_table`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSqlDBWrapper_Stats(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "stats_test.db")

	config := &DBConfig{
		DSN:          dbPath,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	db, err := provider.Open(config)
	require.NoError(t, err)
	defer db.Close()

	stats := db.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
}

func TestSqlDBWrapper_SetConnPoolParams(t *testing.T) {
	provider := NewSQLiteProvider()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "pool_test.db")

	db, err := provider.Open(&DBConfig{DSN: dbPath})
	require.NoError(t, err)
	defer db.Close()

	// 设置连接池参数
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	stats := db.Stats()
	assert.Equal(t, 20, stats.MaxOpenConnections)
}
