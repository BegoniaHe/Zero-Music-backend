package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"zero-music/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupSystemTestEnv 初始化一个用于系统处理器测试的环境。
func setupSystemTestEnv(t *testing.T, musicDirExists bool) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var musicDir string
	if musicDirExists {
		musicDir = t.TempDir()
	} else {
		// 使用一个不存在的目录路径
		musicDir = filepath.Join(t.TempDir(), "nonexistent_dir")
	}

	cfg := &config.Config{
		Music: config.MusicConfig{
			Directory:        musicDir,
			SupportedFormats: []string{".mp3"},
			CacheTTLMinutes:  5,
		},
	}

	router := gin.New()
	handler := NewSystemHandler(cfg)
	router.GET("/health", handler.HealthCheck)
	router.GET("/", handler.APIIndex)

	return router, musicDir
}

// TestHealthCheck_OK 测试当音乐目录可访问时健康检查返回 OK。
func TestHealthCheck_OK(t *testing.T) {
	router, musicDir := setupSystemTestEnv(t, true)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, true, response["music_dir_accessible"])
	assert.Equal(t, musicDir, response["music_directory"])
	assert.Contains(t, response["message"], "服务器正在运行")
}

// TestHealthCheck_Degraded 测试当音乐目录不可访问时健康检查返回 degraded。
func TestHealthCheck_Degraded(t *testing.T) {
	router, musicDir := setupSystemTestEnv(t, false)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "degraded", response["status"])
	assert.Equal(t, false, response["music_dir_accessible"])
	assert.Equal(t, musicDir, response["music_directory"])
}

// TestHealthCheck_DirectoryRemoved 测试当音乐目录在运行时被删除后的行为。
func TestHealthCheck_DirectoryRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建一个临时目录
	musicDir := t.TempDir()
	subDir := filepath.Join(musicDir, "music_subdir")
	err := os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	cfg := &config.Config{
		Music: config.MusicConfig{
			Directory:        subDir,
			SupportedFormats: []string{".mp3"},
			CacheTTLMinutes:  5,
		},
	}

	router := gin.New()
	handler := NewSystemHandler(cfg)
	router.GET("/health", handler.HealthCheck)

	// 第一次检查：目录存在
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 删除目录
	err = os.Remove(subDir)
	assert.NoError(t, err)

	// 第二次检查：目录不存在
	req, _ = http.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "degraded", response["status"])
}

// TestAPIIndex 测试 API 索引端点返回正确的信息。
func TestAPIIndex(t *testing.T) {
	router, _ := setupSystemTestEnv(t, true)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证必要字段存在
	assert.Equal(t, "zero music API", response["name"])
	assert.Equal(t, APIVersion, response["version"])

	// 验证 endpoints 字段存在且为数组
	endpoints, ok := response["endpoints"].([]interface{})
	assert.True(t, ok, "endpoints 应该是一个数组")
	assert.Greater(t, len(endpoints), 0, "endpoints 不应为空")

	// 验证包含关键端点
	endpointsStr := make([]string, len(endpoints))
	for i, e := range endpoints {
		endpointsStr[i] = e.(string)
	}
	assert.Contains(t, endpointsStr, "GET /health - 健康检查")
	assert.Contains(t, endpointsStr, "GET /api/songs - 获取所有歌曲列表")
}

// TestNewSystemHandler 测试 NewSystemHandler 构造函数。
func TestNewSystemHandler(t *testing.T) {
	cfg := &config.Config{
		Music: config.MusicConfig{
			Directory: "/test/path",
		},
	}

	handler := NewSystemHandler(cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.cfg)
}
