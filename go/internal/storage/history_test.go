package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "history-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logs/data directory structure
	dataDir := filepath.Join(tempDir, "logs", "data")
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	t.Logf("Using temp dir: %s", tempDir)
	t.Logf("Using data dir: %s", dataDir)

	manager := NewHistoryManager(tempDir)

	// Test saving new history
	history := map[string]map[int]string{
		"example.com": {
			30: "2024-01-10",
			14: "2024-01-26",
		},
		"test.com": {
			7: "2024-02-03",
		},
	}

	// First save to create the initial file
	if err := manager.SaveHistory(history); err != nil {
		t.Errorf("First SaveHistory() error = %v", err)
	}

	// Give filesystem a moment to complete write operations
	time.Sleep(100 * time.Millisecond)

	// Save again to trigger backup creation
	if err := manager.SaveHistory(history); err != nil {
		t.Errorf("Second SaveHistory() error = %v", err)
	}

	// Give filesystem another moment
	time.Sleep(100 * time.Millisecond)

	// Test loading saved history
	loaded, err := manager.LoadHistory()
	if err != nil {
		t.Errorf("LoadHistory() error = %v", err)
	}

	// Verify loaded history matches saved history
	for domain, thresholds := range history {
		loadedThresholds, ok := loaded[domain]
		if !ok {
			t.Errorf("LoadHistory() missing domain %s", domain)
			continue
		}

		for threshold, date := range thresholds {
			loadedDate, ok := loadedThresholds[threshold]
			if !ok {
				t.Errorf("LoadHistory() missing threshold %d for domain %s", threshold, domain)
				continue
			}

			if loadedDate != date {
				t.Errorf("LoadHistory() for domain %s threshold %d = %v, want %v",
					domain, threshold, loadedDate, date)
			}
		}
	}

	// Test backup file creation
	mainFile := filepath.Join(dataDir, "alert-history.json")
	backupFile := mainFile + ".backup"

	// Check if main file exists
	if _, err := os.Stat(mainFile); err != nil {
		t.Errorf("Main file does not exist: %v", err)
	} else {
		t.Log("Main file exists")
	}

	// Check backup file
	if _, err := os.Stat(backupFile); err != nil {
		t.Errorf("Backup file does not exist: %v", err)
	} else {
		t.Log("Backup file exists")
	}

	// Test loading with no existing file
	os.Remove(mainFile)
	os.Remove(backupFile)

	emptyHistory, err := manager.LoadHistory()
	if err != nil {
		t.Errorf("LoadHistory() with no file error = %v", err)
	}

	if len(emptyHistory) != 0 {
		t.Errorf("LoadHistory() with no file should return empty map, got %v", emptyHistory)
	}
}
