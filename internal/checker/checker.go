package checker

import (
	"crypto/tls"
	"fmt"
	"sort"
	"time"

	"github.com/mchl18/certcheckbot/internal/alert"
	"github.com/mchl18/certcheckbot/internal/logger"
	"github.com/mchl18/certcheckbot/internal/storage"
)

type CertificateChecker struct {
	domains       []string
	thresholdDays []int
	logger        *logger.Logger
	slackNotifier *alert.SlackNotifier
	history       *storage.HistoryManager
}

func New(domains []string, thresholdDays []int, slackWebhookURL string, logger *logger.Logger, projectRoot string) *CertificateChecker {
	return &CertificateChecker{
		domains:       domains,
		thresholdDays: thresholdDays,
		logger:        logger,
		slackNotifier: alert.NewSlackNotifier(slackWebhookURL),
		history:       storage.NewHistoryManager(projectRoot),
	}
}

func (c *CertificateChecker) Domains() []string {
	return c.domains
}

func (c *CertificateChecker) ThresholdDays() []int {
	return c.thresholdDays
}

func (c *CertificateChecker) CheckAll() error {
	for _, domain := range c.domains {
		if err := c.checkCertificate(domain); err != nil {
			return fmt.Errorf("failed to check certificate for %s: %w", domain, err)
		}
	}
	return nil
}

func (c *CertificateChecker) checkCertificate(domain string) error {
	c.logger.Info(fmt.Sprintf("Starting SSL certificate check for domain: %s", domain), map[string]interface{}{
		"checkTime": time.Now().Format(time.RFC3339),
		"domain":    domain,
		"checkType": "SSL_CERTIFICATE",
	})

	// Connect to the domain
	conn, err := tls.Dial("tcp", domain+":443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		c.logger.Error(fmt.Sprintf("Failed to check certificate for %s", domain), map[string]interface{}{
			"domain":       domain,
			"errorMessage": err.Error(),
		})
		return err
	}
	defer conn.Close()

	// Get the certificate
	cert := conn.ConnectionState().PeerCertificates[0]
	expirationDate := cert.NotAfter

	c.logger.Info(fmt.Sprintf("Certificate details retrieved for %s", domain), map[string]interface{}{
		"domain":       domain,
		"issuer":       cert.Issuer,
		"subject":      cert.Subject,
		"validFrom":    cert.NotBefore.Format(time.RFC3339),
		"validTo":      expirationDate.Format(time.RFC3339),
		"serialNumber": cert.SerialNumber.String(),
		"fingerprint":  fmt.Sprintf("%x", cert.Signature),
		"protocol":     conn.ConnectionState().Version,
	})

	daysToExpiration := int(time.Until(expirationDate).Hours() / 24)

	c.logger.Info(fmt.Sprintf("Certificate expiration analysis for %s", domain), map[string]interface{}{
		"domain":         domain,
		"daysRemaining":  daysToExpiration,
		"expirationDate": expirationDate.Format(time.RFC3339),
		"status":         map[bool]string{true: "WARNING", false: "OK"}[daysToExpiration <= 30],
	})

	return c.checkAndSendAlert(domain, daysToExpiration, expirationDate)
}

func (c *CertificateChecker) checkAndSendAlert(domain string, daysToExpiration int, expirationDate time.Time) error {
	history, err := c.history.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	today := time.Now().Format("2006-01-02")

	if _, exists := history[domain]; !exists {
		history[domain] = make(map[int]string)
	}

	// Sort thresholds in ascending order
	thresholds := make([]int, len(c.thresholdDays))
	copy(thresholds, c.thresholdDays)
	sort.Ints(thresholds)

	var thresholdToAlert *int
	for _, threshold := range thresholds {
		if daysToExpiration <= threshold {
			thresholdToAlert = &threshold
			break
		}
	}

	if thresholdToAlert != nil {
		lastAlert := history[domain][*thresholdToAlert]
		if lastAlert != today {
			if err := c.slackNotifier.SendAlert(domain, daysToExpiration, expirationDate, *thresholdToAlert); err != nil {
				return fmt.Errorf("failed to send alert: %w", err)
			}
			history[domain][*thresholdToAlert] = today
			if err := c.history.SaveHistory(history); err != nil {
				return fmt.Errorf("failed to save history: %w", err)
			}
		}
	}

	return nil
}
