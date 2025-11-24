package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir, err := os.MkdirTemp("", "logger_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// Test Init
	closer, err := Init(logFile)
	assert.NoError(t, err)
	assert.NotNil(t, closer)
	defer closer.Close()

	// Write a log to ensure file is created
	Info("test log message")

	// Verify log file exists
	_, err = os.Stat(logFile)
	assert.NoError(t, err)

	// Verify logger instance
	logger := GetLogger()
	assert.NotNil(t, logger)
	assert.IsType(t, &logrus.Logger{}, logger)
}

func TestWithRequestID(t *testing.T) {
	reqID := "test-request-id"
	entry := WithRequestID(reqID)
	assert.NotNil(t, entry)
	assert.Equal(t, reqID, entry.Data["request_id"])
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
		{"invalid", logrus.InfoLevel}, // Fallback
		{"", logrus.InfoLevel},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}
