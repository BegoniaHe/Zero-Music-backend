package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		expected string
	}{
		{
			name: "基本错误",
			apiError: &APIError{
				Code:    "TEST_ERROR",
				Message: "测试错误消息",
			},
			expected: "[TEST_ERROR] 测试错误消息",
		},
		{
			name: "带详情的错误",
			apiError: &APIError{
				Code:    "TEST_ERROR",
				Message: "测试错误",
				Details: "详细信息",
			},
			expected: "[TEST_ERROR] 测试错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("歌曲")

	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, "歌曲未找到", err.Message)
	assert.Empty(t, err.Details)
}

func TestNewBadRequestError(t *testing.T) {
	err := NewBadRequestError("参数无效")

	assert.Equal(t, "BAD_REQUEST", err.Code)
	assert.Equal(t, "参数无效", err.Message)
	assert.Empty(t, err.Details)
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("拒绝访问")

	assert.Equal(t, "FORBIDDEN", err.Code)
	assert.Equal(t, "拒绝访问", err.Message)
	assert.Empty(t, err.Details)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("未授权")

	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.Equal(t, "未授权", err.Message)
	assert.Empty(t, err.Details)
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("资源冲突")

	assert.Equal(t, "CONFLICT", err.Code)
	assert.Equal(t, "资源冲突", err.Message)
	assert.Empty(t, err.Details)
}

func TestNewInternalError_NoDebugMode(t *testing.T) {
	// 确保调试模式关闭
	os.Unsetenv("ZERO_MUSIC_DEBUG")

	originalErr := assert.AnError
	apiErr := NewInternalError(originalErr)

	assert.Equal(t, "INTERNAL_ERROR", apiErr.Code)
	assert.Equal(t, "内部服务器错误", apiErr.Message)
	assert.Empty(t, apiErr.Details, "非调试模式下不应暴露错误详情")
}

func TestNewInternalError_DebugMode(t *testing.T) {
	// 启用调试模式
	os.Setenv("ZERO_MUSIC_DEBUG", "true")
	defer os.Unsetenv("ZERO_MUSIC_DEBUG")

	testErr := assert.AnError
	apiErr := NewInternalError(testErr)

	assert.Equal(t, "INTERNAL_ERROR", apiErr.Code)
	assert.Equal(t, "内部服务器错误", apiErr.Message)
	assert.Equal(t, testErr.Error(), apiErr.Details, "调试模式下应暴露错误详情")
}

func TestValidateSongID(t *testing.T) {
	tests := []struct {
		name           string
		songID         string
		expectedResult bool
		expectedCode   int
	}{
		{
			name:           "有效的32字符十六进制ID",
			songID:         "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			expectedResult: true,
			expectedCode:   0, // 不检查响应码
		},
		{
			name:           "空ID",
			songID:         "",
			expectedResult: false,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "无效格式 - 太短",
			songID:         "abc123",
			expectedResult: false,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "无效格式 - 包含非法字符",
			songID:         "g1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			expectedResult: false,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "无效格式 - 路径遍历尝试",
			songID:         "../../../etc/passwd",
			expectedResult: false,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "无效格式 - 太长（64字符）",
			songID:         "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			expectedResult: false,
			expectedCode:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			result := ValidateSongID(c, tt.songID)

			assert.Equal(t, tt.expectedResult, result)

			if !tt.expectedResult {
				assert.Equal(t, tt.expectedCode, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "code")
			}
		})
	}
}

func TestIsDebugMode(t *testing.T) {
	// 测试关闭调试模式
	os.Unsetenv("ZERO_MUSIC_DEBUG")
	assert.False(t, isDebugMode())

	// 测试设置为 false
	os.Setenv("ZERO_MUSIC_DEBUG", "false")
	assert.False(t, isDebugMode())

	// 测试设置为 true
	os.Setenv("ZERO_MUSIC_DEBUG", "true")
	assert.True(t, isDebugMode())

	// 清理
	os.Unsetenv("ZERO_MUSIC_DEBUG")
}
