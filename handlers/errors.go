package handlers

import (
	"fmt"
	"net/http"
	"os"

	"zero-music/logger"
	"zero-music/middleware"
	"zero-music/models"

	"github.com/gin-gonic/gin"
)

// isDebugMode 返回是否启用调试模式（每次调用时读取环境变量，允许运行时动态切换）。
func isDebugMode() bool {
	return os.Getenv("ZERO_MUSIC_DEBUG") == "true"
}

// ValidateSongID 验证歌曲 ID 格式，确保是有效的 SHA256 哈希格式，防止路径遍历攻击。
// 返回 true 表示验证通过，返回 false 表示验证失败（并已向客户端发送错误响应）。
func ValidateSongID(c *gin.Context, id string) bool {
	requestID := middleware.GetRequestID(c)

	if id == "" {
		logger.WithRequestID(requestID).Warn("歌曲 ID 为空")
		c.JSON(http.StatusBadRequest, NewBadRequestError("歌曲ID不能为空"))
		return false
	}

	if !models.ValidIDRegex.MatchString(id) {
		logger.WithRequestID(requestID).Warnf("无效的歌曲 ID 格式: %s", id)
		c.JSON(http.StatusBadRequest, NewBadRequestError("无效的歌曲 ID 格式"))
		return false
	}

	return true
}

// APIError 定义了 API 返回的标准化错误结构。
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现了标准错误接口。
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewNotFoundError 创建一个表示资源未找到的 APIError。
func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s未找到", resource),
	}
}

// NewInternalError 创建一个表示内部服务器错误的 APIError。
// 默认不暴露错误详情，仅在明确启用调试模式时显示（ZERO_MUSIC_DEBUG=true）。
func NewInternalError(err error) *APIError {
	apiErr := &APIError{
		Code:    "INTERNAL_ERROR",
		Message: "内部服务器错误",
	}

	// 仅在明确启用调试模式时暴露错误详情，默认不暴露以提高安全性
	if isDebugMode() {
		apiErr.Details = err.Error()
	}

	return apiErr
}

// NewBadRequestError 创建一个表示无效请求的 APIError。
func NewBadRequestError(message string) *APIError {
	return &APIError{
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

// NewForbiddenError 创建一个表示禁止访问的 APIError。
func NewForbiddenError(message string) *APIError {
	return &APIError{
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// NewUnauthorizedError 创建一个表示未授权的 APIError。
func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// NewConflictError 创建一个表示资源冲突的 APIError。
func NewConflictError(message string) *APIError {
	return &APIError{
		Code:    "CONFLICT",
		Message: message,
	}
}
