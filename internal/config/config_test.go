package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		yamlConfig *Config
		envVars    map[string]string
		want       *Config
		wantErr    bool
	}{
		{
			name: "valid yaml configuration only",
			yamlConfig: &Config{
				Domains:         []string{"example.com", "test.com"},
				ThresholdDays:   []int{7, 14, 30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HeartbeatHours:  24,
				IntervalHours:   12,
				HTTPEnabled:     true,
				HTTPPort:        8080,
				HTTPAuthToken:   "test-token",
			},
			want: &Config{
				Domains:         []string{"example.com", "test.com"},
				ThresholdDays:   []int{7, 14, 30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HeartbeatHours:  24,
				IntervalHours:   12,
				HTTPEnabled:     true,
				HTTPPort:        8080,
				HTTPAuthToken:   "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid env configuration only",
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
			name: "env vars override yaml config",
			yamlConfig: &Config{
				Domains:         []string{"yaml.com"},
				ThresholdDays:   []int{10, 20},
				SlackWebhookURL: "https://hooks.slack.com/services/yaml",
				HeartbeatHours:  48,
				IntervalHours:   24,
				HTTPEnabled:     false,
				HTTPPort:        9090,
				HTTPAuthToken:   "yaml-token",
			},
			envVars: map[string]string{
				"DOMAINS":           "env.com",
				"THRESHOLD_DAYS":    "15,30",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/env",
				"HEARTBEAT_HOURS":   "72",
				"HTTP_ENABLED":      "true",
				"HTTP_PORT":         "8080",
				"HTTP_AUTH_TOKEN":   "env-token",
			},
			want: &Config{
				Domains:         []string{"env.com"},
				ThresholdDays:   []int{15, 30},
				SlackWebhookURL: "https://hooks.slack.com/services/env",
				HeartbeatHours:  72,
				IntervalHours:   24,
				HTTPEnabled:     true,
				HTTPPort:        8080,
				HTTPAuthToken:   "env-token",
			},
			wantErr: false,
		},
		{
			name: "minimal yaml configuration",
			yamlConfig: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
			},
			want: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				IntervalHours:   6,  // default value
				HTTPPort:        8080, // default value
			},
			wantErr: false,
		},
		{
			name: "missing required fields in yaml",
			yamlConfig: &Config{
				Domains: []string{"example.com"},
			},
			wantErr: true,
		},
		{
			name: "invalid threshold days in env",
			envVars: map[string]string{
				"DOMAINS":           "example.com",
				"THRESHOLD_DAYS":    "invalid",
				"SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
			},
			wantErr: true,
		},
		{
			name: "invalid heartbeat hours in env",
			yamlConfig: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
			},
			envVars: map[string]string{
				"HEARTBEAT_HOURS": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid interval hours in env",
			yamlConfig: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
			},
			envVars: map[string]string{
				"CHECK_INTERVAL_HOURS": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid HTTP port in env",
			yamlConfig: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HTTPEnabled:     true,
				HTTPAuthToken:   "test-token",
			},
			envVars: map[string]string{
				"HTTP_PORT": "invalid",
			},
			wantErr: true,
		},
		{
			name: "HTTP enabled without token in yaml",
			yamlConfig: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HTTPEnabled:     true,
				HTTPPort:        8080,
			},
			wantErr: true,
		},
		{
			name: "HTTP port out of range in yaml",
			yamlConfig: &Config{
				Domains:         []string{"example.com"},
				ThresholdDays:   []int{30},
				SlackWebhookURL: "https://hooks.slack.com/services/xxx",
				HTTPEnabled:     true,
				HTTPPort:        70000,
				HTTPAuthToken:   "test-token",
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

			// Create config.yaml if yamlConfig is provided
			if tt.yamlConfig != nil {
				yamlData, err := yaml.Marshal(tt.yamlConfig)
				if err != nil {
					t.Fatalf("Failed to marshal YAML config: %v", err)
				}
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), yamlData, 0644); err != nil {
					t.Fatalf("Failed to write config.yaml: %v", err)
				}
			}

			// Create .env file if envVars are provided
			if tt.envVars != nil {
				var envContent strings.Builder
				for k, v := range tt.envVars {
					envContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
				}
				if err := os.WriteFile(filepath.Join(configDir, ".env"), []byte(envContent.String()), 0644); err != nil {
					t.Fatalf("Failed to write .env file: %v", err)
				}
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