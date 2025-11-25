package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"zero-music/middleware"
	"zero-music/models"
	"zero-music/repository"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	tokenExpiration time.Duration
	userRepo        repository.UserRepository
	jwtManager      *middleware.JWTManager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(tokenExpiration time.Duration, userRepo repository.UserRepository, jwtManager *middleware.JWTManager) *AuthHandler {
	return &AuthHandler{
		tokenExpiration: tokenExpiration,
		userRepo:        userRepo,
		jwtManager:      jwtManager,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string               `json:"token"`
	User  *models.UserResponse `json:"user"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError(fmt.Sprintf("请求参数错误: %s", err.Error())))
		return
	}

	// 验证用户名格式（只允许字母、数字、下划线）
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, NewBadRequestError("用户名只能包含字母、数字和下划线"))
		return
	}

	// 检查用户是否已存在
	exists, err := h.userRepo.Exists(req.Username, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}
	if exists {
		c.JSON(http.StatusConflict, NewConflictError("用户名或邮箱已被使用"))
		return
	}

	// 生成密码哈希
	passwordHash, err := models.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 创建用户
	user, err := h.userRepo.Create(req.Username, req.Email, passwordHash, models.RoleUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 生成令牌
	token, err := h.jwtManager.GenerateToken(user, h.tokenExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "注册成功",
		"data": AuthResponse{
			Token: token,
			User:  user.ToResponse(),
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	// 支持用户名或邮箱登录
	var user *models.User
	var err error

	if strings.Contains(req.Username, "@") {
		user, err = h.userRepo.FindByEmail(req.Username)
	} else {
		user, err = h.userRepo.FindByUsername(req.Username)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	if user == nil || !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("用户名或密码错误"))
		return
	}

	// 生成令牌
	token, err := h.jwtManager.GenerateToken(user, h.tokenExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data": AuthResponse{
			Token: token,
			User:  user.ToResponse(),
		},
	})
}

// GetProfile 获取当前用户资料
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("未登录"))
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("用户"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    user.ToResponse(),
	})
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("未登录"))
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("用户"))
		return
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := h.userRepo.FindByEmail(req.Email)
		if existingUser != nil {
			c.JSON(http.StatusConflict, NewConflictError("邮箱已被使用"))
			return
		}
		user.Email = req.Email
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    user.ToResponse(),
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("未登录"))
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewBadRequestError("请求参数错误"))
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, NewNotFoundError("用户"))
		return
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("原密码错误"))
		return
	}

	// 生成新密码哈希
	passwordHash, err := models.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	// 更新密码
	if err := h.userRepo.UpdatePassword(userID, passwordHash); err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "密码修改成功",
	})
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("未登录"))
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, NewUnauthorizedError("用户不存在"))
		return
	}

	// 生成新令牌
	token, err := h.jwtManager.GenerateToken(user, h.tokenExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "刷新成功",
		"data": gin.H{
			"token": token,
		},
	})
}
