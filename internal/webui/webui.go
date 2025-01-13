package webui

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mchl18/ssl-expiration-check-bot/internal/config"
	"github.com/mchl18/ssl-expiration-check-bot/internal/logger"
)

//go:embed templates/*
var templateFS embed.FS

type WebUI struct {
	homeDir    string
	logger     *logger.Logger
	templates  *template.Template
	authToken  string
	configured bool
	mu         sync.RWMutex
}

func New(homeDir string, logger *logger.Logger) (*WebUI, error) {
	// Parse base template first
	templates, err := template.ParseFS(templateFS, "templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %v", err)
	}

	// Parse other templates
	templates, err = templates.ParseFS(templateFS, "templates/login.html", "templates/configure.html", "templates/logs.html", "templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %v", err)
	}

	// Check if config exists and load auth token
	configPath := filepath.Join(homeDir, ".certchecker", "config", ".env")
	configured := false
	var authToken string
	if _, err := os.Stat(configPath); err == nil {
		configured = true
		// Load config to get auth token
		if cfg, err := config.Load(homeDir); err == nil {
			authToken = cfg.HTTPAuthToken
		}
	}

	return &WebUI{
		homeDir:    homeDir,
		logger:     logger,
		templates:  templates,
		configured: configured,
		authToken:  authToken,
	}, nil
}

func (w *WebUI) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", w.handleIndex)
	mux.HandleFunc("/configure", w.handleConfigure)
	mux.HandleFunc("/login", w.handleLogin)
	mux.HandleFunc("/logs", w.handleLogs)

	w.logger.Info("Starting web UI", map[string]interface{}{
		"address": "localhost:8081",
	})

	return http.ListenAndServe("localhost:8081", w.authMiddleware(mux))
}

