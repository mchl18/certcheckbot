package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/madbook/certchecker/internal/checker"
	"github.com/madbook/certchecker/internal/config"
	"github.com/madbook/certchecker/internal/logger"
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

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}
	projectRoot := filepath.Join(wd, "..")

	// Try to load .env from various locations
	envPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".certchecker", "config", ".env"),
		filepath.Join(projectRoot, ".env"),
		".env",
	}

	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			break
		}
	}

	if !envLoaded {
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
	logger := logger.New(filepath.Join(projectRoot, "logs", "cert-checker.log"))

	// Initialize certificate checker with project root
	certChecker := checker.New(domains, thresholdDays, slackWebhookURL, logger, projectRoot)

	// Run initial check
	logger.Info("Certificate monitoring service initialization", map[string]interface{}{
		"startTime": time.Now().Format(time.RFC3339),
		"configuration": map[string]interface{}{
			"domains":       domains,
			"thresholds":   thresholdDays,
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