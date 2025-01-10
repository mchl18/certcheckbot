package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type reader interface {
	ReadString(byte) (string, error)
}

func RunSetup() error {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create config directory
	configDir := filepath.Join(home, ".certchecker", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set up config file path
	configFile := filepath.Join(configDir, ".env")

	return runSetupWithReader(configFile, bufio.NewReader(os.Stdin))
}

func runSetupWithReader(configFile string, reader reader) error {
	fmt.Print("Enter domains to monitor (comma-separated): ")
	domains, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read domains: %w", err)
	}
	domains = strings.TrimSpace(domains)
	if err := validateInput(domains, "domains"); err != nil {
		return err
	}

	fmt.Print("Enter threshold days (comma-separated): ")
	thresholds, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read thresholds: %w", err)
	}
	thresholds = strings.TrimSpace(thresholds)
	if err := validateInput(thresholds, "thresholds"); err != nil {
		return err
	}

	fmt.Print("Enter Slack webhook URL: ")
	webhook, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read webhook URL: %w", err)
	}
	webhook = strings.TrimSpace(webhook)
	if err := validateInput(webhook, "webhook"); err != nil {
		return err
	}

	// Create config file
	content := fmt.Sprintf(`DOMAINS=%s
THRESHOLD_DAYS=%s
SLACK_WEBHOOK_URL=%s
`, domains, thresholds, webhook)

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration saved to %s\n", configFile)
	return nil
}

func validateInput(input, field string) error {
	if input == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}

	if field == "thresholds" {
		// Validate threshold days are numbers
		for _, day := range strings.Split(input, ",") {
			if _, err := strconv.Atoi(strings.TrimSpace(day)); err != nil {
				return fmt.Errorf("invalid threshold day: %s", day)
			}
		}
	}

	return nil
} 