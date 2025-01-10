package checker

import (
	"os"
	"testing"

	"github.com/mchl18/certcheckbot/internal/logger"
)

func TestCheckCertificate(t *testing.T) {
	// Create a temporary directory for test config
	tempDir, err := os.MkdirTemp("", "checker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	t.Setenv("HOME", tempDir)
	logger := logger.New("cert-checker.log")

	// Test domains and thresholds
	domains := []string{"google.com"}
	thresholds := []int{7, 14, 30, 45}
	slackWebhookURL := "https://hooks.slack.com/services/test"

	// Initialize checker
	checker := New(domains, thresholds, slackWebhookURL, logger, "")

	// Run check
	if err := checker.CheckAll(); err != nil {
		t.Errorf("CheckAll() error = %v", err)
	}
}
