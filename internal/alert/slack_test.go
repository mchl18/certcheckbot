package alert

import (
	"testing"
	"time"
)

func TestSendAlert(t *testing.T) {
	tests := []struct {
		name           string
		domain         string
		daysRemaining  int
		expirationDate time.Time
		threshold      int
		wantErr        bool
	}{
		{
			name:           "normal expiration",
			domain:         "example.com",
			daysRemaining:  25,
			expirationDate: time.Now().Add(25 * 24 * time.Hour),
			threshold:      30,
			wantErr:        true, // Will error because webhook URL is not valid
		},
		{
			name:           "urgent expiration",
			domain:         "test.com",
			daysRemaining:  5,
			expirationDate: time.Now().Add(5 * 24 * time.Hour),
			threshold:      7,
			wantErr:        true, // Will error because webhook URL is not valid
		},
	}

	notifier := NewSlackNotifier("mock-webhook")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := notifier.SendAlert(tt.domain, tt.daysRemaining, tt.expirationDate, tt.threshold)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendAlert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
