package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mchl18/ssl-expiration-check-bot/internal/checker"
	"github.com/mchl18/ssl-expiration-check-bot/internal/config"
	"github.com/mchl18/ssl-expiration-check-bot/internal/logger"
	"github.com/mchl18/ssl-expiration-check-bot/internal/server"
	"github.com/mchl18/ssl-expiration-check-bot/internal/webui"
)

func main() {
	// Parse command line flags
	configureFlag := flag.Bool("configure", false, "Run the configuration setup")
	webUIFlag := flag.Bool("webui", false, "Start the web UI")
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
	for _, dir := range []string{"config", "logs", "data"} {
		path := filepath.Join(certCheckerDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", path, err)
			os.Exit(1)
		}
	}

	// Initialize logger
	logger := logger.New(homeDir)

	// Load configuration
	cfg, err := config.Load(homeDir)
	if err != nil {
		if *webUIFlag {
			// Start web UI for initial configuration
			webUI, err := webui.New(homeDir, logger)
			if err != nil {
				logger.Error("Failed to initialize web UI", map[string]interface{}{
					"error": err.Error(),
				})
				os.Exit(1)
			}
			fmt.Println("\nWeb UI is available at: http://localhost:8081")
			fmt.Println("Use this interface to configure the service and view logs.")
			if err := webUI.Start(); err != nil {
				logger.Error("Web UI failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
			select {}
		} else {
			// Only show configuration message if no config exists and no webui flag
			fmt.Printf("Configuration error: %v\n", err)
			fmt.Println("Run with -configure flag for CLI setup or -webui flag for web-based setup.")
			os.Exit(1)
		}
	}

	// Start web UI if enabled
	if *webUIFlag {
		webUI, err := webui.New(homeDir, logger)
		if err != nil {
			logger.Error("Failed to initialize web UI", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
		go func() {
			fmt.Println("\nWeb UI is available at: http://localhost:8081")
			fmt.Println("Use this interface to configure the service and view logs.")
			if err := webUI.Start(); err != nil {
				logger.Error("Web UI failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}()
	}

	// Initialize certificate checker
	certChecker := checker.New(cfg.Domains, cfg.ThresholdDays, cfg.SlackWebhookURL, logger, filepath.Join(certCheckerDir, "data"))

	// Start HTTP server if enabled
	if cfg.HTTPEnabled {
		srv := server.New(certChecker, cfg.HTTPAuthToken, homeDir)
		go func() {
			logger.Info("Starting HTTP server", map[string]interface{}{
				"port": cfg.HTTPPort,
			})
			if err := srv.Start(cfg.HTTPPort); err != nil {
				logger.Error("HTTP server failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}()
	}

	// Start heartbeat if enabled
	var heartbeatTicker *time.Ticker
	if cfg.HeartbeatHours > 0 {
		heartbeatInterval := time.Duration(cfg.HeartbeatHours) * time.Hour
		heartbeatTicker = time.NewTicker(heartbeatInterval)
		go func() {
			for range heartbeatTicker.C {
				if err := certChecker.SendHeartbeat(); err != nil {
					logger.Error("Failed to send heartbeat", map[string]interface{}{
						"error": err.Error(),
					})
				}
			}
		}()
		logger.Info("Heartbeat enabled", map[string]interface{}{
			"interval": heartbeatInterval.String(),
		})
	}

	// Start certificate check loop
	checkInterval := 6 * time.Hour
	if cfg.IntervalHours > 0 {
		checkInterval = time.Duration(cfg.IntervalHours) * time.Hour
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	if heartbeatTicker != nil {
		defer heartbeatTicker.Stop()
	}

	logger.Info("Starting certificate checker", map[string]interface{}{
		"check_interval": checkInterval.String(),
		"domains":        cfg.Domains,
		"thresholds":     cfg.ThresholdDays,
	})

	// Initial check
	if err := certChecker.CheckCertificates(); err != nil {
		logger.Error("Certificate check failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Main loop
	for range ticker.C {
		if err := certChecker.CheckCertificates(); err != nil {
			logger.Error("Certificate check failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
}
