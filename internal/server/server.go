package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mchl18/ssl-expiration-check-bot/internal/checker"
)

type Server struct {
	checker    *checker.CertificateChecker
	authToken  string
	homeDir    string
	startedAt  time.Time
	checkedAt  time.Time
	version    string
}

func New(checker *checker.CertificateChecker, authToken string, homeDir string) *Server {
	return &Server{
		checker:    checker,
		authToken:  authToken,
		homeDir:    homeDir,
		startedAt:  time.Now(),
		version:    "1.0.0",
	}
}

func (s *Server) Start(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.authMiddleware(s.handleHealth))
	mux.HandleFunc("/logs", s.authMiddleware(s.handleLogs))

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		if parts[1] != s.authToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":     "ok",
		"uptime":     time.Since(s.startedAt).String(),
		"domains":    s.checker.GetDomains(),
		"thresholds": s.checker.GetThresholds(),
		"started_at": s.startedAt.Format(time.RFC3339),
		"checked_at": s.checkedAt.Format(time.RFC3339),
		"version":    s.version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	lines := 100 // default number of lines
	if linesStr := r.URL.Query().Get("lines"); linesStr != "" {
		var err error
		lines, err = strconv.Atoi(linesStr)
		if err != nil {
			http.Error(w, "Invalid lines parameter", http.StatusBadRequest)
			return
		}
		if lines < 1 {
			http.Error(w, "Lines parameter must be positive", http.StatusBadRequest)
			return
		}
	}

	logPath := filepath.Join(s.homeDir, ".certchecker", "logs", "cert-checker.log")
	file, err := os.Open(logPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open log file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read all lines into a slice
	var logLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logLines = append(logLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to read log file: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the last N lines
	total := len(logLines)
	start := total - lines
	if start < 0 {
		start = 0
	}

	response := map[string]interface{}{
		"lines":     lines,
		"total":     total,
		"logs":      logLines[start:],
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) SetCheckedAt(t time.Time) {
	s.checkedAt = t
} 