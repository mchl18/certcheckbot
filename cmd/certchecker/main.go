package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	// Create certchecker directory if it doesn't exist
	certCheckerDir := filepath.Join(homeDir, ".certchecker")
	if err := os.MkdirAll(certCheckerDir, 0755); err != nil {
		fmt.Printf("Failed to create certchecker directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logger.New(certCheckerDir)

	// Handle configuration
	if *configureFlag {
		if err := config.RunSetup(); err != nil {
			fmt.Printf("Failed to run setup: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration completed successfully!")
		fmt.Println("Please restart the application to apply the configuration.")
		return
	}

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
			return
		}

		// Prompt for configuration method
		useWebUI, err := promptForConfigMethod()
		if err != nil {
			fmt.Printf("Failed to get configuration method: %v\n", err)
			os.Exit(1)
		}

		if useWebUI {
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
			return
		}

		if err := config.RunSetup(); err != nil {
			fmt.Printf("Failed to run setup: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration completed successfully!")
		fmt.Println("Please restart the application to apply the configuration.")
		return
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

		// Start web UI in a goroutine
		go func() {
			fmt.Println("\nWeb UI is available at: http://localhost:8081")
			fmt.Println("Use this interface to configure the service and view logs.")
			if err := webUI.Start(); err != nil && err != http.ErrServerClosed {
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
			if err := srv.Start(cfg.HTTPPort); err != nil && err != http.ErrServerClosed {
				logger.Error("HTTP server failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}()
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start the certificate checker
	go certChecker.Start(cfg.IntervalHours)

	// Start heartbeat if enabled
	if cfg.HeartbeatHours > 0 {
		logger.Info("Heartbeat enabled", map[string]interface{}{
			"interval": time.Duration(cfg.HeartbeatHours) * time.Hour,
		})
		go certChecker.StartHeartbeat(cfg.HeartbeatHours)
	}

	// Wait for signal
	<-sigChan
}
