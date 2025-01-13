package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"DOMAINS":           "example.com,test.com",
				"THRESHOLD_DAYS":    "7,14,30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
				"HEARTBEAT_HOURS":   "24",
				"HTTP_ENABLED":      "true",
				"HTTP_PORT":         "8080",
				"HTTP_AUTH_TOKEN":   "test-token",
			},
			want: &Config{
				Domains:         []string{"example.com", "test.com"},
				ThresholdDays:   []int{7, 14, 30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HeartbeatHours:  24,
				IntervalHours:   6, // default value
				HTTPEnabled:     true,
				HTTPPort:        8080,
				HTTPAuthToken:   "test-token",
			},
			wantErr: false,
		},
		{
			name: "minimal configuration",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
			},
			want: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HeartbeatHours:  0,
				IntervalHours:   6, // default value
				HTTPEnabled:     false,
				HTTPPort:        0,
				HTTPAuthToken:   "",
			},
			wantErr: false,
		},
		{
			name: "missing domains",
			envVars: map[string]string{
				"THRESHOLD_DAYS":    "30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
			},
			wantErr: true,
		},
		{
			name: "missing threshold days",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
			},
			wantErr: true,
		},
		{
			name: "missing webhook URL",
			envVars: map[string]string{
				"DOMAINS":        "example.com",
				"THRESHOLD_DAYS": "30",
			},
			wantErr: true,
		},
		{
			name: "invalid threshold days",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "invalid",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
			},
			wantErr: true,
		},
		{
			name: "invalid heartbeat hours",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
				"HEARTBEAT_HOURS":   "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid interval hours",
			envVars: map[string]string{
				"DOMAINS":             "example.com",
				"THRESHOLD_DAYS":      "30",
				"SLACK_WEBHOOK_URL":   "https://hooks.slack.com/services/xxx",
				"CHECK_INTERVAL_HOURS": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid HTTP port",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
				"HTTP_ENABLED":      "true",
				"HTTP_PORT":         "invalid",
				"HTTP_AUTH_TOKEN":   "test-token",
			},
			wantErr: true,
		},
		{
			name: "HTTP enabled without token",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
				"HTTP_ENABLED":      "true",
				"HTTP_PORT":         "8080",
			},
			wantErr: true,
		},
		{
			name: "HTTP port out of range",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
				"HTTP_ENABLED":      "true",
				"HTTP_PORT":         "70000",
				"HTTP_AUTH_TOKEN":   "test-token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new temporary directory for each test case
			testDir := filepath.Join(tempDir, tt.name)
			configDir := filepath.Join(testDir, ".certchecker", "config")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config directory: %v", err)
			}

			// Create .env file with test variables
			envPath := filepath.Join(configDir, ".env")
			var envContent strings.Builder
			for k, v := range tt.envVars {
				envContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}

			if err := os.WriteFile(envPath, []byte(envContent.String()), 0644); err != nil {
				t.Fatalf("Failed to write .env file: %v", err)
			}

			got, err := Load(testDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(got.Domains, tt.want.Domains) {
					t.Errorf("Load() domains = %v, want %v", got.Domains, tt.want.Domains)
				}
				if !reflect.DeepEqual(got.ThresholdDays, tt.want.ThresholdDays) {
					t.Errorf("Load() thresholds = %v, want %v", got.ThresholdDays, tt.want.ThresholdDays)
				}
				if got.SlackWebhookURL != tt.want.SlackWebhookURL {
					t.Errorf("Load() webhook URL = %v, want %v", got.SlackWebhookURL, tt.want.SlackWebhookURL)
				}
				if got.HeartbeatHours != tt.want.HeartbeatHours {
					t.Errorf("Load() heartbeat hours = %v, want %v", got.HeartbeatHours, tt.want.HeartbeatHours)
				}
				if got.IntervalHours != tt.want.IntervalHours {
					t.Errorf("Load() interval hours = %v, want %v", got.IntervalHours, tt.want.IntervalHours)
				}
				if got.HTTPEnabled != tt.want.HTTPEnabled {
					t.Errorf("Load() HTTP enabled = %v, want %v", got.HTTPEnabled, tt.want.HTTPEnabled)
				}
				if got.HTTPPort != tt.want.HTTPPort {
					t.Errorf("Load() HTTP port = %v, want %v", got.HTTPPort, tt.want.HTTPPort)
				}
				if got.HTTPAuthToken != tt.want.HTTPAuthToken {
					t.Errorf("Load() HTTP auth token = %v, want %v", got.HTTPAuthToken, tt.want.HTTPAuthToken)
				}
			}
		})
	}
} 