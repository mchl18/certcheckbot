package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSetup(t *testing.T) {
	// Create a temporary directory for test config
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	t.Setenv("HOME", tempDir)

	// Test cases
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, envPath string)
	}{
		{
			name: "valid configuration",
			input: `example.com,test.com
7,14,30
https://hooks.slack.com/services/test
24
6
y
8080
test-token
`,
			wantErr: false,
			validate: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Errorf("Failed to read .env file: %v", err)
					return
				}

				// Check required fields
				fields := map[string]string{
					"DOMAINS":           "example.com,test.com",
					"THRESHOLD_DAYS":    "7,14,30",
					"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/test",
					"HEARTBEAT_HOURS":   "24",
					"INTERVAL_HOURS":    "6",
					"HTTP_ENABLED":      "true",
					"HTTP_PORT":         "8080",
					"HTTP_AUTH_TOKEN":   "test-token",
				}

				for key, want := range fields {
					if !strings.Contains(string(content), key+"="+want) {
						t.Errorf("Missing or incorrect %s: want %s", key, want)
					}
				}
			},
		},
		{
			name: "minimal configuration",
			input: `example.com
30
https://hooks.slack.com/services/test


n


`,
			wantErr: false,
			validate: func(t *testing.T, envPath string) {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Errorf("Failed to read .env file: %v", err)
					return
				}

				// Check required fields
				fields := map[string]string{
					"DOMAINS":           "example.com",
					"THRESHOLD_DAYS":    "30",
					"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/test",
				}

				for key, want := range fields {
					if !strings.Contains(string(content), key+"="+want) {
						t.Errorf("Missing or incorrect %s: want %s", key, want)
					}
				}

				// Check optional fields are not set
				optionalFields := []string{
					"HEARTBEAT_HOURS",
					"HTTP_ENABLED",
					"HTTP_PORT",
					"HTTP_AUTH_TOKEN",
				}

				for _, field := range optionalFields {
					if strings.Contains(string(content), field+"=") {
						t.Errorf("Optional field %s should not be set", field)
					}
				}
			},
		},
		{
			name: "empty domains",
			input: `
30
https://hooks.slack.com/services/test


n


`,
			wantErr: true,
		},
		{
			name: "empty threshold days",
			input: `example.com

https://hooks.slack.com/services/test


n


`,
			wantErr: true,
		},
		{
			name: "empty webhook URL",
			input: `example.com
30



n


`,
			wantErr: true,
		},
		{
			name: "invalid heartbeat hours",
			input: `example.com
30
https://hooks.slack.com/services/test
invalid
6
n


`,
			wantErr: true,
		},
		{
			name: "invalid interval hours",
			input: `example.com
30
https://hooks.slack.com/services/test

invalid
n


`,
			wantErr: true,
		},
		{
			name: "invalid HTTP port",
			input: `example.com
30
https://hooks.slack.com/services/test


y
invalid
test-token
`,
			wantErr: true,
		},
		{
			name: "HTTP enabled without token",
			input: `example.com
30
https://hooks.slack.com/services/test


y
8080

`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for test
			tempDir := t.TempDir()
			t.Setenv("HOME", tempDir)

			// Create a reader with the test input
			reader := bufio.NewReader(strings.NewReader(tt.input))
			err := runSetupWithReader(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("runSetupWithReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				envPath := filepath.Join(tempDir, ".certchecker", "config", ".env")
				tt.validate(t, envPath)
			}
		})
	}
}
