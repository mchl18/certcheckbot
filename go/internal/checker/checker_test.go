package checker

import (
	"testing"
	"time"

	"github.com/mchl18/certcheckbot/internal/alert"
	"github.com/mchl18/certcheckbot/internal/logger"
)

type mockNotifier struct {
	alerts []struct {
		domain          string
		daysRemaining  int
		expirationDate time.Time
		threshold      int
	}
}

func (m *mockNotifier) SendAlert(domain string, daysRemaining int, expirationDate time.Time, threshold int) error {
	m.alerts = append(m.alerts, struct {
		domain          string
		daysRemaining  int
		expirationDate time.Time
		threshold      int
	}{
		domain:          domain,
		daysRemaining:  daysRemaining,
		expirationDate: expirationDate,
		threshold:      threshold,
	})
	return nil
}

func TestNew(t *testing.T) {
	domains := []string{"example.com"}
	thresholds := []int{7, 14, 30}
	logger := logger.New("/dev/null") // Discard logs during tests

	checker := New(domains, thresholds, "mock-webhook", logger, ".")

	if len(checker.domains) != len(domains) {
		t.Errorf("Expected %d domains, got %d", len(domains), len(checker.domains))
	}

	if len(checker.thresholdDays) != len(thresholds) {
		t.Errorf("Expected %d thresholds, got %d", len(thresholds), len(checker.thresholdDays))
	}
}

func TestDomains(t *testing.T) {
	domains := []string{"example.com", "test.com"}
	checker := &CertificateChecker{domains: domains}

	result := checker.Domains()
	if len(result) != len(domains) {
		t.Errorf("Expected %d domains, got %d", len(domains), len(result))
	}

	for i, domain := range domains {
		if result[i] != domain {
			t.Errorf("Expected domain %s at position %d, got %s", domain, i, result[i])
		}
	}
}

func TestThresholdDays(t *testing.T) {
	thresholds := []int{7, 14, 30}
	checker := &CertificateChecker{thresholdDays: thresholds}

	result := checker.ThresholdDays()
	if len(result) != len(thresholds) {
		t.Errorf("Expected %d thresholds, got %d", len(thresholds), len(result))
	}

	for i, threshold := range thresholds {
		if result[i] != threshold {
			t.Errorf("Expected threshold %d at position %d, got %d", threshold, i, result[i])
		}
	}
}

// TestCheckCertificate tests the certificate checking functionality
// Note: This is an integration test that requires internet connectivity
func TestCheckCertificate(t *testing.T) {
	// Skip in CI environment or add flag to explicitly run integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	domains := []string{"google.com"} // Using a reliable domain for testing
	thresholds := []int{365} // Set high threshold to ensure we get an alert
	logger := logger.New("/dev/null")
	mockNotifier := &mockNotifier{}

	checker := &CertificateChecker{
		domains:       domains,
		thresholdDays: thresholds,
		logger:        logger,
		slackNotifier: alert.NewSlackNotifier("mock-webhook"), // Use real notifier
		history:       nil, // We don't need history for this test
	}

	err := checker.CheckAll()
	if err != nil {
		t.Errorf("CheckAll failed: %v", err)
	}

	// Verify that the certificate check completed without error
	if len(mockNotifier.alerts) == 0 {
		t.Error("Expected at least one alert for the test domain")
	}
} 