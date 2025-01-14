package checker

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mchl18/ssl-expiration-check-bot/internal/logger"
	"github.com/mchl18/ssl-expiration-check-bot/internal/storage"
)

// Make getCertificate a variable so it can be mocked in tests
var getCertificate = func(domain string) (*tls.Certificate, error) {
	conn, err := tls.Dial("tcp", domain+":443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	return &tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		Leaf:       cert,
	}, nil
}

type CertificateChecker struct {
	domains       []string
	thresholds   []int
	webhookURL   string
	logger       *logger.Logger
	history      *storage.HistoryManager
}

func New(domains []string, thresholds []int, webhookURL string, logger *logger.Logger, dataDir string) *CertificateChecker {
	return &CertificateChecker{
		domains:     domains,
		thresholds: thresholds,
		webhookURL: webhookURL,
		logger:     logger,
		history:    storage.NewHistoryManager(dataDir),
	}
}

func (c *CertificateChecker) GetDomains() []string {
	return c.domains
}

func (c *CertificateChecker) GetThresholds() []int {
	return c.thresholds
}

func (c *CertificateChecker) CheckCertificates() error {
	c.logger.Info("Starting certificate check", map[string]interface{}{
		"domains": c.domains,
	})

	for _, domain := range c.domains {
		cert, err := getCertificate(domain)
		if err != nil {
			c.logger.Error("Failed to get certificate", map[string]interface{}{
				"domain": domain,
				"error":  err.Error(),
			})
			continue
		}

		daysUntilExpiry := int(time.Until(cert.Leaf.NotAfter).Hours() / 24)
		c.logger.Info("Certificate expiration check", map[string]interface{}{
			"domain":        domain,
			"daysRemaining": daysUntilExpiry,
		})

		// Check if we need to send alerts
		for _, threshold := range c.thresholds {
			if daysUntilExpiry <= threshold {
				// Check if we've already alerted for this threshold
				if !c.history.HasAlertedForThreshold(domain, threshold, cert.Leaf.NotAfter) {
					message := fmt.Sprintf("SSL Certificate for %s will expire in %d days (on %s)",
						domain, daysUntilExpiry, cert.Leaf.NotAfter.Format("2006-01-02"))

					if err := c.sendSlackNotification(message); err != nil {
						c.logger.Error("Failed to send Slack notification", map[string]interface{}{
							"domain": domain,
							"error":  err.Error(),
						})
						continue
					}

					// Record the alert in history
					if err := c.history.RecordAlertForThreshold(domain, threshold, cert.Leaf.NotAfter); err != nil {
						c.logger.Error("Failed to record alert", map[string]interface{}{
							"domain": domain,
							"error":  err.Error(),
						})
					}

					c.logger.Info("Alert sent", map[string]interface{}{
						"domain":    domain,
						"threshold": threshold,
					})
				}
			}
		}
	}

	return nil
}

func (c *CertificateChecker) SendHeartbeat() error {
	message := fmt.Sprintf("SSL Certificate Checker is running\nMonitoring domains: %v\nThresholds: %v days",
		c.domains, c.thresholds)

	if err := c.sendSlackNotification(message); err != nil {
		return fmt.Errorf("failed to send heartbeat: %v", err)
	}

	c.logger.Info("Heartbeat sent", map[string]interface{}{
		"domains":    c.domains,
		"thresholds": c.thresholds,
	})
	return nil
}

func (c *CertificateChecker) sendSlackNotification(message string) error {
	payload := map[string]string{
		"text": message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(c.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Start begins the certificate checking loop with the specified interval
func (c *CertificateChecker) Start(intervalHours int) {
	checkInterval := 6 * time.Hour // default interval
	if intervalHours > 0 {
		checkInterval = time.Duration(intervalHours) * time.Hour
	}

	c.logger.Info("Starting certificate checker", map[string]interface{}{
		"check_interval": checkInterval.String(),
		"domains":        c.domains,
		"thresholds":     c.thresholds,
	})

	// Initial check
	if err := c.CheckCertificates(); err != nil {
		c.logger.Error("Certificate check failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Start periodic checks
	ticker := time.NewTicker(checkInterval)
	for range ticker.C {
		if err := c.CheckCertificates(); err != nil {
			c.logger.Error("Certificate check failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
}

// StartHeartbeat begins sending periodic heartbeat messages
func (c *CertificateChecker) StartHeartbeat(intervalHours int) {
	heartbeatInterval := time.Duration(intervalHours) * time.Hour
	ticker := time.NewTicker(heartbeatInterval)

	// Initial heartbeat
	if err := c.SendHeartbeat(); err != nil {
		c.logger.Error("Failed to send heartbeat", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Start periodic heartbeats
	for range ticker.C {
		if err := c.SendHeartbeat(); err != nil {
			c.logger.Error("Failed to send heartbeat", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
}
