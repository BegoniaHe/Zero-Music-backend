package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// DefaultLogLevel 是默认的日志级别
	DefaultLogLevel = "info"

	// defaultLogMaxSizeMB 定义单个日志文件的最大大小（MB）
	defaultLogMaxSizeMB = 100
	// defaultLogMaxBackups 定义保留的日志文件数量
	defaultLogMaxBackups = 7
	// defaultLogMaxAgeDays 定义日志文件的最长保留天数
	defaultLogMaxAgeDays = 30
)

var (
	log        *logrus.Logger
	loggerOnce sync.Once
)

// ensureLogger 初始化全局日志实例（仅初始化一次）。
func ensureLogger() {
	loggerOnce.Do(func() {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		log.SetOutput(os.Stdout)
		log.SetLevel(parseLogLevel(os.Getenv("LOG_LEVEL")))
	})
}

func parseLogLevel(level string) logrus.Level {
	if level == "" {
		level = DefaultLogLevel
	}
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		// 由于 logger 可能尚未配置好，这里使用标准输出进行降级日志
		fmt.Fprintf(os.Stdout, "[logger] invalid log level '%s', fallback to '%s'\n", level, DefaultLogLevel)
		return logrus.InfoLevel
	}
	return parsedLevel
}

// Init 初始化日志系统并启用日志轮转。
// 返回的 io.Closer 需要在应用关闭时显式关闭，以确保缓冲区刷新。
func Init(logFilePath string) (io.Closer, error) {
	ensureLogger()

	if logFilePath == "" {
		return nil, fmt.Errorf("log file path cannot be empty")
	}

	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory failed: %w", err)
	}

	rotateWriter := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    defaultLogMaxSizeMB,
		MaxBackups: defaultLogMaxBackups,
		MaxAge:     defaultLogMaxAgeDays,
		Compress:   true,
	}

	log.SetOutput(io.MultiWriter(os.Stdout, rotateWriter))
	log.SetLevel(parseLogLevel(os.Getenv("LOG_LEVEL")))

	return rotateWriter, nil
}

// GetLogger 返回全局日志实例。
func GetLogger() *logrus.Logger {
	ensureLogger()
	return log
}

// WithRequestID 创建带有请求 ID 的日志条目。
func WithRequestID(requestID string) *logrus.Entry {
	return GetLogger().WithField("request_id", requestID)
}

// Info 记录信息级别日志。
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 格式化记录信息级别日志。
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 记录警告级别日志。
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 格式化记录警告级别日志。
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 记录错误级别日志。
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 格式化记录错误级别日志。
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 记录致命错误并退出。
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 格式化记录致命错误并退出。
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
