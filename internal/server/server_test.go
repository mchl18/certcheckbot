package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mchl18/ssl-expiration-check-bot/internal/checker"
	"github.com/mchl18/ssl-expiration-check-bot/internal/logger"
)

func TestServer(t *testing.T) {
	// Create a temporary directory for test config
	tempDir, err := os.MkdirTemp("", "server-test")
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
	authToken := "test-token"

	// Initialize checker
	checker := checker.New(domains, thresholds, slackWebhookURL, logger, "")

	// Initialize server
	server := New(checker, authToken, tempDir)

	// Create some test logs
	logDir := filepath.Join(tempDir, ".certchecker", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log dir: %v", err)
	}
	logPath := filepath.Join(logDir, "cert-checker.log")
	testLogs := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(logPath, []byte(testLogs), 0644); err != nil {
		t.Fatalf("Failed to write test logs: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		token          string
		wantStatus     int
		wantFields     []string
		wantFieldTypes map[string]string
	}{
		{
			name:       "health check with valid token",
			path:       "/health",
			token:      authToken,
			wantStatus: http.StatusOK,
			wantFields: []string{"status", "uptime", "domains", "thresholds", "started_at", "checked_at", "version"},
			wantFieldTypes: map[string]string{
				"status":     "string",
				"uptime":     "string",
				"domains":    "[]string",
				"thresholds": "[]int",
				"started_at": "string",
				"checked_at": "string",
				"version":    "string",
			},
		},
		{
			name:       "health check with invalid token",
			path:       "/health",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "logs with valid token",
			path:       "/logs",
			token:      authToken,
			wantStatus: http.StatusOK,
			wantFields: []string{"lines", "total", "logs", "timestamp"},
			wantFieldTypes: map[string]string{
				"lines":     "float64", // JSON numbers are float64 in Go
				"total":     "float64",
				"logs":      "[]interface{}",
				"timestamp": "string",
			},
		},
		{
			name:       "logs with invalid token",
			path:       "/logs",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "logs with line limit",
			path:       "/logs?lines=2",
			token:      authToken,
			wantStatus: http.StatusOK,
			wantFields: []string{"lines", "total", "logs", "timestamp"},
		},
		{
			name:       "logs with invalid line limit",
			path:       "/logs?lines=invalid",
			token:      authToken,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rr := httptest.NewRecorder()
			mux := http.NewServeMux()
			mux.HandleFunc("/health", server.authMiddleware(server.handleHealth))
			mux.HandleFunc("/logs", server.authMiddleware(server.handleLogs))
			mux.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Check required fields
				for _, field := range tt.wantFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Response missing field %q", field)
					}
				}

				// Check field types
				for field, wantType := range tt.wantFieldTypes {
					value := response[field]
					switch wantType {
					case "string":
						if _, ok := value.(string); !ok {
							t.Errorf("Field %q should be string, got %T", field, value)
						}
					case "[]string":
						if arr, ok := value.([]interface{}); !ok {
							t.Errorf("Field %q should be array, got %T", field, value)
						} else {
							for _, v := range arr {
								if _, ok := v.(string); !ok {
									t.Errorf("Field %q array element should be string, got %T", field, v)
								}
							}
						}
					case "[]int":
						if arr, ok := value.([]interface{}); !ok {
							t.Errorf("Field %q should be array, got %T", field, value)
						} else {
							for _, v := range arr {
								if _, ok := v.(float64); !ok {
									t.Errorf("Field %q array element should be number, got %T", field, v)
								}
							}
						}
					case "float64":
						if _, ok := value.(float64); !ok {
							t.Errorf("Field %q should be number, got %T", field, value)
						}
					case "[]interface{}":
						if _, ok := value.([]interface{}); !ok {
							t.Errorf("Field %q should be array, got %T", field, value)
						}
					}
				}
			}
		})
	}
} 