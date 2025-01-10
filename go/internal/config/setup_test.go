package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockReader struct {
	responses []string
	index     int
}

func (m *mockReader) ReadString(delim byte) (string, error) {
	if m.index >= len(m.responses) {
		return "", io.EOF
	}
	response := m.responses[m.index]
	m.index++
	return response + "\n", nil
}

func TestRunSetup(t *testing.T) {
	// Create a temporary directory for test config
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config directory
	configDir := filepath.Join(tempDir, ".certchecker", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Set up test cases
	tests := []struct {
		name      string
		responses []string
		wantErr   bool
	}{
		{
			name: "valid configuration",
			responses: []string{
				"example.com,test.com",           // Domains
				"7,14,30",                        // Threshold days
				"https://hooks.slack.com/test",   // Webhook URL
			},
			wantErr: false,
		},
		{
			name: "invalid domains",
			responses: []string{
				"",                               // Empty domains
				"7,14,30",
				"https://hooks.slack.com/test",
			},
			wantErr: true,
		},
		{
			name: "invalid thresholds",
			responses: []string{
				"example.com",
				"invalid,days",                   // Invalid threshold days
				"https://hooks.slack.com/test",
			},
			wantErr: true,
		},
		{
			name: "empty webhook",
			responses: []string{
				"example.com",
				"7,14,30",
				"",                               // Empty webhook URL
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock reader
			mockReader := &mockReader{responses: tt.responses}

			// Create test environment
			envFile := filepath.Join(configDir, ".env")

			// Run setup with mock input
			err := runSetupWithReader(envFile, mockReader)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("RunSetup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If setup should succeed, verify the config file
			if !tt.wantErr {
				content, err := os.ReadFile(envFile)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}

				config := string(content)

				// Check if all required fields are present
				if !strings.Contains(config, "DOMAINS=") {
					t.Error("Config file missing DOMAINS")
				}
				if !strings.Contains(config, "THRESHOLD_DAYS=") {
					t.Error("Config file missing THRESHOLD_DAYS")
				}
				if !strings.Contains(config, "SLACK_WEBHOOK_URL=") {
					t.Error("Config file missing SLACK_WEBHOOK_URL")
				}

				// Verify values
				for _, response := range tt.responses {
					if !strings.Contains(config, response) {
						t.Errorf("Config file missing value: %s", response)
					}
				}
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		field   string
		wantErr bool
	}{
		{
			name:    "valid domains",
			input:   "example.com,test.com",
			field:   "domains",
			wantErr: false,
		},
		{
			name:    "empty domains",
			input:   "",
			field:   "domains",
			wantErr: true,
		},
		{
			name:    "valid thresholds",
			input:   "7,14,30",
			field:   "thresholds",
			wantErr: false,
		},
		{
			name:    "invalid thresholds",
			input:   "invalid,days",
			field:   "thresholds",
			wantErr: true,
		},
		{
			name:    "valid webhook",
			input:   "https://hooks.slack.com/test",
			field:   "webhook",
			wantErr: false,
		},
		{
			name:    "empty webhook",
			input:   "",
			field:   "webhook",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.input, tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} 