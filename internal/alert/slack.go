package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SlackNotifier struct {
	webhookURL string
}

type slackMessage struct {
	Text string `json:"text"`
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{webhookURL: webhookURL}
}

func (s *SlackNotifier) SendAlert(domain string, daysToExpiration int, expirationDate time.Time, threshold int) error {
	message := slackMessage{
		Text: fmt.Sprintf("🚨 *SSL Certificate Expiration Alert*\nThe SSL certificate for *%s* will expire in *%d* days (%s).\nThreshold reached: %d days\nPlease take action to renew the certificate before it expires.",
			domain,
			daysToExpiration,
			expirationDate.Format(time.RFC3339),
			threshold,
		),
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	resp, err := http.Post(s.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

func (n *SlackNotifier) SendMessage(message string, details map[string]interface{}) error {
	// Convert details to a formatted string
	var detailsStr string
	if details != nil {
		detailsBytes, err := json.MarshalIndent(details, "", "  ")
		if err == nil {
			detailsStr = "\n```" + string(detailsBytes) + "```"
		}
	}

	// Create payload
	payload := map[string]interface{}{
		"text": message + detailsStr,
	}

	// Marshal payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send request
	resp, err := http.Post(n.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
