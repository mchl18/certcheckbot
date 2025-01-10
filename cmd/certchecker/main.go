package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mchl18/certcheckbot/internal/checker"
	"github.com/mchl18/certcheckbot/internal/config"
	"github.com/mchl18/certcheckbot/internal/logger"
)

const (
	checkInterval = 6 * time.Hour
)

func main() {
	// Check for config command
	if len(os.Args) > 1 && os.Args[1] == "config" {
		if err := config.RunSetup(); err != nil {
			log.Fatal("Configuration failed:", err)
		}
		return
	}

	// Get home directory for .certchecker
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory:", err)
	}

	// Try to load .env from .certchecker config
	envPath := filepath.Join(home, ".certchecker", "config", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("No .env file found. Run 'certchecker config' to create one.")
	}

	// Parse configuration
	domains := strings.Split(os.Getenv("DOMAINS"), ",")
	thresholdDaysStr := strings.Split(os.Getenv("THRESHOLD_DAYS"), ",")
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")

	// Validate configuration
	if slackWebhookURL == "" {
		log.Fatal("SLACK_WEBHOOK_URL is not set")
	}
	if len(domains) == 0 || domains[0] == "" {
		log.Fatal("DOMAINS is not set")
	}
	if len(thresholdDaysStr) == 0 || thresholdDaysStr[0] == "" {
		log.Fatal("THRESHOLD_DAYS is not set")
	}

	// Convert threshold days to integers
	thresholdDays := make([]int, len(thresholdDaysStr))
	for i, dayStr := range thresholdDaysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			log.Fatalf("Invalid threshold day value: %s", dayStr)
		}
		thresholdDays[i] = day
	}

	// Initialize logger
	logger := logger.New(filepath.Join(home, "logs", "cert-checker.log"))

	// Initialize certificate checker with project root
	certChecker := checker.New(domains, thresholdDays, slackWebhookURL, logger, home)

	// Run initial check
	logger.Info("Certificate monitoring service initialization", map[string]interface{}{
		"startTime": time.Now().Format(time.RFC3339),
		"configuration": map[string]interface{}{
			"domains":       domains,
			"thresholds":    thresholdDays,
			"checkInterval": checkInterval.String(),
		},
	})

	runCheck(certChecker, logger)

	// Schedule periodic checks
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		runCheck(certChecker, logger)
	}
}

func runCheck(certChecker *checker.CertificateChecker, logger *logger.Logger) {
	logger.Info("Beginning certificate check cycle", map[string]interface{}{
		"domains":       certChecker.Domains(),
		"thresholds":    certChecker.ThresholdDays(),
		"checkInterval": checkInterval.String(),
	})

	if err := certChecker.CheckAll(); err != nil {
		logger.Error("Certificate check cycle failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	logger.Info("Certificate check cycle completed successfully", nil)
}