func (w *WebUI) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w.mu.RLock()
		token := w.authToken
		configured := w.configured
		w.mu.RUnlock()

		// Skip auth for initial setup
		if !configured && r.URL.Path == "/" {
			next.ServeHTTP(rw, r)
			return
		}

		// Skip auth for login page
		if r.URL.Path == "/login" {
			next.ServeHTTP(rw, r)
			return
		}

		// Check session cookie
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != token {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func (w *WebUI) handleIndex(rw http.ResponseWriter, r *http.Request) {
	w.mu.RLock()
	configured := w.configured
	w.mu.RUnlock()

	if !configured {
		http.Redirect(rw, r, "/configure", http.StatusSeeOther)
		return
	}

	data := struct {
		Content    string
		Configured bool
	}{
		Content:    "index",
		Configured: configured,
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := w.templates.ExecuteTemplate(rw, "base.html", data); err != nil {
		http.Error(rw, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (w *WebUI) handleConfigure(rw http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(rw, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get form values
		domains := strings.TrimSpace(r.FormValue("domains"))
		thresholds := strings.TrimSpace(r.FormValue("thresholds"))
		webhookURL := strings.TrimSpace(r.FormValue("slack_webhook_url"))
		heartbeatStr := strings.TrimSpace(r.FormValue("heartbeat_hours"))
		intervalStr := strings.TrimSpace(r.FormValue("interval_hours"))
		httpEnabled := r.FormValue("http_enabled") == "on"
		httpPort := strings.TrimSpace(r.FormValue("http_port"))
		if httpPort == "" || httpPort == "0" {
			httpPort = "8080"
		}
		httpAuthToken := strings.TrimSpace(r.FormValue("http_auth_token"))

		// Validate required fields
		if domains == "" {
			http.Error(rw, "Domains are required", http.StatusBadRequest)
			return
		}
		if thresholds == "" {
			http.Error(rw, "Threshold days are required", http.StatusBadRequest)
			return
		}
		if webhookURL == "" {
			http.Error(rw, "Slack webhook URL is required", http.StatusBadRequest)
			return
		}

		// Clean up domains list
		domainsList := strings.Split(domains, ",")
		for i, d := range domainsList {
			domainsList[i] = strings.TrimSpace(d)
		}

		// Convert threshold days from string to []int
		thresholdStrs := strings.Split(thresholds, ",")
		var thresholdDays []int
		for _, s := range thresholdStrs {
			threshold, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				http.Error(rw, "Invalid threshold value", http.StatusBadRequest)
				return
			}
			thresholdDays = append(thresholdDays, threshold)
		}

		// Convert heartbeat and interval hours to integers
		var heartbeatHours, intervalHours int
		if heartbeatStr != "" {
			var err error
			heartbeatHours, err = strconv.Atoi(heartbeatStr)
			if err != nil {
				http.Error(rw, "Invalid heartbeat hours value", http.StatusBadRequest)
				return
			}
		}
		if intervalStr != "" {
			var err error
			intervalHours, err = strconv.Atoi(intervalStr)
			if err != nil {
				http.Error(rw, "Invalid interval hours value", http.StatusBadRequest)
				return
			}
		}

		cfg := &config.Config{
			Domains:         domainsList,
			ThresholdDays:  thresholdDays,
			SlackWebhookURL: webhookURL,
			HeartbeatHours: heartbeatHours,
			IntervalHours:  intervalHours,
			HTTPEnabled:    httpEnabled,
			HTTPAuthToken:  httpAuthToken,
		}

		if port, err := strconv.Atoi(httpPort); err == nil {
			cfg.HTTPPort = port
		}

		// Save configuration
		if err := w.saveConfig(cfg); err != nil {
			http.Error(rw, fmt.Sprintf("Failed to save configuration: %v", err), http.StatusInternalServerError)
			return
		}

		w.mu.Lock()
		w.configured = true
		w.authToken = httpAuthToken // Set the auth token here
		w.mu.Unlock()

		// Set session cookie for the user who just configured
		http.SetCookie(rw, &http.Cookie{
			Name:     "session",
			Value:    httpAuthToken,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := w.templates.ExecuteTemplate(rw, "base.html", struct{ Content string }{"configure"}); err != nil {
		http.Error(rw, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (w *WebUI) saveConfig(cfg *config.Config) error {
	configDir := filepath.Join(w.homeDir, ".certchecker", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Convert thresholds back to string for .env file
	var thresholdStrs []string
	for _, t := range cfg.ThresholdDays {
		thresholdStrs = append(thresholdStrs, strconv.Itoa(t))
	}

	// Build .env content
	content := fmt.Sprintf(`DOMAINS=%s
THRESHOLD_DAYS=%s
SLACK_WEBHOOK_URL=%s
HEARTBEAT_HOURS=%d
INTERVAL_HOURS=%d
HTTP_ENABLED=%t
HTTP_PORT=%d
HTTP_AUTH_TOKEN=%s
`, strings.Join(cfg.Domains, ","),
		strings.Join(thresholdStrs, ","),
		cfg.SlackWebhookURL,
		cfg.HeartbeatHours,
		cfg.IntervalHours,
		cfg.HTTPEnabled,
		cfg.HTTPPort,
		cfg.HTTPAuthToken)

	// Write to .env file
	envPath := filepath.Join(configDir, ".env")
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %v", err)
	}

	return nil
}

func (w *WebUI) handleLogin(rw http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		token := r.FormValue("token")
		w.mu.RLock()
		validToken := w.authToken
		w.mu.RUnlock()

		w.logger.Info("Login attempt", map[string]interface{}{
			"token":       token,
			"validToken": validToken,
			"configured": w.configured,
		})

		if token == validToken {
			http.SetCookie(rw, &http.Cookie{
				Name:     "session",
				Value:    token,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}

		http.Error(rw, "Invalid token", http.StatusUnauthorized)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := w.templates.ExecuteTemplate(rw, "base.html", struct{ Content string }{"login"}); err != nil {
		http.Error(rw, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (w *WebUI) handleLogs(rw http.ResponseWriter, r *http.Request) {
	logPath := filepath.Join(w.homeDir, ".certchecker", "logs", "cert-checker.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Failed to read logs: %v", err), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}

	data := struct {
		Content string
		Logs    []string
	}{
		Content: "logs",
		Logs:    lines,
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := w.templates.ExecuteTemplate(rw, "base.html", data); err != nil {
		http.Error(rw, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
} 