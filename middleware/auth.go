package middleware

import (
	"net/http"
	"strings"
	"time"

	"zero-music/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig 定义 JWT 配置。
type JWTConfig struct {
	Secret []byte
}

// JWTManager 管理 JWT 令牌的生成和验证。
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager 创建 JWT 管理器实例。
func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		config: &JWTConfig{
			Secret: []byte(secret),
		},
	}
}

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   int64       `json:"user_id"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 使用 JWTManager 生成JWT令牌
func (m *JWTManager) GenerateToken(user *models.User, expireDuration time.Duration) (string, error) {
	return generateTokenWithSecret(user, expireDuration, m.config.Secret)
}

// generateTokenWithSecret 使用指定密钥生成JWT令牌
func generateTokenWithSecret(user *models.User, expireDuration time.Duration, secret []byte) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zero-music",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 使用 JWTManager 解析JWT令牌
func (m *JWTManager) ParseToken(tokenString string) (*JWTClaims, error) {
	return parseTokenWithSecret(tokenString, m.config.Secret)
}

// parseTokenWithSecret 使用指定密钥解析JWT令牌
func parseTokenWithSecret(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// JWTAuth 创建 JWT 认证中间件，需要传入 JWTManager 实例。
func JWTAuth(manager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "JWT管理器未初始化",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请提供认证令牌",
			})
			c.Abort()
			return
		}

		// Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		claims, err := manager.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalJWTAuth 创建可选 JWT 认证中间件（不强制要求认证）。
func OptionalJWTAuth(manager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || manager == nil {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			if claims, err := manager.ParseToken(parts[1]); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
				c.Set("claims", claims)
			}
		}

		c.Next()
	}
}

// AdminOnly 仅管理员可访问中间件
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		if role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无权限访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	// 使用安全的类型断言，避免 panic
	id, ok := userID.(int64)
	if !ok {
		return 0, false
	}
	return id, true
}

// GetCurrentUsername 从上下文获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	// 使用安全的类型断言，避免 panic
	name, ok := username.(string)
	if !ok {
		return "", false
	}
	return name, true
}

// GetCurrentRole 从上下文获取当前用户角色
func GetCurrentRole(c *gin.Context) (models.Role, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	// 使用安全的类型断言，避免 panic
	r, ok := role.(models.Role)
	if !ok {
		return "", false
	}
	return r, true
}
