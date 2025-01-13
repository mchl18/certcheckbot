package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Domains         []string
	ThresholdDays   []int
	SlackWebhookURL string
	HeartbeatHours  int
	IntervalHours   int
	HTTPEnabled     bool
	HTTPPort        int
	HTTPAuthToken   string
}

func Load(homeDir string) (*Config, error) {
	configDir := filepath.Join(homeDir, ".certchecker", "config")
	envPath := filepath.Join(configDir, ".env")

	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Clear any existing environment variables
	os.Unsetenv("DOMAINS")
	os.Unsetenv("THRESHOLD_DAYS")
	os.Unsetenv("SLACK_WEBHOOK_URL")
	os.Unsetenv("HEARTBEAT_HOURS")
	os.Unsetenv("CHECK_INTERVAL_HOURS")
	os.Unsetenv("HTTP_ENABLED")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("HTTP_AUTH_TOKEN")

	// Load environment variables from the .env file
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	domains := strings.Split(os.Getenv("DOMAINS"), ",")
	if len(domains) == 0 || domains[0] == "" {
		return nil, fmt.Errorf("DOMAINS environment variable is required")
	}

	thresholdDaysStr := os.Getenv("THRESHOLD_DAYS")
	if thresholdDaysStr == "" {
		return nil, fmt.Errorf("THRESHOLD_DAYS environment variable is required")
	}

	thresholdDays := []int{}
	for _, d := range strings.Split(thresholdDaysStr, ",") {
		days, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil {
			return nil, fmt.Errorf("invalid THRESHOLD_DAYS value: %w", err)
		}
		thresholdDays = append(thresholdDays, days)
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return nil, fmt.Errorf("SLACK_WEBHOOK_URL environment variable is required")
	}

	heartbeatHoursStr := os.Getenv("HEARTBEAT_HOURS")
	var heartbeatHours int
	if heartbeatHoursStr != "" {
		var err error
		heartbeatHours, err = strconv.Atoi(heartbeatHoursStr)
		if err != nil {
			return nil, fmt.Errorf("invalid HEARTBEAT_HOURS value: %w", err)
		}
	}

	intervalHoursStr := os.Getenv("CHECK_INTERVAL_HOURS")
	intervalHours := 6 // default value
	if intervalHoursStr != "" {
		var err error
		intervalHours, err = strconv.Atoi(intervalHoursStr)
		if err != nil {
			return nil, fmt.Errorf("invalid CHECK_INTERVAL_HOURS value: %w", err)
		}
	}

	httpEnabled := os.Getenv("HTTP_ENABLED") == "true"
	var httpPort int
	var httpAuthToken string

	if httpEnabled {
		httpAuthToken = os.Getenv("HTTP_AUTH_TOKEN")
		if httpAuthToken == "" {
			return nil, fmt.Errorf("HTTP_AUTH_TOKEN is required when HTTP server is enabled")
		}

		httpPortStr := os.Getenv("HTTP_PORT")
		if httpPortStr == "" {
			httpPort = 8080 // default value
		} else {
			var err error
			httpPort, err = strconv.Atoi(httpPortStr)
			if err != nil {
				return nil, fmt.Errorf("invalid HTTP_PORT value: %w", err)
			}
			if httpPort < 1 || httpPort > 65535 {
				return nil, fmt.Errorf("HTTP_PORT must be between 1 and 65535")
			}
		}
	}

	return &Config{
		Domains:         domains,
		ThresholdDays:   thresholdDays,
		SlackWebhookURL: webhookURL,
		HeartbeatHours:  heartbeatHours,
		IntervalHours:   intervalHours,
		HTTPEnabled:     httpEnabled,
		HTTPPort:        httpPort,
		HTTPAuthToken:   httpAuthToken,
	}, nil
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

	// Create or truncate .env file
	envPath := filepath.Join(configDir, ".env")
	file, err := os.Create(envPath)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %v", err)
	}
	defer file.Close()

	// Get user input
	fmt.Print("Enter domains to monitor (comma-separated): ")
	domains, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read domains: %v", err)
	}
	domains = strings.TrimSpace(domains)
	if domains == "" {
		return fmt.Errorf("domains cannot be empty")
	}

	// Threshold days
	fmt.Print("Enter threshold days for alerts (comma-separated): ")
	thresholdDays, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read threshold days: %v", err)
	}
	thresholdDays = strings.TrimSpace(thresholdDays)
	if thresholdDays == "" {
		return fmt.Errorf("threshold days cannot be empty")
	}

	// Slack webhook URL
	fmt.Print("Enter Slack webhook URL: ")
	slackWebhookURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read Slack webhook URL: %v", err)
	}
	slackWebhookURL = strings.TrimSpace(slackWebhookURL)
	if slackWebhookURL == "" {
		return fmt.Errorf("Slack webhook URL cannot be empty")
	}

	// Heartbeat hours
	fmt.Print("Enter heartbeat interval in hours (optional, press Enter to skip): ")
	heartbeatHours, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read heartbeat hours: %v", err)
	}
	heartbeatHours = strings.TrimSpace(heartbeatHours)
	if heartbeatHours != "" {
		if _, err := strconv.Atoi(heartbeatHours); err != nil {
			return fmt.Errorf("invalid heartbeat hours: %v", err)
		}
	}

	// Check interval hours
	fmt.Print("Enter check interval in hours (optional, default 6): ")
	intervalHours, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read check interval hours: %v", err)
	}
	intervalHours = strings.TrimSpace(intervalHours)
	if intervalHours != "" {
		if _, err := strconv.Atoi(intervalHours); err != nil {
			return fmt.Errorf("invalid check interval hours: %v", err)
		}
	}

	// HTTP server settings
	fmt.Print("Enable HTTP server? (y/N): ")
	enableHTTP, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read HTTP server enable: %v", err)
	}
	enableHTTP = strings.ToLower(strings.TrimSpace(enableHTTP))

	var httpPort, httpAuthToken string
	if enableHTTP == "y" || enableHTTP == "yes" {
		fmt.Print("Enter HTTP server port (optional, default 8080): ")
		httpPort, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read HTTP port: %v", err)
		}
		httpPort = strings.TrimSpace(httpPort)
		if httpPort != "" {
			port, err := strconv.Atoi(httpPort)
			if err != nil {
				return fmt.Errorf("invalid HTTP port: %v", err)
			}
			if port < 1 || port > 65535 {
				return fmt.Errorf("HTTP port must be between 1 and 65535")
			}
		}

		fmt.Print("Enter HTTP authentication token: ")
		httpAuthToken, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read HTTP auth token: %v", err)
		}
		httpAuthToken = strings.TrimSpace(httpAuthToken)
		if httpAuthToken == "" {
			return fmt.Errorf("HTTP authentication token cannot be empty when HTTP server is enabled")
		}
	}

	// Write configuration to file
	var config strings.Builder
	config.WriteString(fmt.Sprintf("DOMAINS=%s\n", domains))
	config.WriteString(fmt.Sprintf("THRESHOLD_DAYS=%s\n", thresholdDays))
	config.WriteString(fmt.Sprintf("SLACK_WEBHOOK_URL=%s\n", slackWebhookURL))
	if heartbeatHours != "" {
		config.WriteString(fmt.Sprintf("HEARTBEAT_HOURS=%s\n", heartbeatHours))
	}
	if intervalHours != "" {
		config.WriteString(fmt.Sprintf("CHECK_INTERVAL_HOURS=%s\n", intervalHours))
	}
	if enableHTTP == "y" || enableHTTP == "yes" {
		config.WriteString("HTTP_ENABLED=true\n")
		if httpPort != "" {
			config.WriteString(fmt.Sprintf("HTTP_PORT=%s\n", httpPort))
		}
		config.WriteString(fmt.Sprintf("HTTP_AUTH_TOKEN=%s\n", httpAuthToken))
	}

	if _, err := file.WriteString(config.String()); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	fmt.Printf("Configuration saved to %s\n", envPath)
	return nil
}

func RunSetup() error {
	reader := bufio.NewReader(os.Stdin)
	return runSetupWithReader(reader)
} 