package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	t.Setenv("HOME", tempDir)
	logFile := "test.log"
	logger := New(logFile)

	tests := []struct {
		name    string
		level   string
		message string
		data    map[string]interface{}
		want    []string // Strings that should be in the log entry
	}{
		{
			name:    "info with no data",
			level:   "INFO",
			message: "test info message",
			data:    nil,
			want: []string{
				"[INFO]",
				"test info message",
			},
		},
		{
			name:    "error with data",
			level:   "ERROR",
			message: "test error message",
			data: map[string]interface{}{
				"error": "something went wrong",
				"code":  500,
			},
			want: []string{
				"[ERROR]",
				"test error message",
				"something went wrong",
				"500",
			},
		},
		{
			name:    "info with complex data",
			level:   "INFO",
			message: "test info with data",
			data: map[string]interface{}{
				"domain":    "example.com",
				"days_left": 30,
			},
			want: []string{
				"[INFO]",
				"test info with data",
				"example.com",
				"30",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the log file before each test
			logPath := filepath.Join(tempDir, ".certchecker", "logs", logFile)
			if err := os.WriteFile(logPath, []byte(""), 0644); err != nil {
				t.Fatalf("Failed to clear log file: %v", err)
			}

			switch tt.level {
			case "INFO":
				logger.Info(tt.message, tt.data)
			case "ERROR":
				logger.Error(tt.message, tt.data)
			}

			// Read log file
			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			logContent := string(content)

			// Check if all required strings are in the log content
			for _, want := range tt.want {
				if !strings.Contains(logContent, want) {
					t.Errorf("Log content missing %q, got: %s", want, logContent)
				}
			}
		})
	}
}

func TestLoggerCreateDirectory(t *testing.T) {
	// Test that logger creates directory if it doesn't exist
	tempDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	t.Setenv("HOME", tempDir)
	logFile := "test.log"
	logger := New(logFile)
	logger.Info("test message", nil)

	// Check if directory was created
	logDir := filepath.Join(tempDir, ".certchecker", "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Log directory was not created")
	}

	// Check if log file was created
	logPath := filepath.Join(logDir, logFile)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}
