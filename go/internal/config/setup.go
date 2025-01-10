package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Domains         []string
	ThresholdDays   []string
	SlackWebhookURL string
}

func RunSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🔧 SSL Certificate Checker Configuration Setup")
	fmt.Println("===========================================")

	// Get domains
	fmt.Print("\n📋 Enter domains to monitor (comma-separated, e.g., example.com,test.example.com): ")
	domainsStr, _ := reader.ReadString('\n')
	domains := strings.TrimSpace(domainsStr)

	// Get threshold days
	fmt.Print("\n⏰ Enter alert threshold days (comma-separated, e.g., 7,14,30,45): ")
	thresholdStr, _ := reader.ReadString('\n')
	thresholds := strings.TrimSpace(thresholdStr)

	// Get Slack webhook URL
	fmt.Print("\n🔔 Enter Slack webhook URL: ")
	webhookURL, _ := reader.ReadString('\n')
	webhookURL = strings.TrimSpace(webhookURL)

	// Create config
	config := Config{
		Domains:         strings.Split(domains, ","),
		ThresholdDays:   strings.Split(thresholds, ","),
		SlackWebhookURL: webhookURL,
	}

	// Validate config
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Write config file
	if err := writeConfig(config); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	fmt.Println("\n✅ Configuration saved successfully!")
	return nil
}

func validateConfig(config Config) error {
	if len(config.Domains) == 0 || (len(config.Domains) == 1 && config.Domains[0] == "") {
		return fmt.Errorf("at least one domain must be specified")
	}

	if len(config.ThresholdDays) == 0 || (len(config.ThresholdDays) == 1 && config.ThresholdDays[0] == "") {
		return fmt.Errorf("at least one threshold day must be specified")
	}

	if config.SlackWebhookURL == "" {
		return fmt.Errorf("Slack webhook URL must be specified")
	}

	return nil
}

func writeConfig(config Config) error {
	// Find config directory
	configDir := filepath.Join(os.Getenv("HOME"), ".certchecker", "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Create .env file content
	content := fmt.Sprintf(`# Comma-separated list of domains to monitor
DOMAINS=%s

# Comma-separated list of days before expiration to send alerts
THRESHOLD_DAYS=%s

# Slack webhook URL for notifications
SLACK_WEBHOOK_URL=%s
`, strings.Join(config.Domains, ","),
		strings.Join(config.ThresholdDays, ","),
		config.SlackWebhookURL)

	// Write to .env file
	envPath := filepath.Join(configDir, ".env")
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
} 