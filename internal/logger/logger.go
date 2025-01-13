package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var startTime = time.Now()

type Logger struct {
	logFile string
}

func New(homeDir string) *Logger {
	// Ensure log directory exists within .certchecker
	logDir := filepath.Join(homeDir, ".certchecker", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	// Set log file path within .certchecker
	logPath := filepath.Join(logDir, "cert-checker.log")

	// Create or truncate the log file
	if err := os.WriteFile(logPath, []byte(""), 0644); err != nil {
		panic(fmt.Sprintf("Failed to create log file: %v", err))
	}

	return &Logger{logFile: logPath}
}

func (l *Logger) log(level string, message string, details map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	processInfo := map[string]string{
		"pid":    fmt.Sprintf("%d", os.Getpid()),
		"memory": fmt.Sprintf("%dMB", memStats.HeapAlloc/1024/1024),
		"uptime": fmt.Sprintf("%ds", int(time.Since(startTime).Seconds())),
	}

	logMessage := fmt.Sprintf("[%s] [%s] [PID:%s] [MEM:%s] [UPTIME:%s] %s\n",
		timestamp,
		level,
		processInfo["pid"],
		processInfo["memory"],
		processInfo["uptime"],
		message,
	)

	if details != nil {
		detailsJSON, err := json.MarshalIndent(details, "", "  ")
		if err == nil {
			logMessage += string(detailsJSON) + "\n"
		}
	}

	// Write to log file
	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(logMessage); err != nil {
		fmt.Printf("Failed to write to log file: %v\n", err)
		return
	}

	// Also log to console
	fmt.Print(logMessage)
}

func (l *Logger) Info(message string, details map[string]interface{}) {
	l.log("INFO", message, details)
}

func (l *Logger) Error(message string, details map[string]interface{}) {
	l.log("ERROR", message, details)
}

func (l *Logger) Warning(message string, details map[string]interface{}) {
	l.log("WARNING", message, details)
}
