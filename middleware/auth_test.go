package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"zero-music/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret"
	manager := NewJWTManager(secret)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.Equal(t, []byte(secret), manager.config.Secret)
}

func TestJWTManager_GenerateAndParseToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// 生成令牌
	token, err := manager.GenerateToken(user, 24*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 解析令牌
	claims, err := manager.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "zero-music", claims.Issuer)
}

func TestJWTManager_ParseToken_Invalid(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "空令牌",
			token: "",
		},
		{
			name:  "无效格式",
			token: "invalid-token",
		},
		{
			name:  "错误签名",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.wrong-signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ParseToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestJWTManager_ParseToken_DifferentSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-1")
	manager2 := NewJWTManager("secret-2")

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := manager1.GenerateToken(user, 24*time.Hour)
	require.NoError(t, err)

	// 使用不同密钥解析应该失败
	claims, err := manager2.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTAuth_Success(t *testing.T) {
	manager := NewJWTManager("test-secret")
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := manager.GenerateToken(user, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := JWTAuth(manager)
	middleware(c)

	assert.False(t, c.IsAborted())
	userID, exists := c.Get("user_id")
	assert.True(t, exists)
	assert.Equal(t, int64(1), userID)
}

func TestJWTAuth_NoHeader(t *testing.T) {
	manager := NewJWTManager("test-secret")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	middleware := JWTAuth(manager)
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	manager := NewJWTManager("test-secret")

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "缺少Bearer前缀",
			header: "some-token",
		},
		{
			name:   "错误的前缀",
			header: "Basic some-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request.Header.Set("Authorization", tt.header)

			middleware := JWTAuth(manager)
			middleware(c)

			assert.True(t, c.IsAborted())
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestJWTAuth_NilManager(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer some-token")

	middleware := JWTAuth(nil)
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOptionalJWTAuth_WithValidToken(t *testing.T) {
	manager := NewJWTManager("test-secret")
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     models.RoleAdmin,
	}

	token, err := manager.GenerateToken(user, 24*time.Hour)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := OptionalJWTAuth(manager)
	middleware(c)

	assert.False(t, c.IsAborted())
	userID, exists := c.Get("user_id")
	assert.True(t, exists)
	assert.Equal(t, int64(1), userID)
}

func TestOptionalJWTAuth_NoHeader(t *testing.T) {
	manager := NewJWTManager("test-secret")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	middleware := OptionalJWTAuth(manager)
	middleware(c)

	// 应该继续而不是中止
	assert.False(t, c.IsAborted())
	_, exists := c.Get("user_id")
	assert.False(t, exists)
}

func TestOptionalJWTAuth_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	middleware := OptionalJWTAuth(manager)
	middleware(c)

	// 应该继续而不是中止，但不设置用户信息
	assert.False(t, c.IsAborted())
	_, exists := c.Get("user_id")
	assert.False(t, exists)
}

func TestAdminOnly_AsAdmin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("role", models.RoleAdmin)

	middleware := AdminOnly()
	middleware(c)

	assert.False(t, c.IsAborted())
}

func TestAdminOnly_AsUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("role", models.RoleUser)

	middleware := AdminOnly()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminOnly_NoRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	middleware := AdminOnly()
	middleware(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUserID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*gin.Context)
		expectedID int64
		expectedOK bool
	}{
		{
			name: "有效用户ID",
			setup: func(c *gin.Context) {
				c.Set("user_id", int64(42))
			},
			expectedID: 42,
			expectedOK: true,
		},
		{
			name:       "未设置用户ID",
			setup:      func(c *gin.Context) {},
			expectedID: 0,
			expectedOK: false,
		},
		{
			name: "错误类型的用户ID",
			setup: func(c *gin.Context) {
				c.Set("user_id", "not-an-int")
			},
			expectedID: 0,
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c)

			id, ok := GetCurrentUserID(c)

			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func TestGetCurrentUsername(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(*gin.Context)
		expectedUsername string
		expectedOK       bool
	}{
		{
			name: "有效用户名",
			setup: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			expectedUsername: "testuser",
			expectedOK:       true,
		},
		{
			name:             "未设置用户名",
			setup:            func(c *gin.Context) {},
			expectedUsername: "",
			expectedOK:       false,
		},
		{
			name: "错误类型的用户名",
			setup: func(c *gin.Context) {
				c.Set("username", 123)
			},
			expectedUsername: "",
			expectedOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c)

			username, ok := GetCurrentUsername(c)

			assert.Equal(t, tt.expectedUsername, username)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func TestGetCurrentRole(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*gin.Context)
		expectedRole models.Role
		expectedOK   bool
	}{
		{
			name: "管理员角色",
			setup: func(c *gin.Context) {
				c.Set("role", models.RoleAdmin)
			},
			expectedRole: models.RoleAdmin,
			expectedOK:   true,
		},
		{
			name: "普通用户角色",
			setup: func(c *gin.Context) {
				c.Set("role", models.RoleUser)
			},
			expectedRole: models.RoleUser,
			expectedOK:   true,
		},
		{
			name:         "未设置角色",
			setup:        func(c *gin.Context) {},
			expectedRole: "",
			expectedOK:   false,
		},
		{
			name: "错误类型的角色",
			setup: func(c *gin.Context) {
				c.Set("role", "invalid-type")
			},
			expectedRole: "",
			expectedOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c)

			role, ok := GetCurrentRole(c)

			assert.Equal(t, tt.expectedRole, role)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}
