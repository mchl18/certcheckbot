package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Domains         []string `yaml:"domains"`
	ThresholdDays   []int    `yaml:"threshold_days"`
	SlackWebhookURL string   `yaml:"slack_webhook_url"`
	HeartbeatHours  int      `yaml:"heartbeat_hours"`
	IntervalHours   int      `yaml:"interval_hours"`
	HTTPEnabled     bool     `yaml:"http_enabled"`
	HTTPPort        int      `yaml:"http_port"`
	HTTPAuthToken   string   `yaml:"http_auth_token"`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) (int, error) {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s: %w", key, err)
	}
	return value, nil
}

func Load(homeDir string) (*Config, error) {
	configDir := filepath.Join(homeDir, ".certchecker", "config")
	configPath := filepath.Join(configDir, "config.yaml")
	envPath := filepath.Join(configDir, ".env")

	// Initialize with default values
	config := &Config{
		IntervalHours: 6,
		HTTPPort:      8080,
	}

	// Try to load YAML config first
	yamlExists := false
	if yamlData, err := os.ReadFile(configPath); err == nil {
		// Create a temporary config to unmarshal into
		tempConfig := &Config{}
		if err := yaml.Unmarshal(yamlData, tempConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
		}
		
		// Copy values while preserving defaults if not set
		config.Domains = tempConfig.Domains
		config.ThresholdDays = tempConfig.ThresholdDays
		config.SlackWebhookURL = tempConfig.SlackWebhookURL
		config.HeartbeatHours = tempConfig.HeartbeatHours
		config.HTTPEnabled = tempConfig.HTTPEnabled
		config.HTTPAuthToken = tempConfig.HTTPAuthToken
		
		// Only override defaults if explicitly set in YAML
		if tempConfig.IntervalHours != 0 {
			config.IntervalHours = tempConfig.IntervalHours
		}
		if tempConfig.HTTPPort != 0 {
			config.HTTPPort = tempConfig.HTTPPort
		}
		
		yamlExists = true
	}

	// Clear any existing environment variables that might interfere with our tests
	os.Unsetenv("DOMAINS")
	os.Unsetenv("THRESHOLD_DAYS")
	os.Unsetenv("SLACK_WEBHOOK_URL")
	os.Unsetenv("HEARTBEAT_HOURS")
	os.Unsetenv("CHECK_INTERVAL_HOURS")
	os.Unsetenv("HTTP_ENABLED")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("HTTP_AUTH_TOKEN")

	// Load .env file if it exists (for backward compatibility)
	envExists := false
	if _, err := os.Stat(envPath); err == nil {
		if envData, err := os.ReadFile(envPath); err == nil {
			envMap := make(map[string]string)
			for _, line := range strings.Split(string(envData), "\n") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = strings.TrimSpace(parts[1])
				}
			}
			
			// Set environment variables from .env file
			for k, v := range envMap {
				os.Setenv(k, v)
			}
			envExists = true
		}
	}

	// Only use environment variables if no YAML config exists or if they're from a .env file
	if !yamlExists || envExists {
		// Override with environment variables if they exist
		if domains := os.Getenv("DOMAINS"); domains != "" {
			config.Domains = strings.Split(domains, ",")
		}

		if thresholdDays := os.Getenv("THRESHOLD_DAYS"); thresholdDays != "" {
			days := []int{}
			for _, d := range strings.Split(thresholdDays, ",") {
				day, err := strconv.Atoi(strings.TrimSpace(d))
				if err != nil {
					return nil, fmt.Errorf("invalid THRESHOLD_DAYS value: %w", err)
				}
				days = append(days, day)
			}
			config.ThresholdDays = days
		}

		if webhookURL := os.Getenv("SLACK_WEBHOOK_URL"); webhookURL != "" {
			config.SlackWebhookURL = webhookURL
		}

		if heartbeatHours, err := getEnvIntOrDefault("HEARTBEAT_HOURS", config.HeartbeatHours); err != nil {
			return nil, err
		} else {
			config.HeartbeatHours = heartbeatHours
		}

		if intervalHours, err := getEnvIntOrDefault("CHECK_INTERVAL_HOURS", config.IntervalHours); err != nil {
			return nil, err
		} else {
			config.IntervalHours = intervalHours
		}

		if httpEnabled := os.Getenv("HTTP_ENABLED"); httpEnabled != "" {
			config.HTTPEnabled = httpEnabled == "true"
		}

		if httpPort, err := getEnvIntOrDefault("HTTP_PORT", config.HTTPPort); err != nil {
			return nil, err
		} else {
			config.HTTPPort = httpPort
		}

		if httpAuthToken := os.Getenv("HTTP_AUTH_TOKEN"); httpAuthToken != "" {
			config.HTTPAuthToken = httpAuthToken
		}
	}

	// Validate required fields
	if len(config.Domains) == 0 {
		return nil, fmt.Errorf("domains must be specified either in config.yaml or DOMAINS environment variable")
	}

	if len(config.ThresholdDays) == 0 {
		return nil, fmt.Errorf("threshold days must be specified either in config.yaml or THRESHOLD_DAYS environment variable")
	}

	if config.SlackWebhookURL == "" {
		return nil, fmt.Errorf("Slack webhook URL must be specified either in config.yaml or SLACK_WEBHOOK_URL environment variable")
	}

	if config.HTTPEnabled {
		if config.HTTPAuthToken == "" {
			return nil, fmt.Errorf("HTTP auth token is required when HTTP server is enabled")
		}
		if config.HTTPPort < 1 || config.HTTPPort > 65535 {
			return nil, fmt.Errorf("HTTP port must be between 1 and 65535")
		}
	}

	return config, nil
}

