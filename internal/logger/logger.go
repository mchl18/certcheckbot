package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Logger struct {
	logFile string
}

func New(logFile string) *Logger {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get home directory: %v", err))
	}

	// Ensure log directory exists within .certchecker
	logDir := filepath.Join(home, ".certchecker", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	// Set log file path within .certchecker
	logPath := filepath.Join(logDir, logFile)
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
	if err == nil {
		defer f.Close()
		f.WriteString(logMessage)
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

var startTime = time.Now()
