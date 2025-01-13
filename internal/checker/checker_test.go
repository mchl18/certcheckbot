package checker

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/mchl18/ssl-expiration-check-bot/internal/logger"
)

// Mock certificate for testing
func createMockCertificate(notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
}

// Override getCertificate for testing
func mockGetCertificate(domain string) (*tls.Certificate, error) {
	// Create a mock certificate that will expire in 30 days
	cert := createMockCertificate(time.Now().Add(30 * 24 * time.Hour))
	return &tls.Certificate{
		Leaf: cert,
	}, nil
}

func TestChecker(t *testing.T) {
	// Save the original getCertificate function and restore it after the test
	originalGetCertificate := getCertificate
	getCertificate = mockGetCertificate
	defer func() { getCertificate = originalGetCertificate }()

	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "checker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	t.Setenv("HOME", tempDir)
	logger := logger.New("cert-checker.log")

	// Test configuration
	domains := []string{"example.com", "test.com"}
	thresholds := []int{7, 14, 30}
	slackWebhookURL := "https://hooks.slack.com/services/test"

	// Initialize checker
	checker := New(domains, thresholds, slackWebhookURL, logger, tempDir)

	// Test GetDomains
	gotDomains := checker.GetDomains()
	if len(gotDomains) != len(domains) {
		t.Errorf("GetDomains() = %v, want %v", gotDomains, domains)
	} else {
		for i := range domains {
			if gotDomains[i] != domains[i] {
				t.Errorf("GetDomains()[%d] = %v, want %v", i, gotDomains[i], domains[i])
			}
		}
	}

	// Test GetThresholds
	gotThresholds := checker.GetThresholds()
	if len(gotThresholds) != len(thresholds) {
		t.Errorf("GetThresholds() = %v, want %v", gotThresholds, thresholds)
	} else {
		for i := range thresholds {
			if gotThresholds[i] != thresholds[i] {
				t.Errorf("GetThresholds()[%d] = %v, want %v", i, gotThresholds[i], thresholds[i])
			}
		}
	}

	// Test CheckCertificates
	if err := checker.CheckCertificates(); err != nil {
		t.Errorf("CheckCertificates() error = %v", err)
	}

	// Test SendHeartbeat
	if err := checker.SendHeartbeat(); err != nil {
		t.Errorf("SendHeartbeat() error = %v", err)
	}
}
