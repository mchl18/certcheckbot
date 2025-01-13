package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mchl18/ssl-expiration-check-bot/internal/checker"
	"github.com/mchl18/ssl-expiration-check-bot/internal/config"
	"github.com/mchl18/ssl-expiration-check-bot/internal/logger"
	"github.com/mchl18/ssl-expiration-check-bot/internal/server"
	"github.com/mchl18/ssl-expiration-check-bot/internal/webui"
)

func promptForConfigMethod() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nWould you like to configure via web UI or command line? [web/cli]: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read input: %v", err)
		}
		
		choice = strings.TrimSpace(strings.ToLower(choice))
		switch choice {
		case "web", "webui", "w":
			return true, nil
		case "cli", "cmd", "c":
			return false, nil
		default:
			fmt.Println("Please enter 'web' or 'cli'")
		}
	}
}

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
		fmt.Printf("Configuration error: %v\n", err)
		
		// If webui flag is not explicitly set, prompt for method
		if !*webUIFlag {
			useWebUI, err := promptForConfigMethod()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			*webUIFlag = useWebUI
			*configureFlag = !useWebUI
		}

		if *configureFlag {
			if err := config.RunSetup(); err != nil {
				fmt.Printf("Failed to run setup: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Configuration completed successfully!")
			fmt.Println("Please restart the application to apply the configuration.")
			return
		}

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
