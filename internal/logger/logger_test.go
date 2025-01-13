package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		level    string
		details  map[string]interface{}
		wantLogs []string
	}{
		{
			name:    "info with no data",
			message: "test info message",
			level:   "INFO",
			wantLogs: []string{
				"[INFO]",
				"test info message",
			},
		},
		{
			name:    "error with data",
			message: "test error message",
			level:   "ERROR",
			details: map[string]interface{}{
				"code":  500,
				"error": "something went wrong",
			},
			wantLogs: []string{
				"[ERROR]",
				"test error message",
				`"code": 500`,
				`"error": "something went wrong"`,
			},
		},
		{
			name:    "info with complex data",
			message: "test info with data",
			level:   "INFO",
			details: map[string]interface{}{
				"domain":    "example.com",
				"days_left": 30,
			},
			wantLogs: []string{
				"[INFO]",
				"test info with data",
				`"domain": "example.com"`,
				`"days_left": 30`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for test
			tempDir := t.TempDir()

			// Initialize logger with temp directory
			logger := New(tempDir)

			// Log the message
			switch tt.level {
			case "INFO":
				logger.Info(tt.message, tt.details)
			case "ERROR":
				logger.Error(tt.message, tt.details)
			case "WARNING":
				logger.Warning(tt.message, tt.details)
			}

			// Read the log file
			content, err := os.ReadFile(logger.logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			// Check if all expected strings are in the log
			logContent := string(content)
			for _, want := range tt.wantLogs {
				if !strings.Contains(logContent, want) {
					t.Errorf("Log does not contain %q\nLog content:\n%s", want, logContent)
				}
			}
		})
	}
}

func TestLoggerCreateDirectory(t *testing.T) {
	// Create a temporary directory for test
	tempDir := t.TempDir()

	// Initialize logger with temp directory
	logger := New(tempDir)

	// Log a message
	logger.Info("test message", nil)

	// Check if log directory was created
	logDir := filepath.Join(tempDir, ".certchecker", "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Log directory was not created")
	}

	// Check if log file was created
	if _, err := os.Stat(logger.logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Check if log file contains the message
	content, err := os.ReadFile(logger.logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "test message") {
		t.Errorf("Log does not contain test message\nLog content:\n%s", logContent)
	}
}
