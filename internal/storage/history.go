package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type HistoryManager struct {
	dataDir string
}

type AlertHistory struct {
	Alerts map[string]map[int]time.Time `json:"alerts"` // domain -> threshold -> last alert time
}

func NewHistoryManager(dataDir string) *HistoryManager {
	return &HistoryManager{
		dataDir: dataDir,
	}
}

func (h *HistoryManager) HasAlertedForThreshold(domain string, threshold int, expiryDate time.Time) bool {
	history, err := h.loadHistory()
	if err != nil {
		return false
	}

	if alerts, ok := history.Alerts[domain]; ok {
		if lastAlert, ok := alerts[threshold]; ok {
			// Check if we've already alerted for this expiry date
			return lastAlert.Equal(expiryDate)
		}
	}

	return false
}

func (h *HistoryManager) RecordAlertForThreshold(domain string, threshold int, expiryDate time.Time) error {
	history, err := h.loadHistory()
	if err != nil {
		history = &AlertHistory{
			Alerts: make(map[string]map[int]time.Time),
		}
	}

	// Initialize domain map if it doesn't exist
	if _, ok := history.Alerts[domain]; !ok {
		history.Alerts[domain] = make(map[int]time.Time)
	}

	// Record the alert
	history.Alerts[domain][threshold] = expiryDate

	// Save the updated history
	return h.saveHistory(history)
}

func (h *HistoryManager) loadHistory() (*AlertHistory, error) {
	historyPath := h.getHistoryPath()

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(historyPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Read history file
	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AlertHistory{
				Alerts: make(map[string]map[int]time.Time),
			}, nil
		}
		return nil, fmt.Errorf("failed to read history file: %v", err)
	}

	// Parse history
	var history AlertHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %v", err)
	}

	return &history, nil
}

func (h *HistoryManager) saveHistory(history *AlertHistory) error {
	historyPath := h.getHistoryPath()

	// Create backup of existing file
	if _, err := os.Stat(historyPath); err == nil {
		backupPath := historyPath + ".backup"
		if err := os.Rename(historyPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
	}

	// Marshal history to JSON
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %v", err)
	}

	// Write to file
	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %v", err)
	}

	return nil
}

func (h *HistoryManager) getHistoryPath() string {
	return filepath.Join(h.dataDir, "alert-history.json")
}
