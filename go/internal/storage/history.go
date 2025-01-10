package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type HistoryManager struct {
	historyPath string
}

// History is a map of domain names to their alert history
// The inner map is threshold days to last alert date
type History map[string]map[int]string

func NewHistoryManager(projectRoot string) *HistoryManager {
	return &HistoryManager{
		historyPath: filepath.Join(projectRoot, "logs", "data", "alert-history.json"),
	}
}

func (h *HistoryManager) LoadHistory() (History, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(h.historyPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	data, err := os.ReadFile(h.historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(History), nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return history, nil
}

func (h *HistoryManager) SaveHistory(history History) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(h.historyPath), 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Create backup if file exists
	if _, err := os.Stat(h.historyPath); err == nil {
		data, err := os.ReadFile(h.historyPath)
		if err != nil {
			return fmt.Errorf("failed to read existing history file for backup: %w", err)
		}

		backupPath := h.historyPath + ".backup"
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write history backup file: %w", err)
		}
	}

	// Marshal and save new history
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(h.historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
} 