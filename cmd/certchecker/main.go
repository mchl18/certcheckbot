package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mchl18/certcheckbot/internal/checker"
	"github.com/mchl18/certcheckbot/internal/config"
	"github.com/mchl18/certcheckbot/internal/logger"
	"github.com/mchl18/certcheckbot/internal/server"
)

func main() {
	// Parse command line flags
	configureFlag := flag.Bool("configure", false, "Run the configuration setup")
	flag.Parse()

	if *configureFlag {
		if err := config.RunSetup(); err != nil {
			fmt.Printf("Failed to run setup: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration completed successfully!")
		return
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	// Create necessary directories
	certCheckerDir := filepath.Join(homeDir, ".certchecker")
	configDir := filepath.Join(certCheckerDir, "config")
	logsDir := filepath.Join(certCheckerDir, "logs")
	dataDir := filepath.Join(certCheckerDir, "data")

	for _, dir := range []string{certCheckerDir, configDir, logsDir, dataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Initialize logger
	logger := logger.New(homeDir)

	// Load configuration
	cfg, err := config.Load(homeDir)
	if err != nil {
		logger.Error("Failed to load configuration", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Initialize certificate checker
	certChecker := checker.New(cfg.Domains, cfg.ThresholdDays, cfg.SlackWebhookURL, logger, dataDir)

	// Initialize HTTP server if enabled
	if cfg.HTTPEnabled {
		server := server.New(certChecker, cfg.HTTPAuthToken, homeDir)
		go func() {
			logger.Info("Starting HTTP server", map[string]interface{}{"port": cfg.HTTPPort})
			if err := server.Start(cfg.HTTPPort); err != nil {
				logger.Error("HTTP server failed", map[string]interface{}{"error": err.Error()})
			}
		}()
	}

	// Start heartbeat if enabled
	if cfg.HeartbeatHours > 0 {
		heartbeatInterval := time.Duration(cfg.HeartbeatHours) * time.Hour
		logger.Info("Heartbeat enabled", map[string]interface{}{"interval": heartbeatInterval.String()})
		go func() {
			ticker := time.NewTicker(heartbeatInterval)
			defer ticker.Stop()

			for range ticker.C {
				if err := certChecker.SendHeartbeat(); err != nil {
					logger.Error("Failed to send heartbeat", map[string]interface{}{"error": err.Error()})
				}
			}
		}()
	}

	// Start certificate checking loop
	checkInterval := time.Duration(cfg.IntervalHours) * time.Hour
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Initial check
	if err := certChecker.CheckCertificates(); err != nil {
		logger.Error("Failed to check certificates", map[string]interface{}{"error": err.Error()})
	}

	// Continuous checking
	for range ticker.C {
		if err := certChecker.CheckCertificates(); err != nil {
			logger.Error("Failed to check certificates", map[string]interface{}{"error": err.Error()})
		}
	}
}