func runSetupWithReader(reader *bufio.Reader) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create .certchecker/config directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".certchecker", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	config := Config{
		IntervalHours: 6,
		HTTPPort:      8080,
	}

	// Get user input
	fmt.Print("Enter domains to monitor (comma-separated): ")
	domains, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read domains: %v", err)
	}
	config.Domains = strings.Split(strings.TrimSpace(domains), ",")

	fmt.Print("Enter threshold days for alerts (comma-separated): ")
	thresholdDaysStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read threshold days: %v", err)
	}
	for _, d := range strings.Split(strings.TrimSpace(thresholdDaysStr), ",") {
		days, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil {
			return fmt.Errorf("invalid threshold days: %v", err)
		}
		config.ThresholdDays = append(config.ThresholdDays, days)
	}

	fmt.Print("Enter Slack webhook URL: ")
	webhookURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read Slack webhook URL: %v", err)
	}
	config.SlackWebhookURL = strings.TrimSpace(webhookURL)

	fmt.Print("Enter heartbeat interval in hours (optional, press Enter to skip): ")
	heartbeatHours, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read heartbeat hours: %v", err)
	}
	if heartbeatStr := strings.TrimSpace(heartbeatHours); heartbeatStr != "" {
		hours, err := strconv.Atoi(heartbeatStr)
		if err != nil {
			return fmt.Errorf("invalid heartbeat hours: %v", err)
		}
		config.HeartbeatHours = hours
	}

	fmt.Print("Enter check interval in hours (optional, default 6): ")
	intervalHours, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read check interval hours: %v", err)
	}
	if intervalStr := strings.TrimSpace(intervalHours); intervalStr != "" {
		hours, err := strconv.Atoi(intervalStr)
		if err != nil {
			return fmt.Errorf("invalid check interval hours: %v", err)
		}
		config.IntervalHours = hours
	}

	fmt.Print("Enable HTTP server? (y/N): ")
	enableHTTP, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read HTTP server enable: %v", err)
	}
	enableHTTP = strings.ToLower(strings.TrimSpace(enableHTTP))

	if enableHTTP == "y" || enableHTTP == "yes" {
		config.HTTPEnabled = true

		fmt.Print("Enter HTTP server port (optional, default 8080): ")
		httpPort, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read HTTP port: %v", err)
		}
		if portStr := strings.TrimSpace(httpPort); portStr != "" {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid HTTP port: %v", err)
			}
			if port < 1 || port > 65535 {
				return fmt.Errorf("HTTP port must be between 1 and 65535")
			}
			config.HTTPPort = port
		}

		fmt.Print("Enter HTTP authentication token: ")
		httpAuthToken, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read HTTP auth token: %v", err)
		}
		config.HTTPAuthToken = strings.TrimSpace(httpAuthToken)
		if config.HTTPAuthToken == "" {
			return fmt.Errorf("HTTP authentication token cannot be empty when HTTP server is enabled")
		}
	}

	// Save as YAML
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	fmt.Printf("Configuration saved to %s\n", configPath)
	return nil
}

func RunSetup() error {
	reader := bufio.NewReader(os.Stdin)
	return runSetupWithReader(reader)
} 