package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryManager(t *testing.T) {
	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "history-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize history manager
	manager := NewHistoryManager(tempDir)

	// Test domain and threshold
	domain := "example.com"
	threshold := 30
	expiryDate := time.Now().Add(30 * 24 * time.Hour)

	// Test recording an alert
	if err := manager.RecordAlertForThreshold(domain, threshold, expiryDate); err != nil {
		t.Errorf("Failed to record alert: %v", err)
	}

	// Test checking if alerted
	if !manager.HasAlertedForThreshold(domain, threshold, expiryDate) {
		t.Error("Expected HasAlertedForThreshold to return true")
	}

	// Test checking if alerted with different expiry date
	differentDate := expiryDate.Add(24 * time.Hour)
	if manager.HasAlertedForThreshold(domain, threshold, differentDate) {
		t.Error("Expected HasAlertedForThreshold to return false for different expiry date")
	}

	// Test checking if alerted with different threshold
	differentThreshold := 14
	if manager.HasAlertedForThreshold(domain, differentThreshold, expiryDate) {
		t.Error("Expected HasAlertedForThreshold to return false for different threshold")
	}

	// Test checking if alerted with different domain
	differentDomain := "test.com"
	if manager.HasAlertedForThreshold(differentDomain, threshold, expiryDate) {
		t.Error("Expected HasAlertedForThreshold to return false for different domain")
	}
}

func TestHistoryManagerBackup(t *testing.T) {
	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "history-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize history manager
	manager := NewHistoryManager(tempDir)

	// Test domain and threshold
	domain := "example.com"
	threshold := 30
	expiryDate := time.Now().Add(30 * 24 * time.Hour)

	// Record an alert
	if err := manager.RecordAlertForThreshold(domain, threshold, expiryDate); err != nil {
		t.Errorf("Failed to record alert: %v", err)
	}

	// Check if backup file was created
	historyPath := filepath.Join(tempDir, "alert-history.json")
	backupPath := historyPath + ".backup"

	// Record another alert to trigger backup
	if err := manager.RecordAlertForThreshold(domain, 14, expiryDate); err != nil {
		t.Errorf("Failed to record second alert: %v", err)
	}

	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Expected backup file to exist")
	}
}
